package game

const (
	OUTCOME_NONE = iota
	OUTCOME_PASS = iota
	OUTCOME_FAIL = iota
)

type TeamMember struct {
	Player  Player
	Outcome int
}
