# doogle 

_Web search of the people, by the people, for the people with Go._

[![CircleCI](https://circleci.com/gh/mathetake/doogle.svg?style=shield)](https://circleci.com/gh/mathetake/doogle)
[![MIT License](http://img.shields.io/badge/license-MIT-blue.svg?style=flat)](LICENSE)


doogle is a Proof of Concept software of __decentralized search engine__ based on gRPC written in Go.

This is just a PoC, and so I am not serious about making doogle secure, scalable, usable in production, etc.


## algorithms behind doogle

### 1. Distributed Hash Table based on S/Kademlia
Baumgart, Ingmar, and Sebastian Mies. "S/kademlia: A practicable approach towards secure key-based routing." Parallel and Distributed Systems, 2007 International Conference on. IEEE, 2007.

### 2. local estimation of PageRank with `WorldNode`
Parreira, Josiane Xavier, et al. "Efficient and decentralized pagerank approximation in a peer-to-peer web search network." Proceedings of the 32nd international conference on Very large data bases. VLDB Endowment, 2006.


## development


### build and test

```bash
❯ go get -u -d github.com/mathetake/doogle
❯ cd $GOPATH/src/github.com/mathetake/doogle
❯ go build .
❯ go test -v -race ./...
```


### start node

```
❯ ./doogle --help
Usage of ./doogle:
  -c int
        crawler's channel capacity
  -d int
        difficulty for cryptographic puzzle
  -p string
        port for node
  -w int
        number of crawler's worker
        
❯ ./doogle -c 4 -d 1 -p :12312 -w 4
INFO[0000] node created: doogleAddress=ad97676370397f6eb23dc165a34bf74a9c11d243 
INFO[0000] crawler is ready                             
INFO[0000] node listen on port: :12312, num of crawler's worker: 0  
INFO[0000] difficulty: 1, crawler's queue capacity: 4
```

You can connect to the node with, for example, [grpcc](https://github.com/njpatel/grpcc):

```
❯ grpcc --proto grpc/doogle.proto --address localhost:12312 --insecure

Connecting to doogle.Doogle on localhost:12312. Available globals:

  client - the client connection to Doogle
    storeItem (StoreItemRequest, callback) returns Empty
    findIndex (FindIndexRequest, callback) returns FindIndexReply
    findNode (FindNodeRequest, callback) returns NodeInfos
    pingWithCertificate (NodeCertificate, callback) returns NodeCertificate
    ping (StringMessage, callback) returns StringMessage
    pingTo (NodeInfo, callback) returns StringMessage
    getIndex (StringMessage, callback) returns GetIndexReply
    postUrl (StringMessage, callback) returns StringMessage

  printReply - function to easily print a unary call reply (alias: pr)
  streamReply - function to easily print stream call replies (alias: sr)
  createMetadata - convert JS objects into grpc metadata instances (alias: cm)
  printMetadata - function to easily print a unary call's metadata (alias: pm)
```

and then call `Ping`:

```
Doogle@localhost:12312> client.ping({ message: 'ping' }, printReply)
EventEmitter {}
Doogle@localhost:12312>
{
  "message": "pong"
}
```


### start node using docker

```bash
docker run -p 12312:8080 -it doogle bash -c "./doogle -d 2 -p :8080"
```


### modify interface

install protc,  then run:

```bash
protoc -I grpc/ grpc/doogle.proto --go_out=plugins=grpc:grpc
```

## References

see my survey: [Towards decentralized information retrieval: research papers](https://github.com/mathetake/notes/issues/1)


Also there are two articles in __Japanese__:
- [Distributed Hash Table and p2p Search Engine](https://scrapbox.io/layerx/Distributed_Hash_Table_and_p2p_Search_Engine)
- [Peer-to-Peer Information Retrieval: An Overview](https://scrapbox.io/layerx/%5BWIP%5DPeer-to-Peer_Information_Retrieval:_An_Overview)

## Author

[@mathetake](https://twitter.com/mathetake)

## LICENSE

MIT
