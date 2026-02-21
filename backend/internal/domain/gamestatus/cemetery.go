package gamestatus

import "github.com/alelopezbcn/thecampaign/internal/domain/cards"

type Cemetery struct {
	Corps    int  `json:"corps"`
	LastCorp Card `json:"last_corp"`
}

func NewCemetery(corps int, lastCorp cards.Warrior) Cemetery {
	c := Cemetery{
		Corps: corps,
	}
	if lastCorp != nil {
		c.LastCorp = FromDomainCard(lastCorp)
	}

	return c
}
