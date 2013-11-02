package game

import (
    "errors"
    "strconv"
    "net/http"
    "math/rand"
    "time"
    "resistance/utils"
    "resistance/users"
)

const (
    GAME_STATUS_LOBBY = "LOBBY"
    GAME_STATUS_IN_PROGRESS = "IN_PROGRESS"
    GAME_STATUS_DONE = "DONE"
    TITLE_KEY = "title"
    HOST_ID_KEY = "host"
    RESISTANCE_ROLE = "R"
    SPY_ROLE = "S"
    RESISTANCE_RESULT_STRING = "RESISTANCE"
    SPY_RESULT_STRING = "SPY"
    CANCELED_RESULT_STRING = "CANCELED"
    
    CREATE_GAME_QUERY = "insert into games (`title`, `host_id`, `status`) values (?, ?, \"" + GAME_STATUS_LOBBY + "\")"
    GET_GAME_NAME_QUERY = "select title from games where game_id = ?"
    ADD_PLAYER_QUERY = "insert into players (`game_id`, `user_id`) values (?, ?)"
    GET_PLAYERS_QUERY = "select user_id from players where game_id = ? order by join_date"
    GET_GAME_STATUS_QUERY = "select status from games where game_id = ?"
    SET_GAME_STATUS_QUERY = "update games set status = ? where game_id = ?"
    NUM_PLAYERS_QUERY = "select user_id from players where game_id = ?"
    SET_PLAYER_ROLE_QUERY = "update players set role = ? where user_id = ? and game_id = ?"
    PLAYER_ROLE_QUERY = "select role from players where user_id = ? and game_id = ?"
    MISSION_LEADER_QUERY = "select leader_id from missions where mission_id = (" + CURRENT_MISSION_ID_QUERY + ")"
    CURRENT_MISSION_NUM_QUERY = "select ifnull(max(mission_num),0) from missions where game_id = ?"
    CREATE_MISSION_QUERY = "insert into missions (`game_id`, `mission_num`, `leader_id`) values (?, ?, ?)"
    CREATE_TEAM_MEMBER_QUERY = "insert into teams (`mission_id`, `user_id`) values (?, ?)"
    CURRENT_MISSION_ID_QUERY = "select mission_id from missions where game_id = ? order by mission_id desc limit 1"
    ADD_VOTE_QUERY = "insert into votes (`mission_id`, `user_id`, `vote`) values (?, ?, ?)"
    ALL_VOTES_IN_QUERY = "select (select count(*) from votes where mission_id = ?) >= (select count(*) from players join missions on players.game_id = missions.game_id where mission_id = ?)"
    TEAM_APPROVED_QUERY = "select sum(vote) > sum(vote = 0) from votes where mission_id = ?"
    SET_MISSION_RESULT_QUERY = "update missions set result = ? where mission_id = ?"
    GET_CURRENT_MISSION_RESULT_QUERY = "select result from missions where mission_id = (" + CURRENT_MISSION_ID_QUERY + ")"
)

// numPlayersToNumSpies gives you how many spies there should be in a game
// for the given the number of players
var numPlayersToNumSpies = map[int]int {
    5:2,
    6:2,
    7:3,
    8:3,
    9:3,
    10:4}

// numPlayersOnTeam gives you how many players should be on a team
// given the total number of players and the mission number
var numPlayersOnTeam = map[int]map[int]int {
    5:{1:2, 2:3, 3:2, 4:3, 5:3},
    6:{1:2, 2:3, 3:4, 4:3, 5:4},
    7:{1:2, 2:3, 3:3, 4:4, 5:4},
    8:{1:3, 2:4, 3:4, 4:5, 5:5},
    9:{1:3, 2:4, 3:4, 4:5, 5:5},
    10:{1:3, 2:4, 3:4, 4:5, 5:5}}

// CreateGame creates the game by storing the relevant information
// in the games table in the DB.
func CreateGame(request *http.Request) (int64, error) {
    db, err := utils.ConnectToDB()
    if err != nil {
        return 0, err
    }

    title := request.FormValue(TITLE_KEY)
    hostId := request.FormValue(HOST_ID_KEY)
    result, err := db.Exec(CREATE_GAME_QUERY, title, hostId)
    if err != nil {
        return 0, err
    }
    id, err := result.LastInsertId()
    if err != nil {
        return 0, err
    }
    return id, nil
}

// ValidateGameRequest takes in a game id and validates that it is
// ok for the given user to join the given game
func ValidateGameRequest(gameIdString string, user *users.User) (int, error) {

    // Error if no game id is not specified
    if gameIdString == "" {
        return -1, errors.New("Game not specified.")
    }
    
    // Error if game id can't be parsed
    gameId, err := strconv.Atoi(gameIdString)
    if err != nil {
        return -1, errors.New("Game Id is not valid.")
    }
    
    db, err := utils.ConnectToDB()
    if err != nil {
        return -1, err
    }
    results, err := db.Query(GET_GAME_STATUS_QUERY, gameId)
    if err != nil {
        return -1, err
    }
    // Error if no rows returned
    if results.Next() {
        var gameStatus string
        if err := results.Scan(&gameStatus); err == nil {
            // Error if game is already done.
            if gameStatus == GAME_STATUS_DONE {
                return -1, errors.New("Cannot join a game that is already done.")
            }
            if gameStatus == GAME_STATUS_IN_PROGRESS {
                // TODO: error check for joining games in progress
            }
        } else {
            return -1, err
        }
    } else {
        return -1, errors.New("Game does not exist.")
    }
    
    return gameId, nil
}

// AddPlayer adds the given user to the given game by storing the
// relevant information in the players table. This can only be done
// while the game is still in the LOBBY stage.
func AddPlayer(gameId int, userId int) error {
    db, err := utils.ConnectToDB()
    if err != nil {
        return err
    }

    // TODO: add validation that the game is still in LOBBY status
    _, err = db.Exec(ADD_PLAYER_QUERY, gameId, userId)
    if err != nil {
        return err
    }
    return nil
}

// GetGameName retrieves the game name from the database.
// This should only be used for display purposes.
func GetGameName(gameId int) (string, error) {
    db, err := utils.ConnectToDB()
    if err != nil {
        return "", err
    }

    results, err := db.Query(GET_GAME_NAME_QUERY, gameId)
    if err != nil {
        return "", err
    }
    
    if (results.Next()) {
        var gameTitle string
        if err := results.Scan(&gameTitle); err == nil {
            return gameTitle, nil
        } else {
            return "", err
        }
    }
    return "", nil
}

// DeletePlayer deletes the given user from the given game by
// removing the user/game pair from the players table. This can
// only be done while the game is still in the LOBBY stage.
func DeletePlayer(userId int, gameId int) {

}

// GetPlayers retrieves all the current players under the given
// game.
func GetPlayers(gameId int) ([]*users.User, error) {
    var playerList = make([]*users.User, 0)
    db, err := utils.ConnectToDB()
    if err != nil {
        return playerList, err
    }

    results, err := db.Query(GET_PLAYERS_QUERY, gameId)
    if err != nil {
        return playerList, err
    }
    
    for results.Next() {
        var userId int
        if err := results.Scan(&userId); err == nil {
            user := users.LookupUserById(userId)
            if user.IsValidUser() {
                playerList = append(playerList, user)
            } else {
                utils.LogMessage("User not found: " + strconv.Itoa(userId), utils.RESISTANCE_LOG_PATH)
            }
        }
    }
    
    return playerList, nil
}

// SetGameStatus sets the given status for the given game id.
func SetGameStatus(gameId int, status string) error {
    // TODO: validate status here
    db, err := utils.ConnectToDB()
    if err != nil {
        return err
    }

    _, err = db.Exec(SET_GAME_STATUS_QUERY, status, gameId)
    if err != nil {
        return err
    }
    return nil
}

// AssignPlayerRoles assigns the roles of the players in the given
// game randomly.
func AssignPlayerRoles(gameId int) error {
    db, err := utils.ConnectToDB()
    if err != nil {
        return err
    }

    results, err := db.Query(GET_PLAYERS_QUERY, gameId)
    if err != nil {
        return err
    }
    
    // Retrieve all players for this game
    var players = make([]int,0)
    numPlayers := 0
    for results.Next() {
        var playerId int
        if err := results.Scan(&playerId); err == nil {
            players = append(players, playerId)
            numPlayers = numPlayers + 1
        }
    }
    
    if numSpies, ok := numPlayersToNumSpies[numPlayers]; ok {
        // Determine which players are spies
        spies := selectSpies(players, numSpies)
        // Sets their roles in the DB
        for _, playerId := range players {
            if spies[playerId] {
                err = setPlayerRole(playerId, gameId, SPY_ROLE)
            } else {
                err = setPlayerRole(playerId, gameId, RESISTANCE_ROLE)
            }
            if err != nil {
                utils.LogMessage("Error while persisting roles " + err.Error(), utils.RESISTANCE_LOG_PATH)
                // TODO: error checking of role persistence
            }
        }
    } else {
        // TODO: error out here, not valid number of players
    }
    
    return nil
}

// selectSpies performs the random selection of spies given
// the players and number of spies.
func selectSpies(players []int, numSpies int) map[int]bool {
    // TODO: error checking that numSpies < len(players)
    var spies = make(map[int]bool)
    var randIndex int
    rand.Seed(time.Now().UnixNano())
    for len(spies) < numSpies {
        randIndex = rand.Intn(len(players))
        spies[players[randIndex]] = true
    }
    
    return spies
}

// setPlayerRole persists the players role in the DB.
func setPlayerRole(userId int, gameId int, role string) error {
    db, err := utils.ConnectToDB()
    if err != nil {
        return err
    }

    _, err = db.Query(SET_PLAYER_ROLE_QUERY, role, userId, gameId)
    if err != nil {
        return err
    }
    
    return nil
}

// GetPlayerRole returns the users role (RESISTANCE or SPY)
// for the given game.
func GetPlayerRole(userId int, gameId int) (string, error) {
    db, err := utils.ConnectToDB()
    if err != nil {
        return "", err
    }

    result, err := db.Query(PLAYER_ROLE_QUERY, userId, gameId)
    if err != nil {
        return "", err
    }
    
    // We only expect one result
    if result.Next() {
        var role []byte
        if err := result.Scan(&role); err == nil {
            utils.LogMessage("Role found: " + string(role), utils.RESISTANCE_LOG_PATH)
            if string(role) == RESISTANCE_ROLE {
                return "RESISTANCE", nil
            } else if string(role) == SPY_ROLE {
                return "SPY", nil
            }
        }
    }
    
    return "", errors.New("Something went wrong with getting the player roles")
}

// IsUserMissionLeader returns whether the given user is
// the mission leader of the current mission. Assumes that
// the game is in progress.
func IsUserMissionLeader(userId int, gameId int) (bool, error) {
    db, err := utils.ConnectToDB()
    if err != nil {
        return false, err
    }

    result, err := db.Query(MISSION_LEADER_QUERY, gameId)
    if err != nil {
        return false, err
    }
    
    // We only expect one result
    if result.Next() {
        var leaderId int
        if err := result.Scan(&leaderId); err == nil {
            return leaderId == userId, nil
        }
    }
    
    return false, errors.New("Something went wrong with retrieving the leader")
}

// StartNextMission starts the next mission for the given game.
// Assumes given game exists and has started, and has more than
// one player.
func StartNextMission(gameId int) error {

    var nextMissionNum int
    var currentLeaderId int
    var newLeaderId int
        
    db, err := utils.ConnectToDB()
    if err != nil {
        return err
    }

    // Do query for the next mission number
    missionNum, err := GetCurrentMissionNum(gameId)
    if err != nil {
        return err
    } else {
        if missionNum == 0 {
            nextMissionNum = 1
        } else {
	        result, err := db.Query(GET_CURRENT_MISSION_RESULT_QUERY, gameId)
	        if err != nil {
	            return err
	        } else {
	            var missionResult string
	            if result.Next() {
	                if err := result.Scan(&missionResult); err == nil {
	                    if missionResult != CANCELED_RESULT_STRING {
	                        nextMissionNum = missionNum + 1
	                    } else {
	                        nextMissionNum = missionNum
	                    }
	                }
	                // TODO error checking
	            }
	            // TODO error checking
	        }
	    }
    }
    
    // Do query for the new leader
    // First get the current leader
    result, err := db.Query(MISSION_LEADER_QUERY, gameId)
    if err != nil {
        return err
    }
    if result.Next() {
        if err := result.Scan(&currentLeaderId); err != nil {
            return err
        }
    } else {
        currentLeaderId = users.UNKNOWN_USER.UserId
    }
    // Then get the current players
    results, err := db.Query(GET_PLAYERS_QUERY, gameId)
    if err != nil {
        return err
    }
    var userIds = make([]int, 0)
    for results.Next() {
        var userId int
        if err := results.Scan(&userId); err == nil {
            userIds = append(userIds, userId)
        }
    }
    // If there is no current leader, this must be the first
    // mission, so the new leader is just the first player.
    if currentLeaderId == users.UNKNOWN_USER.UserId {
        newLeaderId = userIds[0]
    }
    // And select the next leader from the current players
    for i, _ := range userIds {
        if userIds[i] == currentLeaderId {
            newLeaderId = userIds[(i + 1) % len(userIds)]
            break
        }
    }
    
    // Do query for creating the mission
    _, err = db.Exec(CREATE_MISSION_QUERY, gameId, nextMissionNum, newLeaderId)
    if err != nil {
        return err
    }
    
    return nil
}

func GetCurrentMissionNum(gameId int) (int, error) {
    missionNum := 0
    
    db, err := utils.ConnectToDB()
    if err != nil {
        return -1, err
    }

    result, err := db.Query(CURRENT_MISSION_NUM_QUERY, gameId)
    if err != nil {
        return -1, err
    }
    // We only expect one result
    if result.Next() {
        if err := result.Scan(&missionNum); err != nil {
            return -1, err
        }
    }
    
    return missionNum, nil
}

// CreateTeam creates the team for the mission with the given players
// Assumes that the number of players is valid.
func CreateTeam(gameId int, playerIds []int) error {
    var missionId int

    db, err := utils.ConnectToDB()
    if err != nil {
        return err
    }
    
    result, err := db.Query(CURRENT_MISSION_ID_QUERY, gameId)
    if err != nil {
        return err
    }
    // We only expect one result
    if result.Next() {
        if err := result.Scan(&missionId); err != nil {
            return err
        }
    }
    
    for _, playerId := range playerIds {
        _, err = db.Exec(CREATE_TEAM_MEMBER_QUERY, missionId, playerId)
        if err != nil {
            return err
        }
    }
    
    return nil
}

// GetTeamSize gets the size of the current team that needs to be sent.
func GetTeamSize(gameId int) (int, error) {
    // Do query for the current mission number
    missionNum, err := GetCurrentMissionNum(gameId)
    if err != nil {
        return -1, err
    }
    
    users, err := GetPlayers(gameId)
    if err != nil {
        return -1, err
    }
    
    teamSize, ok := numPlayersOnTeam[len(users)][missionNum]
    if ok {
        return teamSize, nil
    }
    
    return -1, errors.New("Could not find how many players should be on this team")
}

// AddTeamVote adds a vote for the team of the current mission
// under the given user id
func AddTeamVote(gameId int, userId int, vote bool) error {
    var missionId int

    db, err := utils.ConnectToDB()
    if err != nil {
        return err
    }
    
    result, err := db.Query(CURRENT_MISSION_ID_QUERY, gameId)
    if err != nil {
        return err
    }
    // We only expect one result
    if result.Next() {
        if err := result.Scan(&missionId); err != nil {
            return err
        }
    }
    
    _, err = db.Exec(ADD_VOTE_QUERY, missionId, userId, vote)
    if err != nil {
        return err
    }
    
    return nil
}

// CheckMissionVotes checks that the current mission's votes are all in.
// If it is, it will also return whether the mission was approved in the
// format (missionApproved, allVotesIn, error)
func CheckMissionVotes(gameId int) (bool, bool, error) {
    var missionId int
    var allVotesIn bool
    var missionApproved bool

    db, err := utils.ConnectToDB()
    if err != nil {
        return false, false, err
    }
    
    // Query for current mission id
    result, err := db.Query(CURRENT_MISSION_ID_QUERY, gameId)
    if err != nil {
        return false, false, err
    }
    // We only expect one result
    if result.Next() {
        if err := result.Scan(&missionId); err != nil {
            return false, false, err
        }
    }
    
    // Query to see if all votes are in
    result, err = db.Query(ALL_VOTES_IN_QUERY, missionId, missionId)
    if err != nil {
        return false, false, err
    }
    // We only expect one result
    if result.Next() {
        if err := result.Scan(&allVotesIn); err != nil {
            return false, false, err
        }
    }
    
    // Query to see whether the mission got approved
    result, err = db.Query(TEAM_APPROVED_QUERY, missionId)
    if err != nil {
        return false, false, err
    }
    // We only expect one result
    if result.Next() {
        if err := result.Scan(&missionApproved); err != nil {
            return false, false, err
        }
    }
    
    return missionApproved, allVotesIn, nil
}

// SetMissionResult sets the given result as the result for
// the current mission
func SetMissionResult(gameId int, result string) error {
    if result != RESISTANCE_RESULT_STRING && result != SPY_RESULT_STRING && result != CANCELED_RESULT_STRING {
        return errors.New("Invalid result string supplied to SetMissionResult")
    }

    var missionId int

    db, err := utils.ConnectToDB()
    if err != nil {
        return err
    }
    
    // Query for current mission id
    queryResult, err := db.Query(CURRENT_MISSION_ID_QUERY, gameId)
    if err != nil {
        return err
    }
    // We only expect one result
    if queryResult.Next() {
        if err := queryResult.Scan(&missionId); err != nil {
            return err
        }
    }
    
    utils.LogMessage(result, utils.RESISTANCE_LOG_PATH)
    utils.LogMessage(strconv.Itoa(missionId), utils.RESISTANCE_LOG_PATH)
    
    _, err = db.Exec(SET_MISSION_RESULT_QUERY, result, missionId)
    if err != nil {
        return err
    }
    
    return nil
}