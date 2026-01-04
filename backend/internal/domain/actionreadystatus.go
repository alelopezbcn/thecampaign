package domain

import (
	"fmt"
	"strings"
)

type ActionReadyStatus struct {
	Player      string
	Hand        []Card
	OwnField    []Card
	OwnCastle   *Castle
	EnemyField  []Card
	EnemyCastle *Castle
}

func (a *ActionReadyStatus) String() string {
	sb := strings.Builder{}

	if a.EnemyCastle == nil {
		sb.WriteString("Enemy's castle: not constructed \n")
	} else {
		sb.WriteString(fmt.Sprintf("Enemy's castle: %s \n", a.EnemyCastle.String()))
	}

	sb.WriteString("Enemy's field: \n")
	for _, c := range a.EnemyField {
		sb.WriteString("  - " + c.String() + "\n")
	}
	sb.WriteString("---\n")

	sb.WriteString("Your field: \n")
	for _, c := range a.OwnField {
		sb.WriteString("  - " + c.String() + "\n")
	}

	if a.OwnCastle == nil {
		sb.WriteString("Your castle: not constructed \n")
	} else {
		sb.WriteString(fmt.Sprintf("Your castle: %s \n", a.OwnCastle.String()))
	}
	sb.WriteString("---\n")

	sb.WriteString("Your hand: \n")
	for _, c := range a.Hand {
		sb.WriteString("  - " + c.String() + "\n")
	}
	sb.WriteString("\n---\n")
	return sb.String()
}
