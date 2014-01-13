package game

import (
	"resistance/users"
)

const (
	RESULT_RESISTANCE = iota
	RESULT_SPY        = iota
	RESULT_NONE       = iota
)

const (
	OUTCOME_NONE = iota
	OUTCOME_PASS = iota
	OUTCOME_FAIL = iota
)

const (
	VOTE_ALLOW = iota
	VOTE_VETO  = iota
)

type Mission struct {
	Game       *Game
	MissionId  int
	MissionNum int
	Leader     *users.User
	Result     int
	Team       map[*users.User]int
	Votes      map[*users.User]int
}

func NewMission(currentGame *Game) *Mission {
	currentMission := currentGame.GetCurrentMission()

	var nextMissionNum int
	if currentMission == nil {
		nextMissionNum = 1
	} else if currentMission.Result == RESULT_NONE {
		nextMissionNum = currentMission.MissionNum
	} else {
		nextMissionNum = currentMission.MissionNum + 1
	}

	newMission := new(Mission)
	newMission.Game = currentGame
	newMission.MissionNum = nextMissionNum
	newMission.Leader = currentGame.GetNextLeader(currentMission.Leader)
	newMission.Result = RESULT_NONE

	currentGame.Missions = append(currentGame.Missions, newMission)

	return newMission
}

// CreateTeam creates the team for this mission with the
// given list of users
func (mission *Mission) CreateTeam(team []*users.User) {
	for _, user := range team {
		mission.Team[user] = OUTCOME_NONE
	}
}

// AddVote adds the vote of approval for the chosen team from a given
// user to the mission.
func (mission *Mission) AddVote(user *users.User, vote bool) {
	if vote {
		mission.Votes[user] = VOTE_ALLOW
	} else {
		mission.Votes[user] = VOTE_VETO
	}
}

// IsAllVotesCollected returns whether the voting is complete
// and all the votes for this mission are in. We need to make
// sure we have as many votes as there are players in the game.
func (mission *Mission) IsAllVotesCollected() bool {
	return len(mission.Votes) == len(mission.Game.Players)
}

// IsTeamApproved returns whether the team going on this mission
// was approved. Assumes that all votes were collected
func (mission *Mission) IsTeamApproved() bool {
	var approvalVotes int = 0
	for _, singleVote := range mission.Votes {
		if singleVote == VOTE_ALLOW {
			approvalVotes += 1
		}
	}
	return (2 * approvalVotes) > len(mission.Votes)
}

// AddOutcome adds the outcome of each user who went on the mission
// to the mission.
func (mission *Mission) AddOutcome(user *users.User, outcome bool) {
	if outcome {
		mission.Team[user] = OUTCOME_PASS
	} else {
		mission.Team[user] = OUTCOME_FAIL
	}
}

// IsMissionOver returns whether this mission is over by
// making sure everyone who went on the mission has given
// a PASS or FAIL. Also returns who won if it is over
func (mission *Mission) IsMissionOver() (bool, int) {
	numFails := 0
	for _, outcome := range mission.Team {
		if outcome == OUTCOME_NONE {
			return false, RESULT_NONE
		} else if outcome == OUTCOME_FAIL {
			numFails += 1
		}
	}

	failsRequired := 1
	if mission.IsRequiresTwoFails() {
		failsRequired = 2
	}

	if numFails >= failsRequired {
		return true, RESULT_SPY
	} else {
		return true, RESULT_RESISTANCE
	}
}

// IsRequiresTwoFails returns whether this mission requires two fails to
// fail the mission.
func (mission *Mission) IsRequiresTwoFails() bool {
	return len(mission.Game.Players) >= 7 && mission.MissionNum == 4
}

// EndMission ends the mission by setting the result of the mission.
func (mission *Mission) EndMission(result int) {
	mission.Result = result
}

// GetMissionInfo constructs the mission information of this mission to
// be displayed on the frontend
func (mission *Mission) GetMissionInfo() map[string]interface{} {
	missionInfo := make(map[string]interface{})
	missionInfo["missionNum"] = mission.MissionNum
	missionInfo["missionLeader"] = mission.Leader
	missionInfo["missionResult"] = mission.Result
	return make(map[string]interface{})
}

// IsUserCurrentMissionLeader returns whether the given user is the current
// mission's leader or not.
func (mission *Mission) IsUserCurrentMissionLeader(currentUser *users.User) bool {
	return mission.Leader.UserId == currentUser.UserId
}

// GetCurrentMissionTeamSize returns the how many players need to go on this
// mission.
func (mission *Mission) GetCurrentMissionTeamSize() int {
	return numPlayersOnTeam[len(mission.Game.Players)][mission.MissionNum]
}

// IsUserOnCurrentMission returns whether the given user is going on this
// mission.
func (mission *Mission) IsUserOnCurrentMission(currentUser *users.User) bool {
	if _, ok := mission.Team[currentUser]; ok {
		return true
	}
	return false
}
