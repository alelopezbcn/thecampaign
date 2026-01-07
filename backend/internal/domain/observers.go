package domain

type CardUsedObserver interface {
	OnCardMovedToPile(player *Player, card iCard)
}

type WarriorDeadObserver interface {
	OnWarriorDead(player *Player, card iCard)
}

type CastleCompletionObserver interface {
	OnCastleCompletion(p *Player)
}
