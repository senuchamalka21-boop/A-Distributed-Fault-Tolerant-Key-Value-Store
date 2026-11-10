package main

import (
	"bufio"
	"fmt"
	"net"
	"os"
)

func main() {
	if len(os.Args) < 3 {
		fmt.Println("Usage: go run client/client.go <port> <command>")
		os.Exit(1)
	}
	port := os.Args[1]
	command := os.Args[2]

	conn, err := net.Dial("tcp", "localhost:"+port)
	if err != nil {
		fmt.Println("Error connecting:", err)
		return
	}
	defer conn.Close()

	// Send command
	fmt.Fprintf(conn, "%s\n", command)

	// Read response
	reader := bufio.NewReader(conn)
	resp, _ := reader.ReadString('\n')
	fmt.Print("Server says: ", resp)
}
