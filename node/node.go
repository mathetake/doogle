package node

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"time"

	pb "github.com/mathetake/doogle/grpc"
	"github.com/pkg/errors"
	"golang.org/x/crypto/ed25519"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/peer"
	"google.golang.org/grpc/status"
)

// constant network parameters
const (
	alpha      = 3
	bucketSize = 20
)

type item struct {
	dAddr doogleAddress

	url         string
	title       string
	description string

	// outgoing hyperlinks
	edges []doogleAddress

	// localRank represents computed locally PageRank
	localRank float64
}

type Node struct {
	// should be 160 bits
	dAddr doogleAddress

	// table for routing
	// keys correspond to `distance bits`
	// type: map{int -> *routingBucket}
	routingTable map[int]*routingBucket

	// distributed hash table points to addresses of items
	// type: map{doogleAddressStr -> *dhtValue}
	dht sync.Map

	// map of address to item's pointer
	// type: map{doogleAddressStr -> *item}
	items sync.Map

	// for certification
	publicKey  []byte
	secretKey  []byte
	nonce      []byte
	difficulty int
}

// nodeInfo contains the information for connecting nodes
type nodeInfo struct {
	dAddr      doogleAddress
	host       string
	port       string
	accessedAt int64
}

type routingBucket struct {
	bucket []*nodeInfo
	mux    sync.Mutex
}

// pop item on `idx` and then append `ni`
func (rb *routingBucket) popAndAppend(idx int, ni *nodeInfo) {
	prev := rb.bucket
	l := len(prev)
	rb.bucket = make([]*nodeInfo, l)
	for i := 0; i < l; i++ {
		if i == l-1 {
			rb.bucket[i] = ni
		} else if i < idx {
			rb.bucket[i] = prev[i]
		} else {
			rb.bucket[i] = prev[i+1]
		}
	}
}

type dhtValue struct {
	addresses []doogleAddressStr
	mux       sync.Mutex
}

func (n *Node) isValidSender(ctx context.Context, rawAddr, pk, nonce []byte, difficulty int) bool {
	if len(rawAddr) < addressLength {
		return false
	}

	var da doogleAddress
	copy(da[:], rawAddr[:addressLength])

	if pr, ok := peer.FromContext(ctx); ok {
		addr := strings.Split(pr.Addr.String(), ":")

		// if NodeCertificate is valid, update routing table with nodeInfo
		if verifyAddress(da, addr[0], addr[1], pk, nonce, difficulty) {
			ni := nodeInfo{
				dAddr:      da,
				host:       addr[0],
				port:       addr[1],
				accessedAt: time.Now().UTC().Unix(),
			}

			// update the routing table
			n.updateRoutingTable(&ni)
			return true
		}
	}
	return false
}

// update routingTable using a given nodeInfo
func (n *Node) updateRoutingTable(info *nodeInfo) {
	idx := getMostSignificantBit(n.dAddr.xor(info.dAddr))

	rb, ok := n.routingTable[idx]
	if !ok || rb == nil {
		panic(fmt.Sprintf("the routing table on %d not exist", idx))
	}

	// lock the bucket
	rb.mux.Lock()
	defer rb.mux.Unlock() // unlock the bucket
	for i, n := range rb.bucket {
		if n.dAddr == info.dAddr {
			// Update accessedAt on target node.
			n.accessedAt = time.Now().UTC().Unix()

			// move the target to tail of the bucket
			rb.popAndAppend(i, n)
			return
		}
	}

	ni := &nodeInfo{
		host:       info.host,
		port:       info.port,
		dAddr:      info.dAddr,
		accessedAt: time.Now().UTC().Unix(),
	}

	if len(rb.bucket) < bucketSize {
		rb.bucket = append(rb.bucket, ni)
	} else {
		rb.popAndAppend(0, ni)
	}
}

func (n *Node) StoreItem(ctx context.Context, in *pb.StoreItemRequest) (*pb.Empty, error) {
	return nil, nil
}
func (n *Node) FindIndex(ctx context.Context, in *pb.FindIndexRequest) (*pb.FindIndexReply, error) {
	return nil, nil
}

func (n *Node) FindNode(ctx context.Context, in *pb.FindNodeRequest) (*pb.FindeNodeReply, error) {
	return nil, nil
}

func (n *Node) GetIndex(ctx context.Context, in *pb.StringMessage) (*pb.GetIndexReply, error) {
	return nil, nil
}

func (n *Node) PostUrl(ctx context.Context, in *pb.StringMessage) (*pb.StringMessage, error) {
	return nil, nil
}

func (n *Node) Ping(ctx context.Context, in *pb.NodeCertificate) (*pb.StringMessage, error) {
	// TODO: logging the result of validation
	n.isValidSender(ctx, in.DoogleAddress, in.PublicKey, in.Nonce, int(in.Difficulty))
	return &pb.StringMessage{Message: "Pong"}, nil
}

func (n *Node) PingTo(ctx context.Context, in *pb.NodeInfo) (*pb.StringMessage, error) {
	conn, err := grpc.Dial(in.Host+":"+in.Port, grpc.WithInsecure())
	if err != nil {
		return nil, status.Errorf(codes.Internal, "did not connect: %v", err)
	}
	defer conn.Close()

	c := pb.NewDoogleClient(conn)
	r, err := c.Ping(ctx, &pb.NodeCertificate{
		DoogleAddress: n.dAddr[:addressLength],
		PublicKey:     n.publicKey,
		Nonce:         n.nonce,
		Difficulty:    int32(n.difficulty),
	})
	if err != nil {
		return nil, status.Errorf(codes.Internal, "c.Ping failed: %v", err)
	}

	return &pb.StringMessage{Message: r.Message}, nil
}

func NewNode(difficulty int, host, port string) (*Node, error) {
	pk, sk, err := ed25519.GenerateKey(nil)
	if err != nil {
		return nil, errors.Wrap(err, "failed to generate encryption keys")
	}

	// initialize routing table
	rt := map[int]*routingBucket{}
	for i := 0; i < 160; i++ {
		b := make([]*nodeInfo, 0, bucketSize)
		rt[i] = &routingBucket{bucket: b, mux: sync.Mutex{}}
	}

	// set node parameters
	node := Node{
		publicKey:    pk,
		secretKey:    sk,
		difficulty:   difficulty,
		routingTable: rt,
	}

	// solve network puzzle
	node.dAddr, node.nonce, err = newNodeAddress(host, port, node.publicKey, node.difficulty)
	if err != nil {
		return nil, errors.Wrap(err, "failed to generate address")
	}

	// TODO: start scheduled crawler
	// TODO: start PageRank computing scheduler
	return &node, nil
}

var _ pb.DoogleServer = &Node{}
