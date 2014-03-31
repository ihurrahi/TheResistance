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

var (
	DUMMY_PLAYER = NewPlayer(nil, nil)
)

type Player struct {
	game        *Game
	User        *users.User
	Role        string
	connections int
}

func (player *Player) GetGame() *Game {
	return player.game
}

func (player *Player) setGame(game *Game) {
	player.game = game
}

func (player *Player) AddConnection() {
	player.connections += 1
}

func (player *Player) RemoveConnection() {
	player.connections -= 1
}

func (player *Player) GetConnections() int {
	return player.connections
}

func (player *Player) IsValid() bool {
	return player.User != nil && player.GetGame() != nil
}

func NewPlayer(currentGame *Game, user *users.User) *Player {
	newPlayer := new(Player)
	newPlayer.setGame(currentGame)
	newPlayer.User = user
	newPlayer.Role = ROLE_UNINITIALIZED

	return newPlayer
}
