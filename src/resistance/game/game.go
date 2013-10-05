package game

import (
    "strconv"
    "net/http"
    "resistance/utils"
    "resistance/users"
)

const (
    GAME_STATUS_LOBBY = "LOBBY"
    GAME_STATUS_IN_PROGRESS = "IN_PROGRESS"
    GAME_STATUS_DONE = "DONE"
    TITLE_KEY = "title"
    HOST_ID_KEY = "host"
    CREATE_GAME_QUERY = "insert into games (`title`, `host_id`, `status`) values (?, ?, \"" + GAME_STATUS_LOBBY + "\")"
    ADD_PLAYER_QUERY = "insert into players (`game_id`, `user_id`) values (?, ?)"
    GET_PLAYERS_QUERY = "select user_id from players where game_id = ?"
)

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