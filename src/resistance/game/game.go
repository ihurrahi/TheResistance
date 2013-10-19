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
    
    CREATE_GAME_QUERY = "insert into games (`title`, `host_id`, `status`) values (?, ?, \"" + GAME_STATUS_LOBBY + "\")"
    ADD_PLAYER_QUERY = "insert into players (`game_id`, `user_id`) values (?, ?)"
    GET_PLAYERS_QUERY = "select user_id from players where game_id = ? order by join_date"
    SET_GAME_STATUS_QUERY = "update games set status = ? where game_id = ?"
    NUM_PLAYERS_QUERY = "select user_id from players where game_id = ?"
    SET_PLAYER_ROLE_QUERY = "update players set role = ? where user_id = ? and game_id = ?"
    PLAYER_ROLE_QUERY = "select role from players where user_id = ? and game_id = ?"
    MISSION_LEADER_QUERY = "select leader_id from missions where game_id = ? order by mission_num desc limit 1"
    NEXT_MISSION_NUM_QUERY = "select max(mission_num) + 1 from missions where game_id = ?"
    CREATE_MISSION_QUERY = "insert into missions (`game_id`, `mission_num`, `leader_id`) values (?, ?, ?)"
)

// numPlayersToNumSpies gives you how many spies there should be in a game
// for the given the number of players
var numPlayersToNumSpies = map[int]int{5:2, 6:2, 7:3, 8:3, 9:3, 10:4}

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

// AddPlayer adds the given user to the given game by storing the
// relevant information in the players table. This can only be done
// while the game is still in the LOBBY stage.
func AddPlayer(gameId int, userId int) error {
    db, err := utils.ConnectToDB()
    if err != nil {
        return err
    }

    _, err = db.Exec(ADD_PLAYER_QUERY, gameId, userId)
    if err != nil {
        return err
    }
    return nil
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
func StartNextMission(gameId int) error {

    var nextMissionNum int
    var currentLeaderId int
    var newLeaderId int
        
    db, err := utils.ConnectToDB()
    if err != nil {
        return err
    }

    // Do query for the next mission number
    result, err := db.Query(NEXT_MISSION_NUM_QUERY, gameId)
    if err != nil {
        return err
    }
    // We only expect one result
    if result.Next() {
        if err := result.Scan(&nextMissionNum); err != nil {
            return err
        }
    } else {
        nextMissionNum = 0
    }
    
    // Do query for the new leader
    // First get the current leader
    result, err = db.Query(MISSION_LEADER_QUERY, gameId)
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
