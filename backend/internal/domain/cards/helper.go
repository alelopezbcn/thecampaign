package cards

import (
	"fmt"
)

type Dealer interface {
	WarriorsCards(playerCount int) (warriors []Card)
	OtherCards(playerCount int) (other []Card)
}

type dealer struct{}

func NewDealer() *dealer {
	return &dealer{}
}

func (d *dealer) WarriorsCards(playerCount int) (warriors []Card) {
	totalWarriors := 15
	warriorsPerType := 5

	if playerCount > 2 && playerCount <= 4 {
		totalWarriors = 21
		warriorsPerType = 7
	} else if playerCount > 4 {
		totalWarriors = 27
		warriorsPerType = 9
	}

	warriors = make([]Card, 0, totalWarriors)

	for i := 1; i <= warriorsPerType; i++ {
		k := NewKnight(fmt.Sprintf("k%d", i))
		warriors = append(warriors, k)

		a := NewArcher(fmt.Sprintf("a%d", i))
		warriors = append(warriors, a)

		m := NewMage(fmt.Sprintf("m%d", i))
		warriors = append(warriors, m)
	}

	return warriors
}

func (d *dealer) OtherCards(playerCount int) (other []Card) {
	other = []Card{
		NewDragon("dr1"),
		NewSpecialPower("s1"),
		NewSpecialPower("s2"),
		NewSpecialPower("s3"),
		NewSpecialPower("s4"),
		NewSpy("spy1"),
		NewThief("t1"),
		NewSabotage("sab1"),
		NewDesertion("des1"),
		NewCatapultCard("c1"),
		NewCatapultCard("c2"),
		NewFortress("fw1"),
		NewHarpoon("ha1"),
		NewBloodRain("br1"),
		NewResurrection("res1"),
		NewAmbush("amb1"),
		NewAmbush("amb2"),
	}

	if playerCount > 3 {
		other = append(other, NewDragon("dr2"))
		other = append(other, NewSpecialPower("s5"))
		other = append(other, NewSpecialPower("s6"))
		other = append(other, NewSpy("spy2"))
		other = append(other, NewThief("t2"))
		other = append(other, NewSabotage("sab2"))
		other = append(other, NewDesertion("des2"))
		other = append(other, NewCatapultCard("c3"))
		other = append(other, NewCatapultCard("c4"))
		other = append(other, NewFortress("fw2"))
		other = append(other, NewHarpoon("ha2"))
		other = append(other, NewBloodRain("br2"))
		other = append(other, NewResurrection("res2"))
		other = append(other, NewAmbush("amb3"))
		other = append(other, NewAmbush("amb4"))
	}

	for i := 1; i < 10; i++ {
		other = append(other, NewSword(fmt.Sprintf("e%d", i), i))
		other = append(other, NewArrow(fmt.Sprintf("f%d", i), i))
		other = append(other, NewPoison(fmt.Sprintf("p%d", i), i))
		other = append(other, NewGold(fmt.Sprintf("g%d", i), i))
		if i == 5 || i == 7 {
			other = append(other, NewGold(fmt.Sprintf("gr%d", i), i))
		}
		if playerCount > 3 {
			other = append(other, NewSword(fmt.Sprintf("er%d", i), i))
			other = append(other, NewArrow(fmt.Sprintf("fr%d", i), i))
			other = append(other, NewPoison(fmt.Sprintf("pr%d", i), i))
			other = append(other, NewGold(fmt.Sprintf("grr%d", i), i))
		}
	}

	return append(other, d.customCards()...)
}

func (d *dealer) customCards() []Card {
	return []Card{
		NewDesertion("custom1"),
		NewDesertion("custom2"),
		NewDesertion("custom3"),
		NewDesertion("custom4"),
		NewDesertion("custom5"),
		NewDesertion("custom6"),
		NewDesertion("custom7"),
		NewDesertion("custom8"),
		NewDesertion("custom9"),
		NewDesertion("custom10"),
	}
}
