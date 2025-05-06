// cmd/myapp/main.go
package main

import (
    "context"
    "fmt"
    "os"
    "os/signal"
    "syscall"
    "time"

	"github.com/AbhishekPSingh07/redis-go/server"

)

func main() {
    ctx, cancel := context.WithCancel(context.Background())
    defer cancel()

    _, err := server.StartServer(ctx, "0.0.0.0:6379")
    if err != nil {
        fmt.Println("Failed to start server:", err)
        os.Exit(1)
    }

    // Handle SIGINT / SIGTERM gracefully
    sigs := make(chan os.Signal, 1)
    signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)

    fmt.Println("Server is running. Press Ctrl+C to stop.")
    <-sigs // wait for termination signal

    fmt.Println("Shutdown signal received")
    cancel()
    time.Sleep(200 * time.Millisecond) // wait for server to shut down cleanly
}
