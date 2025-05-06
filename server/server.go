package server

import (
	"bufio"
	"context"
	"fmt"
	"net"
	"os"
	"strings"
)

// Ensures gofmt doesn't remove the "net" and "os" imports in stage 1 (feel free to remove this!)
var _ = net.Listen
var _ = os.Exit

func StartServer(ctx context.Context, addr string) (net.Listener, error) {
	l, err := net.Listen("tcp", addr)
	if err != nil {
		fmt.Printf("Failed to bind to addr %v: %v\n", addr, err)
		return nil, err
	}

	fmt.Printf("Server started on %s\n", l.Addr().String())

	go func() {
		defer l.Close()
		for {
			conn, err := l.Accept()
			if err != nil {
				select {
				case <-ctx.Done():
					fmt.Println("Server shutting down")
					return
				default:
					fmt.Printf("Received error: %v\n", err)
					continue
				}
			}
			go handleConnection(conn)
		}
	}()

	return l, nil
}

func handleConnection(connection net.Conn) {
	defer connection.Close()

	scanner := bufio.NewScanner(connection)
	for scanner.Scan() {
		input := scanner.Text()
		fmt.Println("Received:", input)

		if strings.HasPrefix(input, "PING") {
			rest := strings.TrimSpace(strings.TrimPrefix(input, "PING"))

			if rest == "" {
				connection.Write([]byte("PONG\n"))
			} else {
				connection.Write([]byte("PONG rest\n"))
			}
		} else {
			connection.Write([]byte("-ERR unknown command\r\n"))
		}
	}
}
