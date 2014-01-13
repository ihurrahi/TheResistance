package persist

import (
	"resistance/game"
	"resistance/users"
	"resistance/utils"
)

var gamesCache map[int]*game.Game

func init() {
	gamesCache = make(map[int]*game.Game)
}

func PersistPlayer(currentPlayer *game.Player) error {
	// TODO: implement
	return nil
}

func PersistMission(currentMission *game.Mission) error {
	// TODO: implement
	return nil
}

func PersistGame(currentGame *game.Game) error {
	// TODO: implement
	_, _ = utils.ConnectToDB()
	return nil
}

func IsValidGame(gameId string, requestUser *users.User) (map[string]string, error) {
	// TODO: implement
	return make(map[string]string), nil
}

// ReadGame returns the game corresponding to the given gameId. Tries to
// take advantage of the in memory cache before hitting the database.
// Returns nil if not found.
func ReadGame(gameId int) *game.Game {
	retrievedGame := gamesCache[gameId]

	if retrievedGame == nil {
		retrievedGame = retrieveGame(gameId)
	}

	return retrievedGame
}

func retrieveGame(gameId int) *game.Game {
	return nil
}
