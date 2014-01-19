package game

type GamePersistor interface {
	PersistGame(*Game) error
	PersistMission(*Mission) error
}
