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
	PlayerId int
	Game     *Game
	User     *User
	Role     int
}

func NewPlayer(currentGame *Game, user *User) *Player {
	newPlayer := new(Player)
	newPlayer.Game = currentGame
	newPlayer.User = user
	newPlayer.Role = ROLE_UNINITIALIZED

	utils.PersistPlayer(newPlayer)

	return newPlayer
}
