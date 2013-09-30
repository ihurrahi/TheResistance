package main 

import (
    "net/http"
    "strings"
    "encoding/json"
    zmq "github.com/alecthomas/gozmq"
    "resistance/utils"
    "resistance/users"
)

type MessageHolder struct {
    Message string
    UserCookie string
}

// handlerPlayerDisconnect handles the message that is sent when
// a player disconnects from the web socket proxy.
func handlePlayerDisconnect(message *MessageHolder) string {
    // TODO: implement
    return ""
}

// handlePlayerConnect handles the message that is sent when a player
// first connects by loading the game page.
func handlePlayerConnect(message *MessageHolder, user *users.User) string {
    // TODO: implement sending a message to SUB zmqSocket
    utils.LogMessage(user.Username + " has sent a message!", utils.RESISTANCE_LOG_PATH)
    newMessage := "{\"message\":\"players\",\"acceptUser\":true,\"players\":[\"" + user.Username + "\"]}"
    return newMessage
}

// parseMessage parses every message that comes in and puts it into a Go struct.
// TODO: make this generic to parse rest of the arguments
func parseMessage(msg []byte) *MessageHolder {
    var parsedMessage MessageHolder
    
    utils.LogMessage(string(msg), utils.RESISTANCE_LOG_PATH)
    err := json.Unmarshal(msg, &parsedMessage)
    if err != nil {
        utils.LogMessage("Error parsing message: " + string(msg), utils.RESISTANCE_LOG_PATH)
        return &MessageHolder{}
    }
    
    return &parsedMessage
}

// getUser extracts the user from the parsed message returned from parseMessage().
func getUser(parsedMessage *MessageHolder) *users.User {
    cookies := make([]*http.Cookie, 1)
    parsedCookie := strings.Split(parsedMessage.UserCookie, "=")
    cookies[0] = &http.Cookie{Name:parsedCookie[0], Value:parsedCookie[1]}
    user, success := users.ValidateUserCookie(cookies)
    if !success {
        utils.LogMessage("Something went wrong when validating the user", utils.RESISTANCE_LOG_PATH)
        return nil
    }
    
    return user
}

func main() {
    // Setup ZMQ
    context, _ := zmq.NewContext()
    zmqSocket, _ := context.NewSocket(zmq.REP)
    
    defer context.Close()
    defer zmqSocket.Close()
    
    zmqSocket.Bind("tcp://*:" + utils.GAME_REP_REQ_PORT)
    utils.LogMessage("Game server started, bound to port " + utils.GAME_REP_REQ_PORT, utils.RESISTANCE_LOG_PATH)
    for {
        reply, _ := zmqSocket.Recv(0)
        parsedMessage := parseMessage(reply)
        utils.LogMessage(parsedMessage.Message, utils.RESISTANCE_LOG_PATH)
        
        user := getUser(parsedMessage)
        
        var message string
        switch {
            case user == nil: message = ""
            case parsedMessage.Message == "playerConnect": message = handlePlayerConnect(parsedMessage, user)
            case parsedMessage.Message == "playerDisconnect": message = handlePlayerDisconnect(parsedMessage)
            default: message = ""
        }
        
        zmqSocket.Send([]byte(message), 0)
    }
}

