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
		newSpecialMoveCard("s1"),
		newSpecialMoveCard("s2"),
		newSpecialMoveCard("s3"),
		newSpecialMoveCard("s4"),
		newSpyCard("sp"),
		newThiefCard("t"),
		newCatapultCard("c"),
	}

	for i := 1; i < 10; i++ {
		cards = append(cards, newGoldCard(fmt.Sprintf("goA%d", i), i))
		cards = append(cards, newGoldCard(fmt.Sprintf("goB%d", i), i))
		cards = append(cards, newSwordCard(fmt.Sprintf("swA%d", i), i))
		cards = append(cards, newSwordCard(fmt.Sprintf("swB%d", i), i))
		cards = append(cards, newArrowCard(fmt.Sprintf("arA%d", i), i))
		cards = append(cards, newArrowCard(fmt.Sprintf("arB%d", i), i))
		cards = append(cards, newPoisonCard(fmt.Sprintf("poA%d", i), i))
		cards = append(cards, newPoisonCard(fmt.Sprintf("poB%d", i), i))
	}

	return shuffle(cards)
}
