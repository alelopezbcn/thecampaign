package gamestatus

import (
	"github.com/alelopezbcn/thecampaign/internal/domain/ports"
)

type Cemetery struct {
	Corps    int  `json:"corps"`
	LastCorp Card `json:"last_corp"`
}

func newCemetery(corps int, lastCorp ports.Warrior) Cemetery {
	return Cemetery{
		Corps:    corps,
		LastCorp: FromDomainCard(lastCorp),
	}
}
