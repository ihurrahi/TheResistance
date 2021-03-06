package game

import (
	"errors"
	"math/rand"
	"resistance/users"
	"resistance/utils"
	"strconv"
	"time"
)

const (
	STATUS_LOBBY       = "L"
	STATUS_IN_PROGRESS = "P"
	STATUS_DONE        = "D"
)

type Game struct {
	GameId     int
	Title      string
	Host       *users.User
	GameStatus string
	Missions   []*Mission
	Players    []*Player
	Persister  GamePersistor
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

func NewGame(gameTitle string, hostId string, persister GamePersistor) *Game {
	newGame := new(Game)
	newGame.GameId = -1
	newGame.Title = gameTitle
	userId, err := strconv.Atoi(hostId)
	if err == nil {
		newGame.Host = users.LookupUserById(userId)
	}
	newGame.GameStatus = STATUS_LOBBY
	newGame.Persister = persister

	err = persister.PersistGame(newGame)
	if err != nil {
		utils.LogMessage(err.Error(), utils.RESISTANCE_LOG_PATH)
	}

	return newGame
}

func (game *Game) GetUsers() []*users.User {
	var users = make([]*users.User, 0)
	for _, player := range game.Players {
		if player.IsValid() && player.GetConnections() > 0 {
			users = append(users, player.User)
		}
	}
	return users
}

func (game *Game) getPlayer(userId int) *Player {
	for _, player := range game.Players {
		if player.User.UserId == userId {
			return player
		}
	}
	return DUMMY_PLAYER
}

// AddPlayer adds the given user as a player to the game.
func (game *Game) AddPlayer(user *users.User) {
	newPlayer := game.getPlayer(user.UserId)

	if newPlayer.IsValid() {
		newPlayer.AddConnection()
	} else {
		newPlayer := NewPlayer(game, user)
		newPlayer.AddConnection()
		game.Players = append(game.Players, newPlayer)
	}
}

// PlayerDisconnect handles when a player disconnects.
// The number of connections on a player indicate
// how many players of that user is connected. When one
// disconnects, we need to keep track of that.
func (game *Game) PlayerDisconnect(user *users.User) {
	player := game.getPlayer(user.UserId)

	if player.IsValid() {
		player.RemoveConnection()
	}
}

// IsGameOver determines whether the game is over by looking at all
// the mission results. Also returns a string of who won if the game was over.
func (game *Game) IsGameOver() (bool, string) {
	resistanceWins := 0
	spyWins := 0
	for _, mission := range game.Missions {
		if mission.Winner == WINNER_RESISTANCE {
			resistanceWins += 1
		} else if mission.Winner == WINNER_SPY {
			spyWins += 1
		}
	}
	if resistanceWins >= 3 {
		return true, "Resistance"
	} else if spyWins >= 3 {
		return true, "Spy"
	}
	return false, ""
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

	// Validate the number of players
	var numPlayers = len(game.Players)
	if numPlayers < 5 || numPlayers > 10 {
		return errors.New("Resistance does not support " + strconv.Itoa(numPlayers) + " players")
	}

	// Validate all players have at least one connection open
	if game.GameStatus == STATUS_IN_PROGRESS {
		for _, player := range game.Players {
			if player.GetConnections() <= 0 {
				return errors.New("Not all players are connected")
			}
		}
	}

	return nil
}

// StartGame starts the game by:
// 1. setting the status to IN_PROGRESS
// 2. setting up the player roles
func (game *Game) StartGame() error {
	// Purge all players with no connections
	var newPlayers = make([]*Player, 0)
	for _, player := range game.Players {
		if player.GetConnections() > 0 {
			newPlayers = append(newPlayers, player)
		}
	}
	game.Players = newPlayers

	if err := game.Validate(); err != nil {
		return err
	}
	game.GameStatus = STATUS_IN_PROGRESS
	game.assignPlayerRoles()

	err := game.Persister.PersistGame(game)
	if err != nil {
		utils.LogMessage(err.Error(), utils.RESISTANCE_LOG_PATH)
	}

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

// EndGame ends the game by setting the status to be done.
func (game *Game) EndGame() {
	game.GameStatus = STATUS_DONE

	err := game.Persister.PersistGame(game)
	if err != nil {
		utils.LogMessage(err.Error(), utils.RESISTANCE_LOG_PATH)
	}
}

// GetNextLeader gets the next leader in line to lead the next mission.
func (game *Game) GetNextLeader(currentLeader *users.User) *users.User {
	var nextLeader *users.User

	// The very first leader. Just pick someone at random.
	if currentLeader == nil {
		rand.Seed(time.Now().UnixNano())
		randIndex := rand.Intn(len(game.Players))
		nextLeader = game.Players[randIndex].User
	} else {
		for index, leader := range game.Players {
			if leader.User.UserId == currentLeader.UserId {
				// We have found our leader, get the next one in line.
				nextIndex := (index + 1) % len(game.Players)
				nextLeader = game.Players[nextIndex].User
			}
		}
		if nextLeader == nil {
			// This should never happen!
			utils.LogMessage("Could not find the next leader after "+currentLeader.Username, utils.RESISTANCE_LOG_PATH)
			nextLeader = game.Players[0].User
		}
	}
	return nextLeader
}

// GetMissionInfo gets all the mission information from this game to be
// displayed to the user.
func (game *Game) GetMissionInfo() []map[string]interface{} {
	missionInfo := make([]map[string]interface{}, len(game.Missions))
	for index, mission := range game.Missions {
		missionInfo[index] = mission.GetMissionInfo()
	}
	return missionInfo
}

// IsPlayer determines whether the given user is a part of that game.
func (game *Game) IsPlayer(unknownUser *users.User) bool {
	for _, player := range game.Players {
		if player.User.UserId == unknownUser.UserId {
			return true
		}
	}
	return false
}
