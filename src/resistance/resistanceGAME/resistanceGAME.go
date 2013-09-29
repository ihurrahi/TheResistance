package main 

import (
    "net/http"
    "strings"
    "strconv"
    "encoding/json"
    zmq "github.com/alecthomas/gozmq"
    "resistance/utils"
    "resistance/users"
)

type MessageHolder struct {
    Message string
    UserCookie string
}

func notifyNewPlayer(message *MessageHolder) string {
    cookies := make([]*http.Cookie, 1)
    parsedCookie := strings.Split(message.UserCookie, "=")
    cookies[0] = &http.Cookie{Name:parsedCookie[0], Value:parsedCookie[1]}
    user, success := users.ValidateUserCookie(cookies)
    if !success {
        utils.LogMessage("Something went wrong when validating the user", utils.RESISTANCE_LOG_PATH)
    }

    utils.LogMessage(user.Username + " has sent a message!", utils.RESISTANCE_LOG_PATH)
    newMessage := "{\"Message\":\"newPlayer\",\"UserName\":\"" + user.Username + "\"}"
    return newMessage
}

func parseMessage(msg string) *MessageHolder {
    var parsedMessage MessageHolder
    
    utils.LogMessage(msg, utils.RESISTANCE_LOG_PATH)
    err := json.Unmarshal([]byte(msg), &parsedMessage)
    if err != nil {
        utils.LogMessage("Error parsing message: " + msg, utils.RESISTANCE_LOG_PATH)
        return nil
    }
    
    return &parsedMessage
}

func main() {
    // Setup ZMQ
    context, _ := zmq.NewContext()
    zmqSocket, _ := context.NewSocket(zmq.REP)
    
    defer context.Close()
    defer zmqSocket.Close()
    
    zmqSocket.Bind("tcp://*:" + strconv.Itoa(utils.GAME_REP_REQ_PORT))
    utils.LogMessage("Game server started, bound to port " + strconv.Itoa(utils.GAME_REP_REQ_PORT), utils.RESISTANCE_LOG_PATH)
    for {
        reply, _ := zmqSocket.Recv(0)
        parsedMessage := parseMessage(string(reply))
        utils.LogMessage(parsedMessage.Message, utils.RESISTANCE_LOG_PATH)
        
        var message string
        
        switch {
            case parsedMessage.Message == "firstConnection": message = notifyNewPlayer(parsedMessage)
            default: message = ""
        }
        
        zmqSocket.Send([]byte(message), 0)
    }
}

