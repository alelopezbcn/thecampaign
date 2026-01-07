package domain

import "fmt"

func warriorsCards(o WarriorDeadObserver) (warriors []iCard) {
	warriors = make([]iCard, 0, 15)
	for i := 1; i < 6; i++ {
		k := newKnightCard(fmt.Sprintf("k%d", i))
		k.AddObserver(o)
		warriors = append(warriors, k)

		a := newArcherCard(fmt.Sprintf("a%d", i))
		a.AddObserver(o)
		warriors = append(warriors, a)

		m := newMageCard(fmt.Sprintf("m%d", i))
		m.AddObserver(o)
		warriors = append(warriors, m)
	}

	return warriors
}

func otherButWarriorsCards(o WarriorDeadObserver) (cards []iCard) {
	d := newDragonCard("d")
	d.AddObserver(o)

	cards = []iCard{
		d,
		d,
		d,
		d,
		d,
		d,
		d,
		d,
		d,
		d,
		d,
		d,
		d,
		d,
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
