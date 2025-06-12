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

func TestMultiplePingCommand(t *testing.T) {
	conn, err := net.Dial("tcp", serverAddr)
	if err != nil {
		t.Fatalf("Failed to connect: %v", err)
	}
	defer conn.Close()
	for i := 0; i < 5; i++ {
		_, err = conn.Write([]byte(fmt.Sprintf("PING %d\n", i)))
		if err != nil {
			t.Fatalf("Failed to write: %v", err)
		}

		reader := bufio.NewReader(conn)
		response, err := reader.ReadString('\n')
		if err != nil {
			t.Fatalf("Failed to read: %v", err)
		}

		expected := fmt.Sprintf("PONG %d", i)
		if strings.TrimSpace(response) != expected {
			t.Errorf("Expected %q, got %q", expected, response)
		}
	}
}

func TestEchoCommand(t *testing.T) {
	conn, err := net.Dial("tcp", serverAddr)
	if err != nil {
		t.Fatalf("Failed to connect;; %v", err)
	}
	defer conn.Close()

	_, err = conn.Write([]byte("echo Abhishek test\n"))
	if err != nil {
		t.Fatalf("Failed to write %v", err)
	}

	reader := bufio.NewReader(conn)
	response, err := reader.ReadString('\n')
	if err != nil {
		t.Fatalf("Failed to read %v", err)
	}

	expected := "Abhishek test"
	if strings.TrimSpace(response) != expected {
		t.Errorf("Expected %q got %q", expected, response)
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

func TestSetCommand(t *testing.T) {
	tests := []struct {
		name        string
		input       string
		expected    string
		wantErr     bool
		expectedErr string
	}{
		{
			name:        "Unknown Command for case Insensitivity",
			input:       "set\n",
			expectedErr: "-ERR",
			wantErr:     true,
		},
		{
			name:        "Unknown Command for case Insensitivity",
			input:       "get\n",
			expectedErr: "-ERR",
			wantErr:     true,
		},
		{
			name:     "Succesful Set Command",
			input:    "SET abhishek\n",
			expected: "\n",
			wantErr:  false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			conn, err := net.Dial("tcp", serverAddr)
			if err != nil {
				t.Fatalf("Failed to connect: %v", err)
			}
			defer conn.Close()

			_, err = conn.Write([]byte(tc.input))
			if err != nil {
				t.Fatalf("Failed to write: %v", err)
			}

			reader := bufio.NewReader(conn)
			response, err := reader.ReadString('\n')
			if err != nil {
				t.Fatalf("Failed to read %v", err)
			}
			if tc.wantErr {
				if !strings.Contains(response, tc.expectedErr) {
					t.Fatalf("Expected error : %s Got : %s", tc.expectedErr, response)
				}
			} else {
				if tc.expected != response {
					t.Fatalf("Expected : %s: Got : %s", tc.expected, response)
				}
			}

		})
	}
}

func TestGetCommand(t *testing.T) {
	tests := []struct {
		name        string
		input       string
		expected    string
		wantErr     bool
		expectedErr string
	}{
		{
			name:        "Unknown Command for case Insensitivity",
			input:       "get\n",
			expectedErr: "-ERR",
			wantErr:     true,
		},
		{
			name:        "Unknown Command for case Insensitivity",
			input:       "GET aBhishek\n",
			expectedErr: "-ERR",
			wantErr:     true,
		},
		{
			name:     "Succesful Get Command",
			input:    "SET abhishek\n",
			expected: "\n",
			wantErr:  false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			conn, err := net.Dial("tcp", serverAddr)
			if err != nil {
				t.Fatalf("Failed to connect: %v", err)
			}
			defer conn.Close()
			// execute test command first TestGetCommand
			_, err = conn.Write([]byte("SET abhishek singh\n"))
			if err != nil {
				t.Fatalf("Failed to write: %v", err)
			}
			reader := bufio.NewReader(conn)
			response, err := reader.ReadString('\n')
			if err != nil {
				t.Fatalf("Failed to read %v", err)
			}
			_, err = conn.Write([]byte(tc.input))
			if err != nil {
				t.Fatalf("Failed to write: %v", err)
			}

			response, err = reader.ReadString('\n')
			if err != nil {
				t.Fatalf("Failed to read %v", err)
			}
			if tc.wantErr {
				if !strings.Contains(response, tc.expectedErr) {
					t.Fatalf("Expected error : %s Got : %s", tc.expectedErr, response)
				}
			} else {
				if tc.expected != response {
					t.Fatalf("Expected : %s: Got : %s", tc.expected, response)
				}
			}

		})
	}
}
