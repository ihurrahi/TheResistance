package game

import (
    "net/http"
    "resistance/utils"
)

const (
    GAME_STATUS_LOBBY = "LOBBY"
    GAME_STATUS_IN_PROGRESS = "IN_PROGRESS"
    GAME_STATUS_DONE = "DONE"
    TITLE_KEY = "title"
    HOST_ID_KEY = "host"
    CREATE_GAME_QUERY = "insert into games (`title`, `host_id`, `status`) values (?, ?, \"" + GAME_STATUS_LOBBY + "\")"
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
// relevant information in the players table. 
func AddPlayer(userId int, gameId int) {

}

// DeletePlayer deletes the given user from the given game by
// removing the user/game pair from the players table. This can
// only be done while the game is still in the LOBBY stage.
func DeletePlayer(userId int, gameId int) {

}