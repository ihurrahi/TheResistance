package game

import (
	"resistance/utils"
	"strconv"
)

const (
	STATUS_LOBBY       = iota
	STATUS_IN_PROGRESS = iota
	STATUS_DONE        = iota
)

type Game struct {
	GameId     int
	Title      string
	Host       Player
	GameStatus int
	Missions   []Mission
	Players    []Player
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

// GetCurrentMission returns the mission with the highest
// mission number - which should be the most current mission.
func (game *Game) GetCurrentMission() *Mission {
	highestMissionNum := 0
	var selectedMission Mission
	for _, mission := range game.Missions {
		if mission.MissionNum > highestMissionNum {
			selectedMission = mission
			highestMissionNum = mission.MissionNum
		}
	}
	return selectedMission
}

// Validate validates the game.
func (game *Game) Validate() error {
	if game.Host == nil {
		return errors.New("No host found for this game.")
	}
	var numPlayers = len(game.Players)
	if numPlayers < 5 || numPlayers > 10 {
		return errors.New("Resistance does not support " + strconv.Atoi(numPlayers) + " players")
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
	utils.PersistGame(game)
}

// AssignPlayerRoles assigns the players of the game to their
// roles. This is random and based on the number of players
// in the game, which should end to be about 1/3 being spies.
func (game *Game) assignPlayerRoles() {
	var numSpies = numPlayersToNumSpies[len(game.Players)]
	var spies = selectSpies(len(game.Players), numSpies)

	for index, singlePlayer := range game.Players {
		if spies[index] {
			singePlayer.Role = player.ROLE_SPY
		} else {
			singlePlayer.Role = player.ROLE_RESISTANCE
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
