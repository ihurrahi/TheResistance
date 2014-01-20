package game

import (
	"resistance/users"
)

const (
	ROLE_UNINITIALIZED      = iota
	ROLE_UNINITIALIZED_NAME = "None"
	ROLE_RESISTANCE         = iota
	ROLE_RESISTANCE_NAME    = "Resistance"
	ROLE_SPY                = iota
	ROLE_SPY_NAME           = "Spy"
)

type Player struct {
	Game *Game
	User *users.User
	Role int
}

func NewPlayer(currentGame *Game, user *users.User) *Player {
	newPlayer := new(Player)
	newPlayer.Game = currentGame
	newPlayer.User = user
	newPlayer.Role = ROLE_UNINITIALIZED

	return newPlayer
}
