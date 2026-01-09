package domain

import (
	"fmt"

	"github.com/alelopezbcn/thecampaign/internal/domain/cards"
	"github.com/alelopezbcn/thecampaign/internal/domain/ports"
)

func warriorsCards() (warriors []ports.Card) {
	warriors = make([]ports.Card, 0, 15)
	for i := 1; i < 6; i++ {
		k := cards.NewKnightCard(fmt.Sprintf("k%d", i))
		warriors = append(warriors, k)

		a := cards.NewArcherCard(fmt.Sprintf("a%d", i))
		warriors = append(warriors, a)

		m := cards.NewMageCard(fmt.Sprintf("m%d", i))
		warriors = append(warriors, m)
	}

	return warriors
}

func otherButWarriorsCards() (other []ports.Card) {
	d := cards.NewDragonCard("d")

	other = []ports.Card{
		d,
		cards.NewSpecialPowerCard("s1"),
		cards.NewSpecialPowerCard("s2"),
		cards.NewSpecialPowerCard("s3"),
		cards.NewSpyCard("s"),
		cards.NewThiefCard("t"),
		cards.NewCatapultCard("c"),
	}

	for i := 1; i < 10; i++ {
		other = append(other, cards.NewSwordCard(fmt.Sprintf("e%d", i), i))
		other = append(other, cards.NewArrowCard(fmt.Sprintf("f%d", i), i))
		other = append(other, cards.NewPoisonCard(fmt.Sprintf("p%d", i), i))
		other = append(other, cards.NewGoldCard(fmt.Sprintf("g%d", i), i))
		if i == 5 || i == 7 {
			other = append(other, cards.NewGoldCard(fmt.Sprintf("g%d", i), i))
		}
	}

	return shuffle(other)
}
