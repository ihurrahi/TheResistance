package main 

import (
    "net/http"
    "strings"
    "strconv"
    "encoding/json"
    zmq "github.com/alecthomas/gozmq"
    "resistance/utils"
    "resistance/users"
    "resistance/game"
)

const (
    USER_COOKIE_KEY = "userCookie"
    MESSAGE_KEY = "message"
    GAME_ID_KEY = "gameId"
    PLAYERS_KEY = "players"
    ACCEPT_USER_KEY = "acceptUser"
    USER_ID_KEY = "userId"
    PLAYERS_MESSAGE = "players"
)

// handlerPlayerDisconnect handles the message that is sent when
// a player disconnects from the web socket proxy.
func handlePlayerDisconnect(message map[string]interface{}) string {
    // TODO: implement
    return ""
}

// handlePlayerConnect handles the message that is sent when a player
// first connects by loading the game page.
func handlePlayerConnect(message map[string]interface{}, connectingPlayer *users.User, pubSocket *zmq.Socket) string {
    newMessage := ""
    utils.LogMessage(connectingPlayer.Username + " has sent a message!", utils.RESISTANCE_LOG_PATH)
    gameId, err := strconv.Atoi(message[GAME_ID_KEY].(string))
    if err == nil {
        // Add the user to the players for this game
        game.AddPlayer(gameId, connectingPlayer.UserId)
    
        // Retrive all users for this game.
        var usernames = make([]string, 0)
        users, err := game.GetPlayers(gameId)
        if err == nil {
            for _, user := range users {
                usernames = append(usernames, user.Username)
            }
        }
        // TODO: error check here
        utils.LogMessage(strconv.Itoa(len(usernames)), utils.RESISTANCE_LOG_PATH)
        
        // Build up message.
        var message = make(map[string]interface{})
        message[MESSAGE_KEY] = PLAYERS_MESSAGE
        message[PLAYERS_KEY] = usernames
        message[GAME_ID_KEY] = gameId
    
        pubMessage, err := json.Marshal(message)
        if err == nil {
            utils.LogMessage(string(pubMessage), utils.RESISTANCE_LOG_PATH)
            // Send out updated users to all subscribers to this game
            pubSocket.SendMultipart([][]byte{[]byte(strconv.Itoa(gameId)), []byte(pubMessage)}, 0)
        }
        // TODO: error check here
        
        // Add a few more items to tell the proxy to start a subscriber
        // for this player
        message[ACCEPT_USER_KEY] = true
        message[USER_ID_KEY] = connectingPlayer.UserId
        
        potentialMessage, err := json.Marshal(message)
        if err != nil {
            newMessage = ""
        } else {
            newMessage = string(potentialMessage)
        }
    }
    
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
    parsedCookie := strings.Split(parsedMessage[USER_COOKIE_KEY].(string), "=")
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
    pubSocket, _ := context.NewSocket(zmq.PUB)
    
    defer context.Close()
    defer zmqSocket.Close()
    defer pubSocket.Close()
    
    zmqSocket.Bind("tcp://*:" + utils.GAME_REP_REQ_PORT)
    utils.LogMessage("Game server started, bound to port " + utils.GAME_REP_REQ_PORT, utils.RESISTANCE_LOG_PATH)
    pubSocket.Bind("tcp://*:" + utils.GAME_PUB_SUB_PORT)
    utils.LogMessage("Game server started, bound to port " + utils.GAME_PUB_SUB_PORT, utils.RESISTANCE_LOG_PATH)
    
    for {
        reply, _ := zmqSocket.Recv(0)
        parsedMessage := parseMessage(reply)
        utils.LogMessage(parsedMessage[MESSAGE_KEY].(string), utils.RESISTANCE_LOG_PATH)
        
        user := getUser(parsedMessage)
        
        var returnMessage string
        switch {
            case user == nil: returnMessage = ""
            case parsedMessage[MESSAGE_KEY] == "playerConnect":
                returnMessage = handlePlayerConnect(parsedMessage, user, pubSocket)
            case parsedMessage[MESSAGE_KEY] == "playerDisconnect":
                returnMessage = handlePlayerDisconnect(parsedMessage)
            default:
                returnMessage = ""
        }
        
        zmqSocket.Send([]byte(returnMessage), 0)
    }
}

