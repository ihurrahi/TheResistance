package game

struct (
    VOTE_ALLOW = iota
    VOTE_VETO = iota
)

type Vote struct {
	Player Player
	Vote   int
}
