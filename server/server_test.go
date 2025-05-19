package server

import (
	"bufio"
	"context"
	"fmt"
	"net"
	"os"
	"strings"
	"testing"
	"time"
)

var serverAddr string
var stopServer context.CancelFunc

func TestMain(m *testing.M) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	listener, err := StartServer(ctx, "127.0.0.1:0") // use random port
	if err != nil {
		fmt.Printf("Failed to start server: %v\n", err)
		os.Exit(1)
	}
	serverAddr = listener.Addr().String()
	stopServer = cancel
	fmt.Printf("StartServer")

	code := m.Run() // Run tests

	stopServer()                       // Stop server
	time.Sleep(100 * time.Millisecond) // Let server shut down

	os.Exit(code)
}

func TestPingCommand(t *testing.T) {
	conn, err := net.Dial("tcp", serverAddr)
	if err != nil {
		t.Fatalf("Failed to connect: %v", err)
	}
	defer conn.Close()

	_, err = conn.Write([]byte("PING\n"))
	if err != nil {
		t.Fatalf("Failed to write: %v", err)
	}

	reader := bufio.NewReader(conn)
	response, err := reader.ReadString('\n')
	if err != nil {
		t.Fatalf("Failed to read: %v", err)
	}

	expected := "PONG"
	if strings.TrimSpace(response) != expected {
		t.Errorf("Expected %q, got %q", expected, response)
	}
}

func TestUnknownCommand(t *testing.T) {
	conn, err := net.Dial("tcp", serverAddr)
	if err != nil {
		t.Fatalf("Failed to connect: %v", err)
	}
	defer conn.Close()

	_, err = conn.Write([]byte("FOOBAR1\n"))
	if err != nil {
		t.Fatalf("Failed to write: %v", err)
	}

	reader := bufio.NewReader(conn)
	response, err := reader.ReadString('\n')
	if err != nil {
		t.Fatalf("Failed to read: %v", err)
	}

	if !strings.HasPrefix(response, "-ERR") {
		t.Errorf("Expected error response, got %q", response)
	}
}

func TestConcurrentPings(t *testing.T) {
	const numClients = 20
	errs := make(chan error, numClients)

	for i := 0; i < numClients; i++ {
		go func(id int) {
			conn, err := net.Dial("tcp", serverAddr)
			if err != nil {
				errs <- err
				return
			}
			defer conn.Close()

			for j := 0; j < 5; j++ {
				_, err = conn.Write([]byte("PING\n"))
				if err != nil {
					errs <- err
					return
				}

				reader := bufio.NewReader(conn)
				response, err := reader.ReadString('\n')
				if err != nil {
					errs <- err
					return
				}
				if strings.TrimSpace(response) != "PONG" {
					errs <- fmt.Errorf("client %d expected %q got %q", id, "PONG", response)
					return
				}
			}

			errs <- nil
		}(i)
	}
	for i := 0; i < numClients; i++ {
		if err := <-errs; err != nil {
			t.Errorf("Error: %v", err)
		}
	}
}
