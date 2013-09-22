package main 

import (
    "net/http"
    "log"
    "strings"
    "encoding/json"
    "github.com/justinfx/go-socket.io/socketio"
    "resistance/utils"
    "resistance/users"
)

var allConnections []*socketio.Conn = make([]*socketio.Conn, 0)

func notifyNewPlayer(newUser string) {
    type Message struct {
        Message string
        UserName string
    }
    
    var msg = Message{Message:"newUser", UserName:newUser}
    
    for i := range allConnections {
        marshalledJSON, _ := json.Marshal(msg)
        allConnections[i].Send(marshalledJSON)
    }
}

func processMessage(msg socketio.Message) {
    type MessageHolder struct {
        Message string
        UserCookie string
    }
    var parsedMessage MessageHolder
    data := msg.Data()
    
    utils.LogMessage(data, utils.RESISTANCE_LOG_PATH)
    err := json.Unmarshal([]byte(data), &parsedMessage)
    if err != nil {
        utils.LogMessage("Error parsing message: " + data, utils.RESISTANCE_LOG_PATH)
        return
    }
    
    cookies := make([]*http.Cookie, 1)
    parsedCookie := strings.Split(parsedMessage.UserCookie, "=")
    cookies[0] = &http.Cookie{Name:parsedCookie[0], Value:parsedCookie[1]}
    user, success := users.ValidateUserCookie(cookies)
    if !success {
        utils.LogMessage("Something went wrong when validating the user", utils.RESISTANCE_LOG_PATH)
    }

    switch {
        case parsedMessage.Message == "firstConnection": notifyNewPlayer(user.Username)
    }
}

func main() {
    config := socketio.DefaultConfig
    config.Origins = []string{"*:8080"}
    sio := socketio.NewSocketIO(&config)

    sio.OnConnect(func(c *socketio.Conn) {
        allConnections = append(allConnections, c)
    })

    sio.OnDisconnect(func(c *socketio.Conn) {
    })

    sio.OnMessage(func(c *socketio.Conn, msg socketio.Message) {
        utils.LogMessage(c.String() + msg.Data(), utils.RESISTANCE_LOG_PATH)
        processMessage(msg)
    })

    mux := sio.ServeMux()
    mux.Handle("/", http.FileServer(http.Dir("src/github.com/socket.io-client")))

    if err := http.ListenAndServe(":8081", mux); err != nil {
        log.Fatal("ListenAndServe:", err)
    }
}

