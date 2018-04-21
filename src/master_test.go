package main

import (
    "fmt"
    "net"
    "testing"
)

func TestReceivesMessage(t *testing.T) {
    go main()
    conn, err := net.Dial("tcp", "localhost:3333")
    if err != nil {
        fmt.Printf("Error sending: %v", err)
        return
    }
    defer conn.Close()
}
