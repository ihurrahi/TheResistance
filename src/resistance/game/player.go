package game

import (
	"resistance/users"
)

const (
	ROLE_UNINITIALIZED      = ""
	ROLE_UNINITIALIZED_NAME = "None"
	ROLE_RESISTANCE         = "R"
	ROLE_RESISTANCE_NAME    = "Resistance"
	ROLE_SPY                = "S"
	ROLE_SPY_NAME           = "Spy"
)

type Player struct {
	game *Game
	User *users.User
	Role string
}

func (player *Player) GetGame() *Game {
	return player.game
}

func (player *Player) setGame(game *Game) {
	player.game = game
}

func NewPlayer(currentGame *Game, user *users.User) *Player {
	newPlayer := new(Player)
	newPlayer.setGame(currentGame)
	newPlayer.User = user
	newPlayer.Role = ROLE_UNINITIALIZED

	return newPlayer
}
