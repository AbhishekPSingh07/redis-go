package main

import (
	"fmt"
	"net"
	"os"
)

// Ensures gofmt doesn't remove the "net" and "os" imports in stage 1 (feel free to remove this!)
var _ = net.Listen
var _ = os.Exit

func main() {
	// You can use print statements as follows for debugging, they'll be visible when running tests.
	fmt.Println("Logs from your program will appear here!")

	// Uncomment this block to pass the first stage
	//
	l, err := net.Listen("tcp", "0.0.0.0:6379")
	if err != nil {
		fmt.Println("Failed to bind to port 6379")
		os.Exit(1)
	}
	defer l.Close()

	for {
		conn, err := l.Accept()
		if err != nil {
			fmt.Println("Error accepting connection: ", err.Error())
			os.Exit(1)
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

		if input == "PING\r\n" {
			connection.Write([]byte("+PONG\r\n"))
		} else {
			connection.Write([]byte("-ERR unknown command\r\n"))
		}
	}
}
