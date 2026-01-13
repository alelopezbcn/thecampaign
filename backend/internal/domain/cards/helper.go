package cards

import (
	"fmt"

	"github.com/alelopezbcn/thecampaign/internal/domain/ports"
)

type dealer struct {
}

func NewDealer() ports.Dealer {
	return &dealer{}
}

func (d *dealer) WarriorsCards() (warriors []ports.Card) {
	warriors = make([]ports.Card, 0, 15)
	for i := 1; i < 6; i++ {
		k := NewKnight(fmt.Sprintf("k%d", i))
		warriors = append(warriors, k)

		a := NewArcher(fmt.Sprintf("a%d", i))
		warriors = append(warriors, a)

		m := NewMage(fmt.Sprintf("m%d", i))
		warriors = append(warriors, m)
	}

	return warriors
}

func (d *dealer) OtherCards() (other []ports.Card) {
	dr := NewDragon("d")

	other = []ports.Card{
		dr,
		NewSpecialPower("s1"),
		NewSpecialPower("s2"),
		NewSpecialPower("s3"),
		NewSpy("s"),
		NewThief("t"),
		NewCatapultCard("c"),
	}

	for i := 1; i < 10; i++ {
		other = append(other, NewSword(fmt.Sprintf("e%d", i), i))
		other = append(other, NewArrow(fmt.Sprintf("f%d", i), i))
		other = append(other, NewPoison(fmt.Sprintf("p%d", i), i))
		other = append(other, NewGold(fmt.Sprintf("g%d", i), i))
		if i == 5 || i == 7 {
			other = append(other, NewGold(fmt.Sprintf("g%d", i), i))
		}
	}

	return other
}
