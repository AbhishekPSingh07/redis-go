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
	setCommandArgs := ""
	setCommandKey := ""
	for scanner.Scan() {
		input := strings.TrimSpace(scanner.Text())
		fields := strings.Fields(input)
		if len(fields) == 0 {
			continue
		}

		cmd := fields[0]
		args := strings.Join(fields[1:], " ")

		switch cmd {
		case "PING":
			if args == "" {
				connection.Write([]byte("PONG\n"))
			} else {
				connection.Write([]byte(fmt.Sprintf("PONG %s\n", args)))
			}
		case "echo":
			connection.Write([]byte(fmt.Sprintln(args)))
		case "SET":
			setCommandFields := strings.Fields(strings.TrimSpace(args))
			setCommandKey = setCommandFields[0]
			setCommandArgs = strings.Join(setCommandFields[1:], " ")
			connection.Write([]byte("\n"))
		case "GET":
			if strings.TrimSpace(args) == setCommandKey {
				connection.Write([]byte(fmt.Sprintln(setCommandArgs)))
			} else {
				connection.Write([]byte("-ERR no such key set\n"))
			}
		default:
			connection.Write([]byte("-ERR unknown command\n"))
		}
	}
}
