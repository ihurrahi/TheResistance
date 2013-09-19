package main 

import (
    "net/http"
    "log"
    "fmt"
    "github.com/justinfx/go-socket.io/socketio"
)

func main() {
    config := socketio.DefaultConfig
    config.Origins = []string{"*:8080"}
    sio := socketio.NewSocketIO(&config)

    sio.OnConnect(func(c *socketio.Conn) {
        fmt.Println("connected: " + c.String())
    })

    sio.OnDisconnect(func(c *socketio.Conn) {
        fmt.Println("disconnected: " + c.String())
    })

    sio.OnMessage(func(c *socketio.Conn, msg socketio.Message) {
        fmt.Println(c.String() + msg.Data())
    })

    mux := sio.ServeMux()
    mux.Handle("/", http.FileServer(http.Dir("src/github.com/socket.io-client")))

    if err := http.ListenAndServe(":8081", mux); err != nil {
        log.Fatal("ListenAndServe:", err)
    }
}

