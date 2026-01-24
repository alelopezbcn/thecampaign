package gamestatus

import (
	"fmt"

	"github.com/alelopezbcn/thecampaign/internal/domain/ports"
)

type Castle struct {
	IsConstructed bool
	ResourceCards int
	Value         int
}

func newCastle(c ports.Castle) Castle {
	return Castle{
		IsConstructed: c.IsConstructed(),
		ResourceCards: c.ResourceCards(),
		Value:         c.Value(),
	}
}

func (c Castle) String() string {
	if !c.IsConstructed {
		return "Castle not constructed"
	}
	return fmt.Sprintf("Castle: resources: %d, value: %d", c.ResourceCards, c.Value)
}
