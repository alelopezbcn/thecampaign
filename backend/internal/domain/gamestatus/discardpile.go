package gamestatus

import "github.com/alelopezbcn/thecampaign/internal/domain/cards"

type DiscardPile struct {
	Cards    int  `json:"cards"`
	LastCard Card `json:"last_card"`
}

func NewDiscardPile(cards int, lastCard cards.Card) DiscardPile {
	d := DiscardPile{
		Cards: cards,
	}
	if lastCard != nil {
		d.LastCard = fromDomainCard(lastCard)
	}

	return d
}
