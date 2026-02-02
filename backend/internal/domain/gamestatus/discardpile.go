package gamestatus

import (
	"github.com/alelopezbcn/thecampaign/internal/domain/ports"
)

type DiscardPile struct {
	Cards    int  `json:"cards"`
	LastCard Card `json:"last_card"`
}

func NewDiscardPile(discardPile ports.DiscardPile) DiscardPile {
	d := DiscardPile{
		Cards: discardPile.Count(),
	}
	if c := discardPile.GetLast(); c != nil {
		d.LastCard = FromDomainCard(c)
	}

	return d
}
