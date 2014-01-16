package game

import (
	"database/sql"
	"resistance/users"
	"resistance/utils"
	"strconv"
)

const (
	GAMES_TABLE         = "games"
	GAMES_ID_COLUMN     = "game_id"
	GAMES_TITLE_COLUMN  = "title"
	GAMES_HOST_COLUMN   = "host_id"
	GAMES_STATUS_COLUMN = "status"
)

const (
	MISSIONS_TABLE              = "missions"
	MISSIONS_ID_COLUMN          = "mission_id"
	MISSIONS_GAME_ID_COLUMN     = "game_id"
	MISSIONS_MISSION_NUM_COLUMN = "mission_num"
	MISSIONS_LEADER_ID_COLUMN   = "leader_id"
	MISSIONS_RESULT_COLUMN      = "result"
)

const (
	PLAYERS_TABLE            = "players"
	PLAYERS_GAME_ID_COLUMN   = "game_id"
	PLAYERS_USER_ID_COLUMN   = "user_id"
	PLAYERS_ROLE_COLUMN      = "role"
	PLAYERS_JOIN_DATE_COLUMN = "join_date"
)

const (
	TEAMS_TABLE             = "teams"
	TEAMS_MISSION_ID_COLUMN = "mission_id"
	TEAMS_USER_ID_COLUMN    = "user_id"
	TEAMS_OUTCOME_COLUMN    = "outcome"
)

const (
	VOTES_TABLE             = "votes"
	VOTES_MISSION_ID_COLUMN = "mission_id"
	VOTES_USER_ID_COLUMN    = "user_id"
	VOTES_VOTE_COLUMN       = "vote"
)

const (
	GAME_CREATE_QUERY = "INSERT INTO " + GAMES_TABLE +
		" (" + GAMES_TITLE_COLUMN + "," +
		GAMES_HOST_COLUMN + "," +
		GAMES_STATUS_COLUMN + ") " +
		"VALUES (?, ?, ?)"
	GAME_PERSIST_QUERY = "UPDATE " + GAMES_TABLE +
		" (" + GAMES_TITLE_COLUMN + "," +
		GAMES_HOST_COLUMN + "," +
		GAMES_STATUS_COLUMN + ") " +
		"VALUES (?, ?, ?) WHERE " + GAMES_ID_COLUMN + " = ?"
	PLAYER_PERSIST_QUERY = "INSERT INTO " + PLAYERS_TABLE +
		" (" + PLAYERS_GAME_ID_COLUMN + "," +
		PLAYERS_USER_ID_COLUMN + "," +
		PLAYERS_ROLE_COLUMN + ") " +
		"VALUES (?, ?, ?) ON DUPLICATE KEY UPDATE"
	MISSION_PERSIST_QUERY = "INSERT INTO " + MISSIONS_TABLE +
		" (" + MISSIONS_ID_COLUMN + "," +
		MISSIONS_GAME_ID_COLUMN + "," +
		MISSIONS_MISSION_NUM_COLUMN + "," +
		MISSIONS_LEADER_ID_COLUMN + "," +
		MISSIONS_RESULT_COLUMN + ") " +
		"VALUES (?, ?, ?, ?, ?) ON DUPLICATE KEY UPDATE"
	TEAM_PERSIST_QUERY = "INSERT INTO " + TEAMS_TABLE +
		" (" + TEAMS_MISSION_ID_COLUMN + "," +
		TEAMS_USER_ID_COLUMN + "," +
		TEAMS_OUTCOME_COLUMN + ") " +
		"VALUES (?, ?, ?) ON DUPLICATE KEY UPDATE"
	VOTE_PERSIST_QUERY = "INSERT INTO " + VOTES_TABLE +
		" (" + VOTES_MISSION_ID_COLUMN + "," +
		VOTES_USER_ID_COLUMN + "," +
		VOTES_VOTE_COLUMN + ") " +
		"VALUES (?, ?, ?) ON DUPLICATE KEY UPDATE"
)

const (
	GAME_READ_QUERY = "SELECT " +
		GAMES_TABLE + "." + GAMES_TITLE_COLUMN + "," +
		users.USERS_TABLE + "." + users.USERS_ID_COLUMN + "," +
		users.USERS_TABLE + "." + users.USERS_USERNAME_COLUMN + "," +
		GAMES_TABLE + "." + GAMES_STATUS_COLUMN +
		" FROM " + GAMES_TABLE + " LEFT JOIN " + users.USERS_TABLE + " ON " +
		users.USERS_TABLE + "." + users.USERS_ID_COLUMN + " = " + GAMES_TABLE + "." + GAMES_HOST_COLUMN +
		" WHERE " + GAMES_ID_COLUMN + " = ?"
	PLAYERS_READ_QUERY = "SELECT " +
		PLAYERS_TABLE + "." + PLAYERS_ROLE_COLUMN + "," +
		users.USERS_TABLE + "." + users.USERS_ID_COLUMN + "," +
		users.USERS_TABLE + "." + users.USERS_USERNAME_COLUMN +
		" FROM " + PLAYERS_TABLE + " LEFT JOIN " + users.USERS_TABLE + " ON " +
		users.USERS_TABLE + "." + users.USERS_ID_COLUMN + " = " + PLAYERS_TABLE + "." + PLAYERS_USER_ID_COLUMN +
		" WHERE " + PLAYERS_GAME_ID_COLUMN + " = ?"
	MISSION_READ_QUERY = "SELECT * FROM " + MISSIONS_TABLE +
		" WHERE " + MISSIONS_GAME_ID_COLUMN + " = ?"
)

var gamesCache map[int]*Game
var db *sql.DB

func init() {
	// Initialize in memory cache
	gamesCache = make(map[int]*Game)

	// Initialize database. Will panic if this fails.
	db = utils.ConnectToDB()
}

func PersistPlayer(currentPlayer *Player) error {
	utils.LogMessage("Persisting a player...", utils.RESISTANCE_LOG_PATH)
	_, err := db.Exec(PLAYER_PERSIST_QUERY,
		currentPlayer.Game.GameId,
		currentPlayer.User.UserId,
		currentPlayer.Role)
	if err != nil {
		return err
	}

	return nil
}

func PersistMission(currentMission *Mission) error {
	utils.LogMessage("Persisting a mission...", utils.RESISTANCE_LOG_PATH)
	// Persist the actual mission
	_, err := db.Exec(MISSION_PERSIST_QUERY,
		currentMission.MissionId,
		currentMission.Game.GameId,
		currentMission.MissionNum,
		currentMission.Leader.UserId,
		currentMission.Result)
	if err != nil {
		return err
	}

	// Persist the team that went on this mission. Stop on error.
	err = persistTeam(currentMission)
	if err != nil {
		return err
	}

	// Persist the votes that were cast for this mission . Stop on error.
	err = persistVotes(currentMission)
	if err != nil {
		return err
	}

	return nil
}

func persistTeam(currentMission *Mission) error {
	for teamMember, outcome := range currentMission.Team {
		_, err := db.Exec(TEAM_PERSIST_QUERY,
			currentMission.MissionId,
			teamMember.UserId,
			outcome)
		if err != nil {
			return err
		}
	}
	return nil
}

func persistVotes(currentMission *Mission) error {
	for user, vote := range currentMission.Votes {
		_, err := db.Exec(VOTE_PERSIST_QUERY,
			currentMission.MissionId,
			user.UserId,
			vote)
		if err != nil {
			return err
		}
	}
	return nil
}

func PersistGame(currentGame *Game) error {
	utils.LogMessage("Persisting a game...", utils.RESISTANCE_LOG_PATH)
	// Persist the game itself
	var err error
	if currentGame.GameId == 0 {
		result, err := db.Exec(GAME_CREATE_QUERY,
			currentGame.Title,
			currentGame.Host.UserId,
			currentGame.GameStatus)
		if err == nil {
			newGameId, err := result.LastInsertId()
			if err == nil {
				currentGame.GameId = int(newGameId)
			}
		}
	} else {
		_, err = db.Exec(GAME_PERSIST_QUERY,
			currentGame.Title,
			currentGame.Host.UserId,
			currentGame.GameStatus,
			currentGame.GameId)
	}
	if err != nil {
		return err
	}

	// Persist all the players. Stop on error.
	for _, player := range currentGame.Players {
		err = PersistPlayer(player)
		if err != nil {
			return err
		}
	}

	// Persist all the missions. Stop on error.
	for _, mission := range currentGame.Missions {
		PersistMission(mission)
		if err != nil {
			return err
		}
	}

	// Finished persisting, make sure that this game is in the cache
	gamesCache[currentGame.GameId] = currentGame
	return nil
}

// ReadGame returns the game corresponding to the given gameId. Tries to
// take advantage of the in memory cache before hitting the database.
// Returns nil if not found.
func ReadGame(gameId int) *Game {
	utils.LogMessage("Reading game id "+strconv.Itoa(gameId), utils.RESISTANCE_LOG_PATH)
	utils.LogMessage("Size of gamesCache:"+strconv.Itoa(len(gamesCache)), utils.RESISTANCE_LOG_PATH)
	retrievedGame := gamesCache[gameId]

	if retrievedGame == nil {
		retrievedGame = retrieveGame(gameId)

		// Update the cache
		if retrievedGame != nil {
			utils.LogMessage("Updated the cache", utils.RESISTANCE_LOG_PATH)
			gamesCache[gameId] = retrievedGame
		}
	}

	return retrievedGame
}

// retrieveGame hits the DB to find the game
func retrieveGame(gameId int) *Game {
	utils.LogMessage("Reading game id "+strconv.Itoa(gameId)+" from DB", utils.RESISTANCE_LOG_PATH)

	var retrievedGame *Game

	var gameTitle string
	var hostId int
	var hostUsername string
	var gameStatus int

	err := db.QueryRow(GAME_READ_QUERY, gameId).Scan(&gameTitle, &hostId, &hostUsername, &gameStatus)
	if err == nil {
		retrievedGame = new(Game)
		retrievedGame.GameId = gameId
		retrievedGame.GameStatus = gameStatus

		user := new(users.User)
		user.UserId = hostId
		user.Username = hostUsername

		retrievedGame.Host = user
	}

	return retrievedGame
}
