package game

import (
	"errors"
	"math/rand"
	"resistance/users"
	"strconv"
	"time"
)

const (
	STATUS_LOBBY       = iota
	STATUS_IN_PROGRESS = iota
	STATUS_DONE        = iota
)

type Game struct {
	GameId     int
	Title      string
	Host       *Player
	GameStatus int
	Missions   []*Mission
	Players    []*Player
}

// numPlayersToNumSpies gives you how many spies there should be in a game
// for the given the number of players
var numPlayersToNumSpies = map[int]int{
	5:  2,
	6:  2,
	7:  3,
	8:  3,
	9:  3,
	10: 4}

// numPlayersOnTeam gives you how many players should be on a team
// given the total number of players and the mission number
var numPlayersOnTeam = map[int]map[int]int{
	5:  {1: 2, 2: 3, 3: 2, 4: 3, 5: 3},
	6:  {1: 2, 2: 3, 3: 4, 4: 3, 5: 4},
	7:  {1: 2, 2: 3, 3: 3, 4: 4, 5: 4},
	8:  {1: 3, 2: 4, 3: 4, 4: 5, 5: 5},
	9:  {1: 3, 2: 4, 3: 4, 4: 5, 5: 5},
	10: {1: 3, 2: 4, 3: 4, 4: 5, 5: 5}}

func CreateGame(gameTitle string, hostId string) (int64, error) {
	// TODO: implement
	return 0, nil
}

func IsValidGame(gameId string, requestUser *users.User) (map[string]string, error) {
	// TODO: implement
	return make(map[string]string), nil
}

func PersistGame(currentGame *Game) {
	// TODO: implement
}

func ReadGame(gameId int) *Game {
	// TODO: implement
	return nil
}

func (game *Game) AddPlayer(user *users.User) {
	newPlayer := NewPlayer(game, user)
	game.Players = append(game.Players, newPlayer)
}

func (game *Game) CreateTeam(team []*users.User) {
	// TODO: implement
}

func (game *Game) EndMission() {
	// TODO: implement
}

func (game *Game) IsGameOver() (bool, string) {
	// TODO: implement
	return false, "no one"
}

// GetCurrentMission returns the most current mission. This should
// be the mission with the highest mission number. This should also
// the last one in the Missions array.
func (game *Game) GetCurrentMission() *Mission {
	var currentMission *Mission
	if len(game.Missions) == 0 {
		currentMission = nil
	} else {
		currentMission = game.Missions[len(game.Missions)-1]
	}
	return currentMission
}

// Validate validates the game.
func (game *Game) Validate() error {
	if game.Host == nil {
		return errors.New("No host found for this game.")
	}
	var numPlayers = len(game.Players)
	if numPlayers < 5 || numPlayers > 10 {
		return errors.New("Resistance does not support " + strconv.Itoa(numPlayers) + " players")
	}

	return nil
}

// StartGame starts the game by:
// 1. setting the status to IN_PROGRESS
// 2. setting up the player roles
// 3. persisting the game to the DB
func (game *Game) StartGame() error {
	if err := game.Validate(); err != nil {
		return err
	}
	game.GameStatus = STATUS_IN_PROGRESS
	game.assignPlayerRoles()
	PersistGame(game)

	return nil
}

// AssignPlayerRoles assigns the players of the game to their
// roles. This is random and based on the number of players
// in the game, which should end to be about 1/3 being spies.
func (game *Game) assignPlayerRoles() {
	var numSpies = numPlayersToNumSpies[len(game.Players)]
	var spies = selectSpies(len(game.Players), numSpies)

	for index, singlePlayer := range game.Players {
		if spies[index] {
			singlePlayer.Role = ROLE_SPY
		} else {
			singlePlayer.Role = ROLE_RESISTANCE
		}
	}
}

// selectSpies performs the random selection of spies given
// the number of players and number of spies.
func selectSpies(numPlayers int, numSpies int) map[int]bool {
	var spies = make(map[int]bool)
	var randIndex int
	rand.Seed(time.Now().UnixNano())
	for len(spies) < numSpies {
		randIndex = rand.Intn(numPlayers)
		spies[randIndex] = true
	}

	return spies
}

func (game *Game) IsUserCurrentMissionLeader(currentUser *users.User) bool {
	// TODO: move to mission struct
	currentMission := game.GetCurrentMission()
	return currentMission.Leader.UserId == currentUser.UserId
}

func (game *Game) GetCurrentMissionTeamSize() int {
	// TODO: move to mission struct
	currentMission := game.GetCurrentMission()
	return numPlayersOnTeam[len(game.Players)][currentMission.MissionNum]
}

func (game *Game) IsUserOnCurrentMission(currentUser *users.User) bool {
	// TODO: move to mission struct
	currentMission := game.GetCurrentMission()
	if _, ok := currentMission.Team[currentUser]; ok {
		return true
	}
	return false
}

func (game *Game) GetNextLeader(currentLeader *users.User) *users.User {
	// TODO: implement
	return nil
}

func (game *Game) GetMissionInfo() map[string]interface{} {
	// TODO: implement
	return make(map[string]interface{})
}
