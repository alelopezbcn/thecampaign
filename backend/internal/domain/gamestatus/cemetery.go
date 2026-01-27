package gamestatus

import (
	"github.com/alelopezbcn/thecampaign/internal/domain/ports"
)

type Cemetery struct {
	Corps    int  `json:"corps"`
	LastCorp Card `json:"last_corp"`
}

func newCemetery(cementery ports.Cemetery) Cemetery {
	c := Cemetery{
		Corps: cementery.Count(),
	}
	if w := cementery.GetLast(); w != nil {
		c.LastCorp = FromDomainCard(w)
	}

	return c
}
