package main 

import (
    "net/http"
    "log"
    "strconv"
    "encoding/json"
    "github.com/justinfx/go-socket.io/socketio"
    zmq "github.com/alecthomas/gozmq"
    "resistance/utils"
)

type AcceptUser struct {
    AcceptUser bool
    GameId int
    UserId int
}

var allConnections map[*socketio.Conn]int = make(map[*socketio.Conn]int)

// subscribeConnection is meant to run in the background, subscribe to the
// given gameId, and forward messages to the frontend.
func subscribeConnection(context *zmq.Context, socket *socketio.Conn, gameId int) {
    subSocket, _ := context.NewSocket(zmq.SUB)
    defer subSocket.Close()
    
    subSocket.Connect("tcp://localhost:" + utils.GAME_PUB_SUB_PORT)
    subSocket.SetSockOptString(zmq.SUBSCRIBE, strconv.Itoa(gameId))
    utils.LogMessage("SUBCRIBER connected to port " + utils.GAME_REP_REQ_PORT + " with filter " + strconv.Itoa(gameId), utils.RESISTANCE_LOG_PATH)
    
    for allConnections[socket] != 0 {
        game, _ := subSocket.Recv(0)
        rest, _ := subSocket.Recv(0)
        utils.LogMessage("got message on sub socket for game " + string(game) + ":" + string(rest), utils.RESISTANCE_LOG_PATH)
        socket.Send(rest)
    }
    
    utils.LogMessage("Lost SUBSCRIBER connection with filter " + strconv.Itoa(gameId), utils.RESISTANCE_LOG_PATH)
}

// isAcceptUser checks if the reply from the ZMQ socket indicated
// that the user was ok, so the proxy can start a listener on the SUBSCRIBE
// socket for that user. Returns (acceptUser, gameId, userId)
func isAcceptUser(zmqReply []byte) (bool, int, int) {
    var acceptUser AcceptUser
    err := json.Unmarshal(zmqReply, &acceptUser)
    if err != nil {
        return false, 0, 0
    }
    return acceptUser.AcceptUser, acceptUser.GameId, acceptUser.UserId
}

// handleMessage handles a message from the frontend. Basically forwards it
// through ZMQ to the game backend, waits for a reply, then forwards the
// reply to the frontend. If this was a player connect message, and the
// user is accepted, we should start a listener to the SUBSCRIBE socket.  
func handleMessage(msg socketio.Message, socket *socketio.Conn, context *zmq.Context) {
    zmqSocket, _ := context.NewSocket(zmq.REQ)
    defer zmqSocket.Close()
    
    zmqSocket.Connect("tcp://localhost:" + utils.GAME_REP_REQ_PORT)
    utils.LogMessage("WSP connected to port " + utils.GAME_REP_REQ_PORT, utils.RESISTANCE_LOG_PATH)
    
    zmqSocket.Send([]byte(msg.Data()), 0)
    utils.LogMessage("Sending to game backend", utils.RESISTANCE_LOG_PATH)
    utils.LogMessage(msg.Data(), utils.RESISTANCE_LOG_PATH)
    
    reply, _ := zmqSocket.Recv(0)
    utils.LogMessage("Reply received", utils.RESISTANCE_LOG_PATH)
    utils.LogMessage(string(reply), utils.RESISTANCE_LOG_PATH)
    
    accept, gameId, userId := isAcceptUser(reply)
    if (accept) {
        utils.LogMessage("User accepted: " + strconv.Itoa(userId), utils.RESISTANCE_LOG_PATH)
        allConnections[socket] = userId
        go subscribeConnection(context, socket, gameId) 
    }
    
    socket.Send(reply)
    utils.LogMessage("Sent to frontend", utils.RESISTANCE_LOG_PATH)
}

func main() {
    // Setup ZMQ
    context, _ := zmq.NewContext()
    defer context.Close()
    
    // Setup Socket.IO
    config := socketio.DefaultConfig
    config.Origins = []string {"*:80"}
    sio := socketio.NewSocketIO(&config)

    sio.OnConnect(func(c *socketio.Conn) {
    })

    sio.OnDisconnect(func(c *socketio.Conn) {
        delete(allConnections, c)
        // TODO: make sure go routine will stop on disconnect
        // TODO: send message to backend on disconnect
    })

    sio.OnMessage(func(c *socketio.Conn, msg socketio.Message) {
        utils.LogMessage(c.String() + msg.Data(), utils.RESISTANCE_LOG_PATH)
        go handleMessage(msg, c, context)
    })

    // Start server
    mux := sio.ServeMux()
    mux.Handle("/", http.FileServer(http.Dir("src/github.com/socket.io-client")))

    if err := http.ListenAndServe(":" + utils.WSP_PORT, mux); err != nil {
        log.Fatal("ListenAndServe:", err)
    }
}
