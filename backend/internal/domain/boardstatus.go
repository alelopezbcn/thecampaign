package domain

import (
	"fmt"
	"strings"

	"github.com/alelopezbcn/thecampaign/internal/domain/ports"
)

type BoardStatus struct {
	Player                     string
	Hand                       []ports.Card
	OwnField                   []ports.Card
	OwnCastle                  ports.Castle
	EnemyField                 []ports.Card
	EnemyCastle                ports.Castle
	CardsInEnemyHand           int
	ResourceCardsInEnemyCastle int
}

func (a *BoardStatus) String() string {
	sb := strings.Builder{}

	if !a.EnemyCastle.IsConstructed() {
		sb.WriteString("Enemy's castle: not constructed \n")
	} else {
		sb.WriteString(fmt.Sprintf("Enemy's castle: %s \n", a.EnemyCastle.String()))
	}

	sb.WriteString("Enemy's field: \n")
	for _, c := range a.EnemyField {
		sb.WriteString("  - " + c.String() + "\n")
	}
	sb.WriteString("--------\n")

	sb.WriteString("Your field: \n")
	for _, c := range a.OwnField {
		sb.WriteString("  - " + c.String() + "\n")
	}

	if !a.OwnCastle.IsConstructed() {
		sb.WriteString("Your castle: not constructed \n")
	} else {
		sb.WriteString(fmt.Sprintf("Your castle: %s \n", a.OwnCastle.String()))
	}
	sb.WriteString("--------\n")

	sb.WriteString("Your hand: \n")
	for _, c := range a.Hand {
		sb.WriteString("  - " + c.String() + "\n")
	}
	sb.WriteString("\n--------")
	return sb.String()
}
