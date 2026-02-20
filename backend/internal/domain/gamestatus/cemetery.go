package gamestatus

import (
	"github.com/alelopezbcn/thecampaign/internal/domain/ports"
)

type Cemetery struct {
	Corps    int  `json:"corps"`
	LastCorp Card `json:"last_corp"`
}

func NewCemetery(corps int, lastCorp ports.Warrior) Cemetery {
	c := Cemetery{
		Corps: corps,
	}
	if lastCorp != nil {
		c.LastCorp = FromDomainCard(lastCorp)
	}

	return c
}
