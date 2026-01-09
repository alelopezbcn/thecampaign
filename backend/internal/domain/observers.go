package domain

type CardToBeDiscardedObserver interface {
	OnCardToBeDiscarded(card Card)
}

type CardMovedToPileObserver interface {
	OnCardMovedToPile(card Card)
}

type WarriorDeadObserver interface {
	OnWarriorDead(card Warrior)
}

type WarriorMovedToCemeteryObserver interface {
	OnWarriorMovedToCemetery(card Warrior)
}

type CastleCompletionObserver interface {
	OnCastleCompletion(p Player)
}

type MessageObserver interface {
	OnMessage(msg string)
}
