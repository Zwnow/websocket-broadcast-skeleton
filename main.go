package main

import (
    "fmt"
    "net/http"
    "sync"
    "time"

    "github.com/gorilla/websocket"
)
var (
    upgrader = websocket.Upgrader {
        CheckOrigin: func (r *http.Request) bool {
            return true
        },
    }
    connections = make(map[*websocket.Conn]struct{})
    connectionsMutex sync.Mutex
)

func handleWebSocket(w http.ResponseWriter, r *http.Request) {
    conn, err := upgrader.Upgrade(w, r, nil)
    if err != nil {
        fmt.Println(err)
        return
    }
    defer func() {
        conn.Close()
        connectionsMutex.Lock()
        delete(connections, conn)
        connectionsMutex.Unlock()
        fmt.Println("Client disconnected")
    }()

    connectionsMutex.Lock()
    connections[conn] = struct{}{}
    connectionsMutex.Unlock()

    fmt.Println("Client connected")

    for {
        messageType, p, err := conn.ReadMessage()
        if err != nil {
            fmt.Println(err)
            return
        }

        fmt.Printf("Received: %s\n", p)

        err = conn.WriteMessage(messageType, p)
        if err != nil {
            fmt.Println(err)
            return
        }
    }
}

func broadcast(message string) {
    connectionsMutex.Lock()
    defer connectionsMutex.Unlock()

    for conn := range connections {
        err := conn.WriteMessage(websocket.TextMessage, []byte(message))
        if err != nil {
            fmt.Println("Error broadcasting:", err)
        }
    }
}

func main() {
    http.HandleFunc("/ws", handleWebSocket)

    go func() {
        for {
            broadcast("Backend state changed: changeMessage")

            <-time.After(5 * time.Second)
        }
    }()

    fmt.Println("WebSocket server is running.")

    err := http.ListenAndServe(":8080", nil)
    if err != nil {
        fmt.Println(err)
    }
}
