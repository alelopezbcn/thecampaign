package domain

type CardToBeDiscardedObserver interface {
	OnCardToBeDiscarded(player *Player, card Card)
}

type CardMovedToPileObserver interface {
	OnCardMovedToPile(player *Player, card Card)
}

type WarriorDeadObserver interface {
	OnWarriorDead(player *Player, card Warrior)
}

type CastleCompletionObserver interface {
	OnCastleCompletion(p *Player)
}

type MessageObserver interface {
	OnMessage(msg string)
}
