package gamestatus

import "github.com/alelopezbcn/thecampaign/internal/domain/cards"

type DiscardPile struct {
	Cards    int   `json:"cards"`
	LastCard *Card `json:"last_card,omitempty"`
}

func NewDiscardPile(count int, lastCard cards.Card) DiscardPile {
	d := DiscardPile{
		Cards: count,
	}
	if lastCard != nil {
		card := fromDomainCard(lastCard)
		d.LastCard = &card
	}
	return d
}
