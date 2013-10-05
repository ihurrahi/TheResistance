package main 

import (
    "net/http"
    "strings"
    "encoding/json"
    zmq "github.com/alecthomas/gozmq"
    "resistance/utils"
    "resistance/users"
)

const (
    USER_COOKIE = "userCookie"
    MESSAGE = "message"
)

// handlerPlayerDisconnect handles the message that is sent when
// a player disconnects from the web socket proxy.
func handlePlayerDisconnect(message map[string]interface{}) string {
    // TODO: implement
    return ""
}

// handlePlayerConnect handles the message that is sent when a player
// first connects by loading the game page.
func handlePlayerConnect(message map[string]interface{}, user *users.User) string {
    // TODO: implement sending a message to SUB zmqSocket
    utils.LogMessage(user.Username + " has sent a message!", utils.RESISTANCE_LOG_PATH)
    newMessage := "{\"message\":\"players\",\"acceptUser\":true,\"players\":[\"" + user.Username + "\"]}"
    return newMessage
}

// parseMessage parses every message that comes in and puts it into a Go struct.
func parseMessage(msg []byte) map[string]interface{} {
    var parsedMessage = make(map[string]interface{})
    
    utils.LogMessage(string(msg), utils.RESISTANCE_LOG_PATH)
    err := json.Unmarshal(msg, &parsedMessage)
    if err != nil {
        utils.LogMessage("Error parsing message: " + string(msg), utils.RESISTANCE_LOG_PATH)
    }
    
    return parsedMessage
}

// getUser extracts the user from the parsed message returned from parseMessage().
func getUser(parsedMessage map[string]interface{}) *users.User {
    cookies := make([]*http.Cookie, 1)
    parsedCookie := strings.Split(parsedMessage[USER_COOKIE].(string), "=")
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
        utils.LogMessage(parsedMessage[MESSAGE].(string), utils.RESISTANCE_LOG_PATH)
        
        user := getUser(parsedMessage)
        
        var returnMessage string
        switch {
            case user == nil: returnMessage = ""
            case parsedMessage[MESSAGE] == "playerConnect": returnMessage = handlePlayerConnect(parsedMessage, user)
            case parsedMessage[MESSAGE] == "playerDisconnect": returnMessage = handlePlayerDisconnect(parsedMessage)
            default: returnMessage = ""
        }
        
        zmqSocket.Send([]byte(returnMessage), 0)
    }
}

