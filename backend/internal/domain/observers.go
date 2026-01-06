package domain

type CardUsedObserver interface {
	OnCardUsed(player *Player, card iCard)
}

type WarriorDeadObserver interface {
	OnWarriorDead(player *Player, card iCard)
}

type GameEndedObserver interface {
	OnGameEnded(reason string)
}
