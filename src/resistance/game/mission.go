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

func PersistMission(currentMission *Mission) {
	// TODO: implement
}

func NewMission(currentGame *Game) *Mission {
	currentMission := currentGame.GetCurrentMission()

	var nextMissionNum int
	if currentMission == nil {
		nextMissionNum = 1
	} else {
		nextMissionNum = currentMission.MissionNum + 1
	}

	newMission := new(Mission)
	newMission.Game = currentGame
	newMission.MissionNum = nextMissionNum
	newMission.Leader = currentGame.GetNextLeader(currentMission.Leader)
	newMission.Result = RESULT_NONE
	PersistMission(newMission)

	currentGame.Missions = append(currentGame.Missions, newMission)

	return newMission
}

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

func (mission *Mission) AddOutcome(user *users.User, outcome bool) {
	if outcome {
		mission.Team[user] = OUTCOME_PASS
	} else {
		mission.Team[user] = OUTCOME_FAIL
	}
}

// IsMissionOver returns whether this mission is over by
// making sure everyone who went on the mission has given
// a PASS or FAIL.
func (mission *Mission) IsMissionOver() bool {
	for _, outcome := range mission.Team {
		if outcome == OUTCOME_NONE {
			return false
		}
	}
	return true
}

func (mission *Mission) CancelMission() {
	// TODO: implement
}
