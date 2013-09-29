package main 

import (
    "net/http"
    "log"
    "strconv"
    "github.com/justinfx/go-socket.io/socketio"
    zmq "github.com/alecthomas/gozmq"
    "resistance/utils"
)

var allConnections map[*socketio.Conn]bool = make(map[*socketio.Conn]bool)

func handleMessage(zmqSocket *zmq.Socket, socket *socketio.Conn, msg socketio.Message) {
    zmqSocket.Send([]byte(msg.Data()), 0)
    utils.LogMessage("Sending to game backend", utils.RESISTANCE_LOG_PATH)
    utils.LogMessage(msg.Data(), utils.RESISTANCE_LOG_PATH)
    
    reply, _ := zmqSocket.Recv(0)
    utils.LogMessage("Reply received", utils.RESISTANCE_LOG_PATH)
    utils.LogMessage(string(reply), utils.RESISTANCE_LOG_PATH)
    
    socket.Send(reply)
    utils.LogMessage("Sent to frontend", utils.RESISTANCE_LOG_PATH)
}

func main() {
    // Setup ZMQ
    context, _ := zmq.NewContext()
    zmqSocket, _ := context.NewSocket(zmq.REQ)
    defer context.Close()
    defer zmqSocket.Close()
    
    zmqSocket.Connect("tcp://localhost:" + strconv.Itoa(utils.GAME_REP_REQ_PORT))
    utils.LogMessage("WSP connected to port " + strconv.Itoa(utils.GAME_REP_REQ_PORT), utils.RESISTANCE_LOG_PATH)
    
    // Setup Socket.IO
    config := socketio.DefaultConfig
    config.Origins = []string{"*:8080"}
    sio := socketio.NewSocketIO(&config)

    sio.OnConnect(func(c *socketio.Conn) {
        allConnections[c] = true
    })

    sio.OnDisconnect(func(c *socketio.Conn) {
        delete(allConnections, c)
    })

    sio.OnMessage(func(c *socketio.Conn, msg socketio.Message) {
        utils.LogMessage(c.String() + msg.Data(), utils.RESISTANCE_LOG_PATH)
        go handleMessage(zmqSocket, c, msg)
    })

    // Start server
    mux := sio.ServeMux()
    mux.Handle("/", http.FileServer(http.Dir("src/github.com/socket.io-client")))

    if err := http.ListenAndServe(":" + strconv.Itoa(utils.WSP_PORT), mux); err != nil {
        log.Fatal("ListenAndServe:", err)
    }
}
