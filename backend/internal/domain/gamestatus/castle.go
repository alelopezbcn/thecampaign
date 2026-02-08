package gamestatus

import (
	"github.com/alelopezbcn/thecampaign/internal/domain/ports"
)

type Castle struct {
	IsConstructed bool `json:"constructed"`
	ResourceCards int  `json:"resource_cards"`
	Value         int  `json:"value"`
}

func NewCastle(c ports.Castle) Castle {
	return Castle{
		IsConstructed: c.IsConstructed(),
		ResourceCards: c.ResourceCardsCount(),
		Value:         c.Value(),
	}
}
