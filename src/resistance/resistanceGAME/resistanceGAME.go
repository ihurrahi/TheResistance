package main

import (
	"encoding/json"
	zmq "github.com/alecthomas/gozmq"
	"net/http"
	"resistance/game"
	"resistance/persist"
	"resistance/users"
	"resistance/utils"
	"strconv"
	"strings"
)

const (
	USER_COOKIE_KEY   = "userCookie"
	MESSAGE_KEY       = "message"
	GAME_TITLE_KEY    = "title"
	HOST_ID_KEY       = "hostId"
	GAME_ID_KEY       = "gameId"
	IS_HOST_KEY       = "isHost"
	PLAYERS_KEY       = "players"
	ACCEPT_USER_KEY   = "acceptUser"
	USER_ID_KEY       = "userId"
	ROLE_KEY          = "role"
	IS_LEADER_KEY     = "isLeader"
	TEAMS_KEY         = "team"
	TEAM_SIZE_KEY     = "teamSize"
	VOTE_KEY          = "vote"
	USERNAME_KEY      = "username"
	IS_ON_MISSION_KEY = "isOnMission"
	OUTCOME_KEY       = "outcome"
	GAME_WINNER_KEY   = "winner"
	MISSIONS_KEY      = "missions"
	ERROR_KEY         = "error"

	// messages received from the frontend
	CREATE_GAME_MESSAGE         = "createGame"
	IS_VALID_GAME_MESSAGE       = "isValidGame"
	PLAYER_CONNECT_MESSAGE      = "playerConnect"
	PLAYER_DISCONNECT_MESSAGE   = "playerDisconnect"
	GET_PLAYERS_MESSAGE         = "getPlayers"
	START_GAME_MESSAGE          = "startGame"
	QUERY_ROLE_MESSAGE          = "queryRole"
	QUERY_LEADER_MESSAGE        = "queryLeader"
	START_MISSION_MESSAGE       = "startMission"
	APPROVE_TEAM_MESSAGE        = "approveTeam"
	QUERY_IS_ON_MISSION_MESSAGE = "queryIsOnMission"
	MISSION_OUTCOME_MESSAGE     = "missionOutcome"

	// messages sent to the frontend
	PLAYER_CONNECT_SUCCESSFUL_MESSAGE  = "playerConnectSuccessful"
	PLAYERS_MESSAGE                    = "players"
	GAME_STARTED_MESSAGE               = "gameStarted"
	QUERY_ROLE_RESULT_MESSAGE          = "queryRoleResult"
	QUERY_LEADER_RESULT_MESSAGE        = "queryLeaderResult"
	MISSION_PREPARATION_MESSAGE        = "missionPreparation"
	TEAM_APPROVAL_MESSAGE              = "teamApproval"
	APPROVE_TEAM_UPDATE_MESSAGE        = "approveTeamUpdate"
	MISSION_STARTED_MESSAGE            = "missionStarted"
	QUERY_IS_ON_MISSION_RESULT_MESSAGE = "queryIsOnMissionResult"
	GAME_OVER_MESSAGE                  = "gameOver"
	MISSIONS_MESSAGE                   = "missions"
)

var persister *persist.Persister

func init() {
	persister = persist.NewPersister()
}

// handleCreateGame handlers the message that is sent when a
// request is made from the HTTP module to create a new game
func handleCreateGame(parsedMessage map[string]interface{}, connectingPlayer *users.User) map[string]interface{} {
	var returnMessage = make(map[string]interface{})
	newGame := game.NewGame(parsedMessage[GAME_TITLE_KEY].(string), parsedMessage[HOST_ID_KEY].(string), persister)
	if newGame != nil {
		returnMessage[GAME_ID_KEY] = newGame.GameId
	}
	return returnMessage
}

// handleIsValidGame takes in a game id and validates that it is
// ok for the given user to join the given game
func handleIsValidGame(gameIdString string, requestUser *users.User) map[string]interface{} {
	gameInfo := make(map[string]interface{})

	// Error if no game id is not specified
	if gameIdString == "" {
		gameInfo[ERROR_KEY] = "Game not specified."
		return gameInfo
	}

	// Error if game id can't be parsed
	gameId, err := strconv.Atoi(gameIdString)
	if err != nil {
		gameInfo[ERROR_KEY] = "Game Id is not valid."
		return gameInfo
	}

	requestedGame := persister.ReadGame(gameId)
	if requestedGame != nil {
		gameStatus := requestedGame.GameStatus
		switch {
		default:
		case gameStatus == game.STATUS_DONE:
			gameInfo[ERROR_KEY] = "Cannot join a game that is already done."
			return gameInfo
		case gameStatus == game.STATUS_IN_PROGRESS:
			// make sure that the player is an actual player of the game
			if !requestedGame.IsPlayer(requestUser) {
				gameInfo[ERROR_KEY] = "Cannot join a game that is in progress"
				return gameInfo
			}
		case gameStatus == game.STATUS_LOBBY:
			// make sure we're not going over the limit of 10 players
			// we are assuming the player is no longer in the list of players
			// if the player leaves the game while in the lobby status
			if len(requestedGame.Players) >= 10 {
				gameInfo[ERROR_KEY] = "Game has reach maximum capacity"
				return gameInfo
			}
		}
	} else {
		gameInfo[ERROR_KEY] = "Game does not exist."
		return gameInfo
	}

	// If we got here, it means we are good to go.
	gameInfo["GameTitle"] = requestedGame.Title
	return gameInfo
}

// handlePlayerConnect handles the message that is sent when a player
// first connects by loading the game page.
func handlePlayerConnect(currentGame *game.Game, connectingPlayer *users.User, pubSocket *zmq.Socket) map[string]interface{} {
	utils.LogMessage("Player "+strconv.Itoa(connectingPlayer.UserId)+" connecting", utils.RGAME_LOG_PATH)

	var returnMessage = make(map[string]interface{})
	gameId := currentGame.GameId

	// Add the user to the players for this game
	currentGame.AddPlayer(connectingPlayer)

	// Send a message to everyone about the new players
	playersMessage := getPlayersMessage(currentGame)
	sendMessageToSubscribers(gameId, playersMessage, pubSocket)

	// Also send a message back through the proxy to start a subscriber
	// for this player
	returnMessage[MESSAGE_KEY] = PLAYER_CONNECT_SUCCESSFUL_MESSAGE
	returnMessage[GAME_ID_KEY] = gameId
	returnMessage[ACCEPT_USER_KEY] = true
	returnMessage[USER_ID_KEY] = connectingPlayer.UserId

	if currentGame.Host.UserId == connectingPlayer.UserId {
		returnMessage[IS_HOST_KEY] = true
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

// handleGetPlayers handles the message that is sent when the
// frontend needs an update on the players
func handleGetPlayers(currentGame *game.Game) map[string]interface{} {
	return getPlayersMessage(currentGame)
}

// handleStartGame handles the message that is sent when the host
// presses the start game button.
func handleStartGame(currentGame *game.Game, connectingPlayer *users.User, pubSocket *zmq.Socket) map[string]interface{} {
	var returnMessage = make(map[string]interface{})
	gameId := currentGame.GameId

	_ = currentGame.StartGame()

	// Sends the message that the game has officially started
	var gameStartedMessage = make(map[string]interface{})
	gameStartedMessage[MESSAGE_KEY] = GAME_STARTED_MESSAGE
	sendMessageToSubscribers(gameId, gameStartedMessage, pubSocket)

	_ = game.NewMission(currentGame)

	// Send a message to everyone to update their missions view
	sendMissionsMessage(currentGame, pubSocket)

	// Sends the message that a mission is going to start
	var missionPreparationMessage = make(map[string]interface{})
	missionPreparationMessage[MESSAGE_KEY] = MISSION_PREPARATION_MESSAGE
	sendMessageToSubscribers(gameId, missionPreparationMessage, pubSocket)

	return returnMessage
}

// handleQueryRole handles the request from the frontend for which
// team they are on.
func handleQueryRole(currentGame *game.Game, player *users.User) map[string]interface{} {
	var returnMessage = make(map[string]interface{})

	for _, singlePlayer := range currentGame.Players {
		if singlePlayer.User.UserId == player.UserId {
			returnMessage[MESSAGE_KEY] = QUERY_ROLE_RESULT_MESSAGE
			returnMessage[ROLE_KEY] = singlePlayer.Role
			break
		}
	}

	return returnMessage
}

// handleQueryLeader handles the request from the frontend for who
// the leader of the current mission is.
func handleQueryLeader(currentGame *game.Game, player *users.User) map[string]interface{} {
	var returnMessage = make(map[string]interface{})

	isLeader := currentGame.GetCurrentMission().IsUserCurrentMissionLeader(player)

	returnMessage[MESSAGE_KEY] = QUERY_LEADER_RESULT_MESSAGE
	returnMessage[IS_LEADER_KEY] = isLeader

	if isLeader {
		returnMessage[PLAYERS_KEY] = getPlayerUsernames(currentGame)
		returnMessage[TEAM_SIZE_KEY] = currentGame.GetCurrentMission().GetCurrentMissionTeamSize()
	}

	return returnMessage
}

// handleStartMission handles the message when the leader
// sends in the team.
func handleStartMission(message map[string]interface{}, currentGame *game.Game, connectingPlayer *users.User, pubSocket *zmq.Socket) map[string]interface{} {
	// TODO validate user is mission leader

	var returnMessage = make(map[string]interface{})
	teamIds := make([]string, 0)
	rawTeamIds, ok := message[TEAMS_KEY].([]interface{})
	if ok {
		for _, rawTeamId := range rawTeamIds {
			teamId, ok := rawTeamId.(string)
			if ok {
				teamIds = append(teamIds, teamId)
			}
		}
	}
	teamUsernames := make([]string, len(teamIds))
	teamUsers := make([]*users.User, len(teamIds))
	for i, teamId := range teamIds {
		parsedTeamId, _ := strconv.Atoi(teamId)
		user := users.LookupUserById(parsedTeamId)
		if user.IsValidUser() {
			teamUsernames[i] = user.Username
			teamUsers[i] = user
		} else {
			utils.LogMessage("User Id for team not found: "+teamId, utils.RGAME_LOG_PATH)
		}
	}

	gameId := currentGame.GameId
	currentGame.GetCurrentMission().CreateTeam(teamUsers)

	var teamApprovalMessage = make(map[string]interface{})
	teamApprovalMessage[MESSAGE_KEY] = TEAM_APPROVAL_MESSAGE
	teamApprovalMessage[TEAMS_KEY] = teamUsernames
	sendMessageToSubscribers(gameId, teamApprovalMessage, pubSocket)

	sendMissionsMessage(currentGame, pubSocket)

	return returnMessage
}

// handleApproveTeam handles the message from the frontend
// that votes for the whether the team can go on the mission.
func handleApproveTeam(message map[string]interface{}, currentGame *game.Game, connectingPlayer *users.User, pubSocket *zmq.Socket) map[string]interface{} {
	var returnMessage = make(map[string]interface{})

	gameId := currentGame.GameId
	vote, ok := message[VOTE_KEY].(bool)
	if ok {
		currentGame.GetCurrentMission().AddVote(connectingPlayer, vote)

		// send vote to everyone to make it public
		var approveTeamUpdateMessage = make(map[string]interface{})
		approveTeamUpdateMessage[MESSAGE_KEY] = APPROVE_TEAM_UPDATE_MESSAGE
		approveTeamUpdateMessage[USERNAME_KEY] = connectingPlayer.Username
		approveTeamUpdateMessage[VOTE_KEY] = vote
		sendMessageToSubscribers(gameId, approveTeamUpdateMessage, pubSocket)

		allVotesIn := currentGame.GetCurrentMission().IsAllVotesCollected()
		if allVotesIn {
			missionApproved := currentGame.GetCurrentMission().IsTeamApproved()
			if missionApproved {
				var missionApprovedMessage = make(map[string]interface{})
				missionApprovedMessage[MESSAGE_KEY] = MISSION_STARTED_MESSAGE
				sendMessageToSubscribers(gameId, missionApprovedMessage, pubSocket)
			} else {
				currentGame.GetCurrentMission().EndMission(game.RESULT_NONE)

				_ = game.NewMission(currentGame)

				var missionPreparationMessage = make(map[string]interface{})
				missionPreparationMessage[MESSAGE_KEY] = MISSION_PREPARATION_MESSAGE
				sendMessageToSubscribers(gameId, missionPreparationMessage, pubSocket)
			}

			// once all votes are in, if either the mission was approved or not
			// there is an update to the list of missions so we should send it out.
			sendMissionsMessage(currentGame, pubSocket)
		}
	}

	return returnMessage
}

// handleQueryIsOnMission handles the message from the frontend
// asking if the requesting user is on the current mission.
// Assumes that the mission has been approved.
func handleQueryIsOnMission(currentGame *game.Game, connectingPlayer *users.User) map[string]interface{} {
	var returnMessage = make(map[string]interface{})

	isOnMission := currentGame.GetCurrentMission().IsUserOnCurrentMission(connectingPlayer)

	returnMessage[MESSAGE_KEY] = QUERY_IS_ON_MISSION_RESULT_MESSAGE
	returnMessage[IS_ON_MISSION_KEY] = isOnMission

	return returnMessage
}

// handleMissionOutcome handles the message from the frontend
// after a player has put in their mission outcome - a "success"
// or a "fail"
func handleMissionOutcome(message map[string]interface{}, currentGame *game.Game, connectingPlayer *users.User, pubSocket *zmq.Socket) map[string]interface{} {
	var returnMessage = make(map[string]interface{})

	gameId := currentGame.GameId
	missionOutcome, ok := message[OUTCOME_KEY].(bool)
	if ok {
		currentGame.GetCurrentMission().AddOutcome(connectingPlayer, missionOutcome)

		// check if the current mission is over
		isMissionOver, result := currentGame.GetCurrentMission().IsMissionOver()
		if isMissionOver {
			// it is, so set the mission result
			currentGame.GetCurrentMission().EndMission(result)

			// now check if the game is over
			isGameOver, winner := currentGame.IsGameOver()

			if isGameOver {
				currentGame.EndGame()

				// send game over message
				var gameOverMessage = make(map[string]interface{})
				gameOverMessage[MESSAGE_KEY] = GAME_OVER_MESSAGE
				gameOverMessage[GAME_WINNER_KEY] = winner
				sendMessageToSubscribers(gameId, gameOverMessage, pubSocket)
			} else {
				_ = game.NewMission(currentGame)

				// send mission preparation message for next mission
				var missionPreparationMessage = make(map[string]interface{})
				missionPreparationMessage[MESSAGE_KEY] = MISSION_PREPARATION_MESSAGE
				sendMessageToSubscribers(gameId, missionPreparationMessage, pubSocket)

				sendMissionsMessage(currentGame, pubSocket)
			}

		}
	}
	return returnMessage
}

// getPlayersMessage builds up the message to update the list of
// current players
func getPlayersMessage(currentGame *game.Game) map[string]interface{} {
	usernames := getPlayerUsernames(currentGame)

	// Build up players message.
	var playersMessage = make(map[string]interface{})
	playersMessage[MESSAGE_KEY] = PLAYERS_MESSAGE
	playersMessage[PLAYERS_KEY] = usernames
	playersMessage[GAME_ID_KEY] = currentGame.GameId

	return playersMessage
}

func getPlayerUsernames(currentGame *game.Game) []string {
	var usernames = make([]string, 0)
	players := currentGame.Players

	for _, player := range players {
		usernames = append(usernames, player.User.Username)
	}

	return usernames
}

// parseMessage parses every message that comes in and puts it into a Go struct.
func parseMessage(msg []byte) map[string]interface{} {
	var parsedMessage = make(map[string]interface{})

	utils.LogMessage(string(msg), utils.RGAME_LOG_PATH)
	err := json.Unmarshal(msg, &parsedMessage)
	if err != nil {
		utils.LogMessage("Error parsing message: "+string(msg), utils.RGAME_LOG_PATH)
	}

	return parsedMessage
}

// getUser extracts the user from the parsed message returned from parseMessage().
func getUser(parsedMessage map[string]interface{}) *users.User {
	cookies := make([]*http.Cookie, 1)
	parsedCookie := strings.Split(parsedMessage[USER_COOKIE_KEY].(string), "=")
	cookies[0] = &http.Cookie{Name: parsedCookie[0], Value: parsedCookie[1]}
	user, success := users.ValidateUserCookie(cookies)
	if !success {
		utils.LogMessage("Something went wrong when validating the user", utils.RGAME_LOG_PATH)
		return nil
	}

	return user
}

// sendMissionsMessage sends the update mission info message to all subscribers
// to the given game id
func sendMissionsMessage(currentGame *game.Game, pubSocket *zmq.Socket) {
	gameId := currentGame.GameId

	var missionInfoMessage = make(map[string]interface{})
	missionInfoMessage[MESSAGE_KEY] = MISSIONS_MESSAGE

	missionInfo := currentGame.GetMissionInfo()
	missionInfoMessage[MISSIONS_KEY] = missionInfo

	sendMessageToSubscribers(gameId, missionInfoMessage, pubSocket)
}

// sendMessageToSubscribers is a helper method to send the given message to the given
// publisher socket with the given gameId filter
func sendMessageToSubscribers(gameId int, message map[string]interface{}, pubSocket *zmq.Socket) {
	pubMessage, err := json.Marshal(message)
	if err == nil {
		// Send out updated users to all subscribers to this game
		pubSocket.SendMultipart([][]byte{[]byte(strconv.Itoa(gameId)), []byte(pubMessage)}, 0)

		utils.LogMessage("Sent message to all subscribers to game "+strconv.Itoa(gameId), utils.RGAME_LOG_PATH)
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
	utils.LogMessage("Game server started, bound to port "+utils.GAME_REP_REQ_PORT, utils.RGAME_LOG_PATH)
	pubSocket.Bind("tcp://*:" + utils.GAME_PUB_SUB_PORT)
	utils.LogMessage("Game server started, bound to port "+utils.GAME_PUB_SUB_PORT, utils.RGAME_LOG_PATH)

	for {
		reply, _ := zmqSocket.Recv(0)
		parsedMessage := parseMessage(reply)

		var returnMessage = make(map[string]interface{})

		user := getUser(parsedMessage)
		gameIdString, _ := parsedMessage[GAME_ID_KEY].(string)

		if parsedMessage[MESSAGE_KEY] == IS_VALID_GAME_MESSAGE {
			returnMessage = handleIsValidGame(gameIdString, user)
		} else {

			gameId, err := strconv.Atoi(gameIdString)

			if err == nil {

				currentGame := persister.ReadGame(gameId)

				switch {
				default:
				case user == nil:
				case parsedMessage[MESSAGE_KEY] == CREATE_GAME_MESSAGE:
					returnMessage = handleCreateGame(parsedMessage, user)
				case parsedMessage[MESSAGE_KEY] == PLAYER_CONNECT_MESSAGE:
					returnMessage = handlePlayerConnect(currentGame, user, pubSocket)
				case parsedMessage[MESSAGE_KEY] == PLAYER_DISCONNECT_MESSAGE:
					returnMessage = handlePlayerDisconnect(parsedMessage)
				case parsedMessage[MESSAGE_KEY] == GET_PLAYERS_MESSAGE:
					returnMessage = handleGetPlayers(currentGame)
				case parsedMessage[MESSAGE_KEY] == START_GAME_MESSAGE:
					returnMessage = handleStartGame(currentGame, user, pubSocket)
				case parsedMessage[MESSAGE_KEY] == QUERY_ROLE_MESSAGE:
					returnMessage = handleQueryRole(currentGame, user)
				case parsedMessage[MESSAGE_KEY] == QUERY_LEADER_MESSAGE:
					returnMessage = handleQueryLeader(currentGame, user)
				case parsedMessage[MESSAGE_KEY] == START_MISSION_MESSAGE:
					returnMessage = handleStartMission(parsedMessage, currentGame, user, pubSocket)
				case parsedMessage[MESSAGE_KEY] == APPROVE_TEAM_MESSAGE:
					returnMessage = handleApproveTeam(parsedMessage, currentGame, user, pubSocket)
				case parsedMessage[MESSAGE_KEY] == QUERY_IS_ON_MISSION_MESSAGE:
					returnMessage = handleQueryIsOnMission(currentGame, user)
				case parsedMessage[MESSAGE_KEY] == MISSION_OUTCOME_MESSAGE:
					returnMessage = handleMissionOutcome(parsedMessage, currentGame, user, pubSocket)
				}
			}
		}

		marshalledMessage, err := json.Marshal(returnMessage)
		if err != nil {
			utils.LogMessage("Error marshalling response", utils.RGAME_LOG_PATH)
			marshalledMessage = make([]byte, 0)
		}
		zmqSocket.Send(marshalledMessage, 0)
	}
}
