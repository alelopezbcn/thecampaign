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
		newSpyCard("s"),
		// newSpyCard("sp2"),
		// newSpyCard("sp3"),
		// newSpyCard("sp4"),
		// newSpyCard("sp5"),
		// newSpyCard("sp6"),
		// newSpyCard("sp7"),
		// newSpyCard("sp8"),
		// newSpyCard("sp9"),
		// newSpyCard("sp10"),
		// newSpyCard("sp11"),
		// newSpyCard("sp12"),
		// newSpyCard("sp13"),
		// newSpyCard("sp14"),
		newThiefCard("t"),
		// newThiefCard("t2"),
		// newThiefCard("t3"),
		// newThiefCard("t4"),
		// newThiefCard("t5"),
		// newThiefCard("t6"),
		// newThiefCard("t7"),
		// newThiefCard("t8"),
		// newThiefCard("t9"),
		// newThiefCard("t10"),
		// newThiefCard("t11"),
		// newThiefCard("t12"),
		// newThiefCard("t13"),
		// newThiefCard("t14"),

		newCatapultCard("c"),
	}

	for i := 1; i < 10; i++ {
		cards = append(cards, newGoldCard(fmt.Sprintf("g%d", i), i))
		// cards = append(cards, newGoldCard(fmt.Sprintf("gB%d", i), i))
		cards = append(cards, newSwordCard(fmt.Sprintf("e%d", i), i))
		// cards = append(cards, newSwordCard(fmt.Sprintf("eB%d", i), i))
		cards = append(cards, newArrowCard(fmt.Sprintf("f%d", i), i))
		// cards = append(cards, newArrowCard(fmt.Sprintf("arB%d", i), i))
		cards = append(cards, newPoisonCard(fmt.Sprintf("p%d", i), i))
		// cards = append(cards, newPoisonCard(fmt.Sprintf("poB%d", i), i))
	}

	return shuffle(cards)
}
