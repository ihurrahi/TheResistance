package game

import (
	"resistance/users"
)

const (
	ROLE_RESISTANCE = iota
	ROLE_SPY        = iota
)

type Player struct {
	PlayerId int
	User     User
	Role     int
}
