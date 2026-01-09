package domain

import "fmt"

func warriorsCards() (warriors []Card) {
	warriors = make([]Card, 0, 15)
	for i := 1; i < 6; i++ {
		k := newKnightCard(fmt.Sprintf("k%d", i))
		warriors = append(warriors, k)

		a := newArcherCard(fmt.Sprintf("a%d", i))
		warriors = append(warriors, a)

		m := newMageCard(fmt.Sprintf("m%d", i))
		warriors = append(warriors, m)
	}

	return warriors
}

func otherButWarriorsCards() (cards []Card) {
	d := newDragonCard("d")

	cards = []Card{
		d,
		newSpecialPowerCard("s1"),
		newSpecialPowerCard("s2"),
		newSpecialPowerCard("s3"),
		newSpyCard("s"),
		newThiefCard("t"),
		newCatapultCard("c"),
	}

	for i := 1; i < 10; i++ {
		cards = append(cards, newSwordCard(fmt.Sprintf("e%d", i), i))
		cards = append(cards, newArrowCard(fmt.Sprintf("f%d", i), i))
		cards = append(cards, newPoisonCard(fmt.Sprintf("p%d", i), i))
		cards = append(cards, newGoldCard(fmt.Sprintf("g%d", i), i))
		if i == 5 || i == 7 {
			cards = append(cards, newGoldCard(fmt.Sprintf("g%d", i), i))
		}
	}

	return shuffle(cards)
}
