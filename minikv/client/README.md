# MiniKV: A Distributed Key-Value Store

A fault-tolerant, distributed in-memory key-value store built from scratch in Go.

## Architecture
- **Consistent Hashing**: Uses a hash ring with 50 virtual nodes per physical server to evenly distribute keys and minimize data movement when nodes are added/removed.
- **Request Forwarding**: Any node can receive a request. If it is not the owner of the key, it transparently forwards the request to the correct node using TCP.
- **Replication (N=3)**: Every key is written to the primary owner and its next 2 clockwise neighbors on the hash ring. Reads fall back to replicas if the primary node is down.
- **Concurrency**: Handles multiple simultaneous client connections using Go goroutines.

## How to Run
(Add the instructions here: open 3 terminals and run `go run . 8001 ...`, etc.)

## Failover Demo
(Describe the test you just did where you killed Server 1 and retrieved the data from Server 2).