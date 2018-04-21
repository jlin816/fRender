package main

import (
    "fmt"
    "net"
    "testing"
    "time"
)

func TestReceiveMessage(t *testing.T) {
    go main()
    conn, err := net.Dial("tcp", "localhost:3333")
    if err != nil {
        fmt.Printf("Error sending: %v", err)
        return
    }
    defer conn.Close()
}

func TestMultipleReceiveMessage(t *testing.T) {
    go main()
    go connectClient()
    go connectClient()
    time.Sleep(5000 * time.Millisecond)
}

func connectClient() {
    conn, err := net.Dial("tcp", "localhost:3333")
    if err != nil {
        fmt.Printf("Error sending: %v", err)
        return
    }
    fmt.Println("before")
    time.Sleep(500 * time.Millisecond)
    conn.Close()

}
