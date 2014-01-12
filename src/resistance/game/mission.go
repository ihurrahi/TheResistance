package game

const (
	RESULT_RESISTANCE = iota
	RESULT_SPY        = iota
	RESULT_NONE       = iota
)

type Mission struct {
	Game       Game
	MissionId  int
	MissionNum int
	Leader     Player
	Result     int
	Team       []TeamMember
	Votes      []Vote
}

// IsAllVotesCollected returns whether the voting is complete
// and all the votes for this mission are in. We need to make
// sure we have as many votes as there are players in the game.
func (mission *Mission) IsAllVotesCollected() bool {
	return len(mission.Votes) == len(mission.Game.Players)
}

// IsTeamApproved returns whether the team going on this mission
// was approved.
func (mission *Mission) IsTeamApproved() bool {
	var approvalVotes int = 0
	for _, singleVote := range mission.Votes {
		if singleVote.Vote == vote.VOTE_ALLOW {
			approvalVotes += 1
		}
	}
	return (2 * approvalVotes) > len(mission.Votes)
}

// IsMissionOver returns whether this mission is over by
// making sure everyone who went on the mission has given
// a PASS or FAIL.
func (mission *Mission) IsMissionOver() bool {
	for _, teamMember := range mission.Team {
		if teamMember.Outcome == team.OUTCOME_NONE {
			return false
		}
	}
	return true
}
