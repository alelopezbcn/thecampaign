package gamestatus

import (
	"github.com/alelopezbcn/thecampaign/internal/domain/ports"
)

type DiscardPile struct {
	Cards    int  `json:"cards"`
	LastCard Card `json:"last_card"`
}

func NewDiscardPile(cards int, lastCard ports.Card) DiscardPile {
	d := DiscardPile{
		Cards: cards,
	}
	if lastCard != nil {
		d.LastCard = FromDomainCard(lastCard)
	}

	return d
}
