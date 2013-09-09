package game

import (
    "log"
    "net/http"    
    "database/sql"
)

const (
    GAME_STATUS_LOBBY = "LOBBY"
    GAME_STATUS_IN_PROGRESS = "IN_PROGRESS"
    GAME_STATUS_DONE = "DONE"
    TITLE_KEY = "title"
    HOST_ID_KEY = "host"
    CREATE_GAME_QUERY = "insert into games (`title`, `host_id`, `status`) values (?, ?, \"" + GAME_STATUS_LOBBY + "\")"
)

var gameLogger *log.Logger
var db *sql.DB

func CreateGame(request *http.Request) (int64, error) {
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

func Initialize(logger *log.Logger, database *sql.DB) {
    gameLogger = logger
    db = database
}