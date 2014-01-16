package game

import (
	"resistance/users"
)

const (
	ROLE_UNINITIALIZED = iota
	ROLE_RESISTANCE    = iota
	ROLE_SPY           = iota
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
