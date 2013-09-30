package main 

import (
    "net/http"
    "log"
    "encoding/json"
    "github.com/justinfx/go-socket.io/socketio"
    zmq "github.com/alecthomas/gozmq"
    "resistance/utils"
)

type AcceptUser struct {
    AcceptUser bool
}

var allConnections map[*socketio.Conn]bool = make(map[*socketio.Conn]bool)

// subscribeConnection is meant to run in the background, subscribe to the
// given gameId, and forward messages to the frontend.
func subscribeConnection(socket *socketio.Conn, gameId int) {
    // TODO: implement subscribe
}

// isAcceptUser checks if the reply from the ZMQ socket indicated
// that the user was ok, so the proxy can start a listener on the SUBSCRIBE
// socket for that user.
func isAcceptUser(zmqReply []byte) (bool, int) {
    var acceptUser AcceptUser
    err := json.Unmarshal(zmqReply, &acceptUser)
    if err != nil {
        return false, 0
    }
    return acceptUser.AcceptUser, 0
}

// handleMessage handles a message from the frontend. Basically forwards it
// through ZMQ to the game backend, waits for a reply, then forwards the
// reply to the frontend. If this was a player connect message, and the
// user is accepted, we should start a listener to the SUBSCRIBE socket.  
func handleMessage(msg socketio.Message, socket *socketio.Conn, zmqSocket *zmq.Socket) {
    zmqSocket.Send([]byte(msg.Data()), 0)
    utils.LogMessage("Sending to game backend", utils.RESISTANCE_LOG_PATH)
    utils.LogMessage(msg.Data(), utils.RESISTANCE_LOG_PATH)
    
    reply, _ := zmqSocket.Recv(0)
    utils.LogMessage("Reply received", utils.RESISTANCE_LOG_PATH)
    utils.LogMessage(string(reply), utils.RESISTANCE_LOG_PATH)
    
    accept, gameId := isAcceptUser(reply)
    if (accept) {
        utils.LogMessage("User accepted!", utils.RESISTANCE_LOG_PATH)
        go subscribeConnection(socket, gameId) 
    }
    
    socket.Send(reply)
    utils.LogMessage("Sent to frontend", utils.RESISTANCE_LOG_PATH)
}

func main() {
    // Setup ZMQ
    context, _ := zmq.NewContext()
    zmqSocket, _ := context.NewSocket(zmq.REQ)
    defer context.Close()
    defer zmqSocket.Close()
    
    zmqSocket.Connect("tcp://localhost:" + utils.GAME_REP_REQ_PORT)
    utils.LogMessage("WSP connected to port " + utils.GAME_REP_REQ_PORT, utils.RESISTANCE_LOG_PATH)
    
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
        go handleMessage(msg, c, zmqSocket)
    })

    // Start server
    mux := sio.ServeMux()
    mux.Handle("/", http.FileServer(http.Dir("src/github.com/socket.io-client")))

    if err := http.ListenAndServe(":" + utils.WSP_PORT, mux); err != nil {
        log.Fatal("ListenAndServe:", err)
    }
}
