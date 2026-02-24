package gamestatus

import "github.com/alelopezbcn/thecampaign/internal/domain/cards"

// CastleInput is a pre-extracted snapshot of a castle's state.
type CastleInput struct {
	IsConstructed      bool
	IsProtected        bool
	ResourceCardsCount int
	Value              int
}

// FieldInput is a pre-extracted snapshot of a player's field.
type FieldInput struct {
	Warriors  []cards.Warrior
	HasArcher bool
	HasKnight bool
	HasMage   bool
	HasDragon bool
}

// ViewerInput is a pre-extracted snapshot of the viewing player.
type ViewerInput struct {
	Name       string
	Idx        int
	Hand       []cards.Card
	Field      FieldInput
	Castle     CastleInput
	CanBuyWith func(resource cards.Resource) bool
}

// OpponentInput is a pre-extracted snapshot of an opponent.
type OpponentInput struct {
	Name           string
	CardsInHand    int
	Field          FieldInput
	Castle         CastleInput
	IsAlly         bool
	IsEliminated   bool
	IsDisconnected bool
}
