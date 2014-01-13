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

func (game *Game) AddPlayer(user *User) {
	newPlayer := NewPlayer(game, user)
	game.Players = append(game.Players, newPlayer)
}

func (game *Game) AddVote(user *User, vote bool) {
	currentMission := game.GetCurrentMission()
	currentMission.AddVote(user, vote)
}

func (game *Game) AddOutcome(user *User, outcome bool) {
	currentMission := game.GetCurrentMission()
	currentMission.AddOutcome(user, outcome)
}

func (game *Game) IsCurrentMissionOver() bool {
	currentMission := game.GetCurrentMission()
	return currentMission.IsMissionOver()
}

func (game *Game) EndMission() {
	// TODO: implement
}

func (game *Game) IsGameOver() bool {
	// TODO: implement
}

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

func (game *Game) IsUserCurrentMissionLeader(currentUser *User) bool {
	currentMission := game.GetCurrentMission()
	return currentMission.Leader.User.UserId == currentUser.UserId
}

func (game *Game) GetCurrentMissionTeamSize() int {
	currentMission := game.GetCurrentMission()
	return numPlayersOnTeam[len(game.Players)][currentMission.MissionNum]
}

func (game *Game) IsUserOnCurrentMission(currentUser *User) bool {
	currentMission := game.GetCurrentMission()
	for _, teamMember := range currentMission.Team {
		if teamMember.Player.User.UserId == currentUser.UserId {
			return true
		}
	}
	return false
}
