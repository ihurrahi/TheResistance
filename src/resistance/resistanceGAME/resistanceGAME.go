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
    ROLE_KEY = "role"
    IS_LEADER_KEY = "isLeader"
    
    // messages received from the frontend
    PLAYER_CONNECT_MESSAGE = "playerConnect"
    PLAYER_DISCONNECT_MESSAGE = "playerDisconnect"
    START_GAME_MESSAGE = "startGame"
    QUERY_ROLE_MESSAGE = "queryRole"
    QUERY_LEADER_MESSAGE = "queryLeader"
    
    // messages sent to the frontend
    PLAYERS_MESSAGE = "players"
    GAME_STARTED_MESSAGE = "gameStarted"
    QUERY_ROLE_RESULT_MESSAGE = "queryRoleResult"
    QUERY_LEADER_RESULT_MESSAGE = "queryLeaderResult"
    MISSION_PREPARATION_MESSAGE = "missionPreparation"
)

// handlePlayerConnect handles the message that is sent when a player
// first connects by loading the game page.
func handlePlayerConnect(message map[string]interface{}, connectingPlayer *users.User, pubSocket *zmq.Socket) map[string]interface{} {
    var returnMessage = make(map[string]interface{})
    utils.LogMessage(connectingPlayer.Username + " has sent a message!", utils.RESISTANCE_LOG_PATH)
    // TODO: handle case where game_id_key doesnt exist
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
        returnMessage[MESSAGE_KEY] = PLAYERS_MESSAGE
        returnMessage[PLAYERS_KEY] = usernames
        returnMessage[GAME_ID_KEY] = gameId
    
        sendMessageToSubscribers(gameId, returnMessage, pubSocket)
        
        // Add a few more items to tell the proxy to start a subscriber
        // for this player
        returnMessage[ACCEPT_USER_KEY] = true
        returnMessage[USER_ID_KEY] = connectingPlayer.UserId
    }
    
    return returnMessage
}

// handlerPlayerDisconnect handles the message that is sent when
// a player disconnects from the web socket proxy.
func handlePlayerDisconnect(message map[string]interface{}) map[string]interface{} {
    var returnMessage = make(map[string]interface{})
    // TODO: implement
    return returnMessage
}

// handleStartGame handles the message that is sent when the host
// presses the start game button.
func handleStartGame(message map[string]interface{}, connectingPlayer *users.User, pubSocket *zmq.Socket) map[string]interface{} {
    var returnMessage = make(map[string]interface{})
    gameId, err := strconv.Atoi(message[GAME_ID_KEY].(string))
    if err == nil {
        err = game.SetGameStatus(gameId, game.GAME_STATUS_IN_PROGRESS)
        // TODO: error check here
        
        err = game.AssignPlayerRoles(gameId)
        // TODO: error check here
        
        // Sends the message that the game has officially started
        var gameStartedMessage = make(map[string]interface{})
        gameStartedMessage[MESSAGE_KEY] = GAME_STARTED_MESSAGE
        sendMessageToSubscribers(gameId, gameStartedMessage, pubSocket)
        
        err = game.StartNextMission(gameId)
        // TODO: error check here
        
        // Sends the message that a mission is going to start
        var missionPreparationMessage = make(map[string]interface{})
        missionPreparationMessage[MESSAGE_KEY] = MISSION_PREPARATION_MESSAGE
        sendMessageToSubscribers(gameId, missionPreparationMessage, pubSocket)
    }
    // TODO: error check if gameId is not an integer or not given
    
    return returnMessage
}

// handleQueryRole handles the request from the frontend for which
// team they are on.
func handleQueryRole(message map[string]interface{}, player *users.User) map[string]interface{} {
    var returnMessage = make(map[string]interface{})
    gameId, err := strconv.Atoi(message[GAME_ID_KEY].(string))
    if err == nil {
        role, err := game.GetPlayerRole(player.UserId, gameId)
        if err == nil {
            returnMessage[MESSAGE_KEY] = QUERY_ROLE_RESULT_MESSAGE
            returnMessage[ROLE_KEY] = role
        }
        // TODO: error checking
    }
    // TODO: error checking
    
    return returnMessage
}

// handleQueryLeader handles the request from the frontend for who
// the leader of the current mission is.
func handleQueryLeader(message map[string]interface{}, player *users.User) map[string]interface{} {
    var returnMessage = make(map[string]interface{})
    gameId, err := strconv.Atoi(message[GAME_ID_KEY].(string))
    if err == nil {
        isLeader, err := game.IsUserMissionLeader(player.UserId, gameId)
        if err == nil {
            returnMessage[MESSAGE_KEY] = QUERY_LEADER_RESULT_MESSAGE
            returnMessage[IS_LEADER_KEY] = isLeader
            
            if isLeader {
                allPlayers, err := game.GetPlayers(gameId)
                if err == nil {
                    returnMessage[PLAYERS_KEY] = allPlayers
                }
                // TODO: error checking
            }
        }
        // TODO: error checking
        
    }
    // TODO: error checking
    
    return returnMessage
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

// sendMessageToSubscribers is a helper method to send the given message to the given
// publisher socket with the given gameId filter
func sendMessageToSubscribers(gameId int, message map[string]interface{}, pubSocket *zmq.Socket) {
    pubMessage, err := json.Marshal(message)
    if err == nil {
        // Send out updated users to all subscribers to this game
        pubSocket.SendMultipart([][]byte{[]byte(strconv.Itoa(gameId)), []byte(pubMessage)}, 0)
        
        utils.LogMessage("Sent message to all subscribers to game " + strconv.Itoa(gameId), utils.RESISTANCE_LOG_PATH)
    }
    // TODO: error check in case marshalling failed
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
        
        user := getUser(parsedMessage)
        
        var returnMessage = make(map[string]interface{})
        switch {
            default:
            case user == nil:
            case parsedMessage[MESSAGE_KEY] == PLAYER_CONNECT_MESSAGE:
                returnMessage = handlePlayerConnect(parsedMessage, user, pubSocket)
            case parsedMessage[MESSAGE_KEY] == PLAYER_DISCONNECT_MESSAGE:
                returnMessage = handlePlayerDisconnect(parsedMessage)
            case parsedMessage[MESSAGE_KEY] == START_GAME_MESSAGE:
                returnMessage = handleStartGame(parsedMessage, user, pubSocket)
            case parsedMessage[MESSAGE_KEY] == QUERY_ROLE_MESSAGE:
                returnMessage = handleQueryRole(parsedMessage, user)
            case parsedMessage[MESSAGE_KEY] == QUERY_LEADER_MESSAGE:
                returnMessage = handleQueryLeader(parsedMessage, user)
        }
        
        marshalledMessage, err := json.Marshal(returnMessage)
        if err != nil {
            utils.LogMessage("Error marshalling response", utils.RESISTANCE_LOG_PATH)
            marshalledMessage = make([]byte,0)
        }
        zmqSocket.Send(marshalledMessage, 0)
    }
}

