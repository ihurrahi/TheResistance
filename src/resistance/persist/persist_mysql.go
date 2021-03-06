package persist

import (
	"database/sql"
	"errors"
	"resistance/game"
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
	MISSIONS_RESULT_COLUMN      = "winner"
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
		" SET " +
		GAMES_TITLE_COLUMN + " = ?, " +
		GAMES_HOST_COLUMN + " = ?, " +
		GAMES_STATUS_COLUMN + " = ? " +
		" WHERE " + GAMES_ID_COLUMN + " = ?"
	PLAYER_PERSIST_QUERY = "INSERT INTO " + PLAYERS_TABLE +
		" (" + PLAYERS_GAME_ID_COLUMN + "," +
		PLAYERS_USER_ID_COLUMN + "," +
		PLAYERS_ROLE_COLUMN + ") " +
		" VALUES (?, ?, ?) " +
		" ON DUPLICATE KEY UPDATE " +
		PLAYERS_ROLE_COLUMN + " = VALUES(" + PLAYERS_ROLE_COLUMN + ")"
	MISSION_CREATE_QUERY = "INSERT INTO " + MISSIONS_TABLE +
		" (" + MISSIONS_GAME_ID_COLUMN + "," +
		MISSIONS_MISSION_NUM_COLUMN + "," +
		MISSIONS_LEADER_ID_COLUMN + "," +
		MISSIONS_RESULT_COLUMN + ") " +
		" VALUES (?, ?, ?, ?)"
	MISSION_PERSIST_QUERY = "INSERT INTO " + MISSIONS_TABLE +
		" (" + MISSIONS_ID_COLUMN + "," +
		MISSIONS_GAME_ID_COLUMN + "," +
		MISSIONS_MISSION_NUM_COLUMN + "," +
		MISSIONS_LEADER_ID_COLUMN + "," +
		MISSIONS_RESULT_COLUMN + ") " +
		" VALUES (?, ?, ?, ?, ?) " +
		" ON DUPLICATE KEY UPDATE " +
		MISSIONS_RESULT_COLUMN + " = VALUES(" + MISSIONS_RESULT_COLUMN + ")"
	TEAM_PERSIST_QUERY = "INSERT INTO " + TEAMS_TABLE +
		" (" + TEAMS_MISSION_ID_COLUMN + "," +
		TEAMS_USER_ID_COLUMN + "," +
		TEAMS_OUTCOME_COLUMN + ") " +
		" VALUES (?, ?, ?) " +
		" ON DUPLICATE KEY UPDATE " +
		TEAMS_OUTCOME_COLUMN + " = VALUES(" + TEAMS_OUTCOME_COLUMN + ")"
	VOTE_PERSIST_QUERY = "INSERT INTO " + VOTES_TABLE +
		" (" + VOTES_MISSION_ID_COLUMN + "," +
		VOTES_USER_ID_COLUMN + "," +
		VOTES_VOTE_COLUMN + ") " +
		" VALUES (?, ?, ?) " +
		" ON DUPLICATE KEY UPDATE " +
		VOTES_VOTE_COLUMN + " = VALUES(" + VOTES_VOTE_COLUMN + ")"
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
	MISSION_READ_QUERY = "SELECT " +
		MISSIONS_TABLE + "." + MISSIONS_ID_COLUMN + "," +
		MISSIONS_TABLE + "." + MISSIONS_MISSION_NUM_COLUMN + "," +
		users.USERS_TABLE + "." + users.USERS_ID_COLUMN + "," +
		users.USERS_TABLE + "." + users.USERS_USERNAME_COLUMN + "," +
		MISSIONS_TABLE + "." + MISSIONS_RESULT_COLUMN +
		" FROM " + MISSIONS_TABLE + " LEFT JOIN " + users.USERS_TABLE + " ON " +
		users.USERS_TABLE + "." + users.USERS_ID_COLUMN + " = " + MISSIONS_TABLE + "." + MISSIONS_LEADER_ID_COLUMN +
		" WHERE " + MISSIONS_GAME_ID_COLUMN + " = ?"
	VOTE_READ_QUERY = "SELECT " +
		VOTES_TABLE + "." + VOTES_USER_ID_COLUMN + "," +
		VOTES_TABLE + "." + VOTES_VOTE_COLUMN +
		" FROM " + VOTES_TABLE +
		" WHERE " + VOTES_MISSION_ID_COLUMN + " = ?"
	TEAM_READ_QUERY = "SELECT " +
		TEAMS_TABLE + "." + TEAMS_USER_ID_COLUMN + "," +
		TEAMS_TABLE + "." + TEAMS_OUTCOME_COLUMN +
		" FROM " + TEAMS_TABLE +
		" WHERE " + TEAMS_MISSION_ID_COLUMN + " = ?"
	GAME_STATUS_FILTER = "SELECT " +
		GAMES_ID_COLUMN +
		" FROM " + GAMES_TABLE +
		" WHERE " + GAMES_STATUS_COLUMN + " = ?"
)

type Persister struct {
	gamesCache map[int]*game.Game
	db         *sql.DB
}

func NewPersister() *Persister {
	// Initialize in memory cache
	gamesCache := make(map[int]*game.Game)

	// Initialize database. Will panic if this fails.
	db := utils.ConnectToDB()

	return &Persister{gamesCache, db}
}

func (persister *Persister) persistPlayer(currentPlayer *game.Player) error {
	if currentPlayer != nil {
		utils.LogMessage("Persisting a player...", utils.RESISTANCE_LOG_PATH)
		_, err := persister.db.Exec(PLAYER_PERSIST_QUERY,
			currentPlayer.GetGame().GameId,
			currentPlayer.User.UserId,
			currentPlayer.Role)
		if err != nil {
			return err
		}
	}

	return nil
}

func (persister *Persister) PersistMission(currentMission *game.Mission) error {
	if currentMission != nil {
		utils.LogMessage("Persisting a mission...", utils.RESISTANCE_LOG_PATH)
		// Persist the actual mission
		if currentMission.MissionId <= 0 {
			result, err := persister.db.Exec(MISSION_CREATE_QUERY,
				currentMission.GetGame().GameId,
				currentMission.MissionNum,
				currentMission.Leader.UserId,
				currentMission.Winner)
			if err == nil {
				newMissionId, err := result.LastInsertId()
				if err == nil {
					currentMission.MissionId = int(newMissionId)
				}
			}
		} else {
			_, err := persister.db.Exec(MISSION_PERSIST_QUERY,
				currentMission.MissionId,
				currentMission.GetGame().GameId,
				currentMission.MissionNum,
				currentMission.Leader.UserId,
				currentMission.Winner)
			if err != nil {
				return err
			}
		}

		// Persist the team that went on this mission. Stop on error.
		err := persister.persistTeam(currentMission)
		if err != nil {
			return err
		}

		// Persist the votes that were cast for this mission . Stop on error.
		err = persister.persistVotes(currentMission)
		if err != nil {
			return err
		}
	}

	return nil
}

func (persister *Persister) persistTeam(currentMission *game.Mission) error {
	for teamMemberId, outcome := range currentMission.Team {
		_, err := persister.db.Exec(TEAM_PERSIST_QUERY,
			currentMission.MissionId,
			teamMemberId,
			outcome)
		if err != nil {
			return err
		}
	}
	return nil
}

func (persister *Persister) persistVotes(currentMission *game.Mission) error {
	for userId, vote := range currentMission.Votes {
		_, err := persister.db.Exec(VOTE_PERSIST_QUERY,
			currentMission.MissionId,
			userId,
			vote)
		if err != nil {
			return err
		}
	}
	return nil
}

func (persister *Persister) PersistGame(currentGame *game.Game) error {
	if currentGame != nil {
		utils.LogMessage("Persisting a game...", utils.RESISTANCE_LOG_PATH)
		// Persist the game itself
		var err error
		if currentGame.GameId <= 0 {
			result, err := persister.db.Exec(GAME_CREATE_QUERY,
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
			_, err = persister.db.Exec(GAME_PERSIST_QUERY,
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
			// We want to not persist players with no connections
			if player.IsValid() {
				err = persister.persistPlayer(player)
				if err != nil {
					return err
				}
			}
		}

		// Persist all the missions. Stop on error.
		for _, mission := range currentGame.Missions {
			persister.PersistMission(mission)
			if err != nil {
				return err
			}
		}

		// Finished persisting, make sure that this game is in the cache
		persister.gamesCache[currentGame.GameId] = currentGame
	}

	return nil
}

// ReadGame returns the game corresponding to the given gameId. Tries to
// take advantage of the in memory cache before hitting the database.
// Returns nil if not found.
func (persister *Persister) ReadGame(gameId int) (game *game.Game, err error) {
	defer func() {
		if r := recover(); r != nil {
			utils.LogMessage("Could not retrieve game:"+r.(error).Error(), utils.RESISTANCE_LOG_PATH)
			err = errors.New("Could not retrieve game.")
		}
	}()

	utils.LogMessage("Reading game id "+strconv.Itoa(gameId), utils.RESISTANCE_LOG_PATH)
	utils.LogMessage("Size of gamesCache:"+strconv.Itoa(len(persister.gamesCache)), utils.RESISTANCE_LOG_PATH)

	// Don't even try if not a valid game id
	if gameId < 0 {
		return nil, errors.New("Invalid game id: " + strconv.Itoa(gameId))
	}

	retrievedGame := persister.gamesCache[gameId]

	if retrievedGame == nil {
		retrievedGame = persister.retrieveGame(gameId)

		// Update the cache
		if retrievedGame != nil {
			utils.LogMessage("Updated the cache", utils.RESISTANCE_LOG_PATH)
			persister.gamesCache[gameId] = retrievedGame
		}
	}

	return retrievedGame, nil
}

// retrieveGame hits the DB to find the game
func (persister *Persister) retrieveGame(gameId int) *game.Game {
	utils.LogMessage("Reading game id "+strconv.Itoa(gameId)+" from DB", utils.RESISTANCE_LOG_PATH)

	var retrievedGame *game.Game

	var gameTitle string
	var hostId int
	var hostUsername string
	var gameStatus string

	// Query for the game
	err := persister.db.QueryRow(GAME_READ_QUERY, gameId).Scan(&gameTitle, &hostId, &hostUsername, &gameStatus)
	if err != nil {
		utils.LogMessage("Error querying for the game:"+err.Error(), utils.RESISTANCE_LOG_PATH)
		panic(err)
	}

	// Query for the players
	playerRows, err := persister.db.Query(PLAYERS_READ_QUERY, gameId)
	if err != nil {
		utils.LogMessage("Error querying for players:"+err.Error(), utils.RESISTANCE_LOG_PATH)
		panic(err)
	}
	defer playerRows.Close()

	// Query for the missions
	missionRows, err := persister.db.Query(MISSION_READ_QUERY, gameId)
	if err != nil {
		utils.LogMessage("Error querying for missions:"+err.Error(), utils.RESISTANCE_LOG_PATH)
		panic(err)
	}
	defer missionRows.Close()

	// Build up game itself
	retrievedGame = new(game.Game)
	retrievedGame.Title = gameTitle
	retrievedGame.GameId = gameId
	retrievedGame.GameStatus = gameStatus

	hostUser := new(users.User)
	hostUser.UserId = hostId
	hostUser.Username = hostUsername

	retrievedGame.Host = hostUser

	retrievedGame.Persister = persister

	// Build up players
	for playerRows.Next() {
		var playerRole string
		var userId int
		var username string
		err := playerRows.Scan(&playerRole, &userId, &username)
		if err != nil {
			utils.LogMessage("Error parsing the player resluts:"+err.Error(), utils.RESISTANCE_LOG_PATH)
			panic(err)
		}
		user := new(users.User)
		user.UserId = userId
		user.Username = username
		newPlayer := game.NewPlayer(retrievedGame, user)
		newPlayer.Role = playerRole
		retrievedGame.Players = append(retrievedGame.Players, newPlayer)
	}

	// Build up missions
	for missionRows.Next() {
		var missionId int
		var missionNum int
		var leaderId int
		var leaderUsername string
		var missionResult string
		err := missionRows.Scan(&missionId, &missionNum, &leaderId, &leaderUsername, &missionResult)
		if err != nil {
			utils.LogMessage("Error parsing the mission results:"+err.Error(), utils.RESISTANCE_LOG_PATH)
			panic(err)
		}

		mission := new(game.Mission)
		mission.MissionId = missionId
		mission.MissionNum = missionNum
		mission.Leader = new(users.User)
		mission.Leader.UserId = leaderId
		mission.Leader.Username = leaderUsername
		mission.Winner = missionResult
		mission.Team = make(map[int]string)
		mission.Votes = make(map[int]string)

		// Query for the votes
		voteRows, err := persister.db.Query(VOTE_READ_QUERY, missionId)
		if err != nil {
			utils.LogMessage("Error querying for the votes:"+err.Error(), utils.RESISTANCE_LOG_PATH)
			panic(err)
		}
		defer voteRows.Close()

		// Build up votes
		for voteRows.Next() {
			var userId int
			var vote string
			err := voteRows.Scan(&userId, &vote)
			if err != nil {
				utils.LogMessage("Error parsing the vote results:"+err.Error(), utils.RESISTANCE_LOG_PATH)
				panic(err)
			}
			mission.Votes[userId] = vote
		}

		// Query for the team
		teamRows, err := persister.db.Query(TEAM_READ_QUERY, missionId)
		if err != nil {
			utils.LogMessage("Error querying for the team:"+err.Error(), utils.RESISTANCE_LOG_PATH)
			panic(err)
		}
		defer teamRows.Close()

		// Build up team
		for teamRows.Next() {
			var userId int
			var outcome string
			err := teamRows.Scan(&userId, &outcome)
			if err != nil {
				utils.LogMessage("Error parsing the team results:"+err.Error(), utils.RESISTANCE_LOG_PATH)
				panic(err)
			}
			mission.Team[userId] = outcome
		}

		retrievedGame.Missions = append(retrievedGame.Missions, mission)
	}

	return retrievedGame
}

// GetAllGames retrieves all games of the given game status
func (persister *Persister) GetAllGames(gameStatus string) []*game.Game {
	utils.LogMessage("getting all games from persister", utils.RESISTANCE_LOG_PATH)
	allGames := make([]*game.Game, 0)
	if gameStatus != game.STATUS_LOBBY &&
		gameStatus != game.STATUS_IN_PROGRESS &&
		gameStatus != game.STATUS_DONE {
		return allGames
	}

	result, err := persister.db.Query(GAME_STATUS_FILTER, gameStatus)
	if err == nil {
		for result.Next() {
			var gameId int
			err = result.Scan(&gameId)
			if err == nil {
				game, err := persister.ReadGame(gameId)
				if err == nil {
					allGames = append(allGames, game)
				}
			}
		}
	}

	return allGames
}
