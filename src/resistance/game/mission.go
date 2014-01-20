package game

import (
	"resistance/users"
	"resistance/utils"
)

const (
	WINNER_NONE            = ""
	WINNER_NONE_NAME       = ""
	WINNER_RESISTANCE      = "R"
	WINNER_RESISTANCE_NAME = "Success"
	WINNER_SPY             = "S"
	WINNER_SPY_NAME        = "Fail"
)

const (
	OUTCOME_NONE = ""
	OUTCOME_PASS = "P"
	OUTCOME_FAIL = "F"
)

const (
	VOTE_ALLOW = "A"
	VOTE_VETO  = "V"
)

type Mission struct {
	Game       *Game
	MissionId  int
	MissionNum int
	Leader     *users.User
	Winner     string
	Team       map[int]string
	Votes      map[int]string
}

func NewMission(currentGame *Game) *Mission {
	currentMission := currentGame.GetCurrentMission()

	var nextMissionNum int
	var currentLeader *users.User
	if currentMission == nil {
		nextMissionNum = 1
		currentLeader = nil
	} else if currentMission.Winner == WINNER_NONE {
		nextMissionNum = currentMission.MissionNum
		currentLeader = currentMission.Leader
	} else {
		nextMissionNum = currentMission.MissionNum + 1
		currentLeader = currentMission.Leader
	}

	newMission := new(Mission)
	newMission.Game = currentGame
	newMission.MissionNum = nextMissionNum
	newMission.Leader = currentGame.GetNextLeader(currentLeader)
	newMission.Winner = WINNER_NONE
	newMission.Team = make(map[int]string)
	newMission.Votes = make(map[int]string)

	currentGame.Missions = append(currentGame.Missions, newMission)

	err := currentGame.Persister.PersistMission(newMission)
	if err != nil {
		utils.LogMessage(err.Error(), utils.RESISTANCE_LOG_PATH)
	}

	return newMission
}

// CreateTeam creates the team for this mission with the
// given list of users
func (mission *Mission) CreateTeam(team []*users.User) {
	for _, user := range team {
		mission.Team[user.UserId] = OUTCOME_NONE
	}

	err := mission.Game.Persister.PersistMission(mission)
	if err != nil {
		utils.LogMessage(err.Error(), utils.RESISTANCE_LOG_PATH)
	}
}

// AddVote adds the vote of approval for the chosen team from a given
// user to the mission.
func (mission *Mission) AddVote(user *users.User, vote bool) {
	if vote {
		mission.Votes[user.UserId] = VOTE_ALLOW
	} else {
		mission.Votes[user.UserId] = VOTE_VETO
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
		mission.Team[user.UserId] = OUTCOME_PASS
	} else {
		mission.Team[user.UserId] = OUTCOME_FAIL
	}
}

// IsMissionOver returns whether this mission is over by
// making sure everyone who went on the mission has given
// a PASS or FAIL. Also returns who won if it is over
func (mission *Mission) IsMissionOver() (bool, string) {
	numFails := 0
	for _, outcome := range mission.Team {
		if outcome == OUTCOME_NONE {
			return false, WINNER_NONE
		} else if outcome == OUTCOME_FAIL {
			numFails += 1
		}
	}

	failsRequired := 1
	if mission.IsRequiresTwoFails() {
		failsRequired = 2
	}

	if numFails >= failsRequired {
		return true, WINNER_SPY
	} else {
		return true, WINNER_RESISTANCE
	}
}

// IsRequiresTwoFails returns whether this mission requires two fails to
// fail the mission.
func (mission *Mission) IsRequiresTwoFails() bool {
	return len(mission.Game.Players) >= 7 && mission.MissionNum == 4
}

// EndMission ends the mission by setting the result of the mission.
func (mission *Mission) EndMission(result string) {
	mission.Winner = result

	err := mission.Game.Persister.PersistMission(mission)
	if err != nil {
		utils.LogMessage(err.Error(), utils.RESISTANCE_LOG_PATH)
	}
}

// GetMissionInfo constructs the mission information of this mission to
// be displayed on the frontend
func (mission *Mission) GetMissionInfo() map[string]interface{} {
	missionInfo := make(map[string]interface{})
	missionInfo["missionNum"] = mission.MissionNum
	missionInfo["missionLeader"] = mission.Leader
	switch {
	case mission.Winner == WINNER_NONE:
		missionInfo["missionResult"] = WINNER_NONE_NAME
	case mission.Winner == WINNER_RESISTANCE:
		missionInfo["missionResult"] = WINNER_RESISTANCE_NAME
	case mission.Winner == WINNER_SPY:
		missionInfo["missionResult"] = WINNER_SPY_NAME
	}
	return missionInfo
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
	if _, ok := mission.Team[currentUser.UserId]; ok {
		return true
	}
	return false
}
