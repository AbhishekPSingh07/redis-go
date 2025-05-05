package main

import (
	"context"
	"fmt"
	"net"
	"os"
)

// Ensures gofmt doesn't remove the "net" and "os" imports in stage 1 (feel free to remove this!)
var _ = net.Listen
var _ = os.Exit

func StartServer(ctx context.Context, addr string) (net.Listener, error) {

	l, err := net.Listen("tcp", addr)
	if err != nil {
		fmt.Printf("Failed to bind to addr err :%v :%v", addr, err)
		return nil, err
	}

	for {
		defer l.Close()
		conn, err := l.Accept()
		if err != nil {
			select {
			case <-ctx.Done():
				fmt.Println("Server Shutting Down")
				return nil, err
			default:
				fmt.Printf("recieved error: %v", err)
				continue
			}
		}
		go handleConnection(conn)
	}

}

func handleConnection(connection net.Conn) {
	defer connection.Close()

	buf := make([]byte, 1024)

	for {
		n, err := connection.Read(buf)
		if err != nil {
			fmt.Println("Error reading or connection closed:", err.Error())
			return
		}

		input := string(buf[:n])
		fmt.Println("Received:", input)

		if input == "PING\n" {
			connection.Write([]byte("+PONG\r\n"))
		} else {
			connection.Write([]byte("-ERR unknown command\r\n"))
		}
	}
}
