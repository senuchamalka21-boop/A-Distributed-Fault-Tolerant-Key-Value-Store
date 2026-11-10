package main

import (
	"bufio"
	"fmt"
	"net"
	"os"
	"strings"
)

// We want 3 copies of every key in the cluster
const N = 3

func main() {
	if len(os.Args) < 2 {
		fmt.Println("usage: prog <myPort> <peer1> <peer2> ...")
		os.Exit(1)
	}
	myAddr := "localhost:" + os.Args[1]
	peers := os.Args[2:]

	ring := NewRing()
	ring.AddNode(myAddr)
	for _, p := range peers {
		ring.AddNode(p)
	}

	store := NewStore()

	ln, err := net.Listen("tcp", ":"+os.Args[1])
	if err != nil {
		panic(err)
	}
	fmt.Println("listening on", myAddr, "| ring peers:", peers)

	for {
		conn, err := ln.Accept()
		if err != nil {
			continue
		}
		go handleConn(conn, store, ring, myAddr)
	}
}

func handleConn(conn net.Conn, store *Store, ring *Ring, myAddr string) {
	defer conn.Close()
	scanner := bufio.NewScanner(conn)
	writer := bufio.NewWriter(conn)

	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		parts := strings.Fields(line)
		if len(parts) == 0 {
			continue
		}

		cmd := strings.ToUpper(parts[0])
		if len(parts) < 2 {
			writer.WriteString("ERR bad args\n")
			writer.Flush()
			continue
		}
		key := parts[1]

		// REPLICATION: Get the top 3 nodes for this key
		replicas := ring.GetN(key, N)

		// Check if we are one of the replicas
		isReplica := false
		for _, r := range replicas {
			if r == myAddr {
				isReplica = true
				break
			}
		}

		// If I am NOT a replica, forward the request to the primary owner (replicas[0])
		if !isReplica {
			resp, err := forward(replicas[0], line)
			if err != nil {
				writer.WriteString("ERR forward failed\n")
			} else {
				writer.WriteString(resp)
			}
			writer.Flush()
			continue
		}

		// I AM a replica, so handle the command
		switch cmd {
		case "PUT":
			if len(parts) < 3 {
				writer.WriteString("ERR bad args\n")
			} else {
				// 1. Save locally
				store.Put(key, parts[2])
				
				// 2. Send to the OTHER replicas in the background
				for _, r := range replicas {
					if r != myAddr {
						go forward(r, line) // notice: we don't care about the reply
					}
				}
				writer.WriteString("OK\n")
			}
		case "GET":
			// 1. Check our local store first
			if v, ok := store.Get(key); ok {
				writer.WriteString("VALUE " + v + "\n")
			} else {
				// 2. If we don't have it, ask the other replicas
				found := false
				for _, r := range replicas {
					if r != myAddr {
						resp, err := forward(r, line)
						if err == nil && strings.HasPrefix(resp, "VALUE") {
							writer.WriteString(resp)
							found = true
							break
						}
					}
				}
				if !found {
					writer.WriteString("NOTFOUND\n")
				}
			}
		case "DEL":
			store.Delete(key)
			for _, r := range replicas {
				if r != myAddr {
					go forward(r, line)
				}
			}
			writer.WriteString("OK\n")
		default:
			writer.WriteString("ERR unknown command\n")
		}
		writer.Flush()
	}
}

// forward sends a single command to another node and returns its reply
func forward(addr, line string) (string, error) {
	c, err := net.Dial("tcp", addr)
	if err != nil {
		return "", err
	}
	defer c.Close()

	c.Write([]byte(line + "\n"))
	reader := bufio.NewReader(c)
	resp, err := reader.ReadString('\n')
	return resp, err
}