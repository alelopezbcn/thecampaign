package cards

import (
	"fmt"
)

// DeckConfig controls how many of each card type are added to the deck.
type DeckConfig struct {
	Warriors           int // copies per warrior type (Knight, Archer, Mage)
	Dragons            int
	Harpoons           int
	SpecialPowers      int
	Spies              int
	Thieves            int
	Sabotages          int
	Catapults          int
	Fortresses         int
	Ambushes           int
	BloodRains         int
	Resurrections      int
	Desertions         int
	ConstructionCards  int
	HighValueGoldCards int // extra copies of gold 7, 8, and 9 added on top of defaults
}

type Dealer interface {
	WarriorsCards(playerCount int) (warriors []Card)
	OtherCards(playerCount int) (other []Card)
}

type dealer struct {
	cfg DeckConfig
}

func NewDealer(cfg DeckConfig) *dealer {
	return &dealer{cfg: cfg}
}

func (d *dealer) WarriorsCards(_ int) (warriors []Card) {
	warriors = make([]Card, 0, d.cfg.Warriors*3)
	for i := 1; i <= d.cfg.Warriors; i++ {
		warriors = append(warriors, NewKnight(fmt.Sprintf("k%d", i)))
		warriors = append(warriors, NewArcher(fmt.Sprintf("a%d", i)))
		warriors = append(warriors, NewMage(fmt.Sprintf("m%d", i)))
	}
	return warriors
}

func (d *dealer) OtherCards(_ int) (other []Card) {
	for i := 0; i < d.cfg.Dragons; i++ {
		other = append(other, NewDragon(fmt.Sprintf("dr%d", i+1)))
	}
	for i := 0; i < d.cfg.SpecialPowers; i++ {
		other = append(other, NewSpecialPower(fmt.Sprintf("s%d", i+1)))
	}
	for i := 0; i < d.cfg.Spies; i++ {
		other = append(other, NewSpy(fmt.Sprintf("spy%d", i+1)))
	}
	for i := 0; i < d.cfg.Thieves; i++ {
		other = append(other, NewThief(fmt.Sprintf("t%d", i+1)))
	}
	for i := 0; i < d.cfg.Sabotages; i++ {
		other = append(other, NewSabotage(fmt.Sprintf("sab%d", i+1)))
	}
	for i := 0; i < d.cfg.Desertions; i++ {
		other = append(other, NewDesertion(fmt.Sprintf("des%d", i+1)))
	}
	for i := 0; i < d.cfg.Catapults; i++ {
		other = append(other, NewCatapultCard(fmt.Sprintf("c%d", i+1)))
	}
	for i := 0; i < d.cfg.Fortresses; i++ {
		other = append(other, NewFortress(fmt.Sprintf("fw%d", i+1)))
	}
	for i := 0; i < d.cfg.Harpoons; i++ {
		other = append(other, NewHarpoon(fmt.Sprintf("ha%d", i+1)))
	}
	for i := 0; i < d.cfg.BloodRains; i++ {
		other = append(other, NewBloodRain(fmt.Sprintf("br%d", i+1)))
	}
	for i := 0; i < d.cfg.Resurrections; i++ {
		other = append(other, NewResurrection(fmt.Sprintf("res%d", i+1)))
	}
	for i := 0; i < d.cfg.Ambushes; i++ {
		other = append(other, NewAmbush(fmt.Sprintf("amb%d", i+1)))
	}

	for i := 1; i <= d.cfg.ConstructionCards; i++ {
		other = append(other, NewSword(fmt.Sprintf("econ%d", i), 1))
		other = append(other, NewArrow(fmt.Sprintf("acon%d", i), 1))
		other = append(other, NewPoison(fmt.Sprintf("pcon%d", i), 1))
		other = append(other, NewGold(fmt.Sprintf("gcon%d", i), 1))
	}
	for i := 2; i < 10; i++ {
		other = append(other, NewSword(fmt.Sprintf("e%d", i), i))
		other = append(other, NewArrow(fmt.Sprintf("f%d", i), i))
		other = append(other, NewPoison(fmt.Sprintf("p%d", i), i))
		other = append(other, NewGold(fmt.Sprintf("g%d", i), i))
		if i == 5 || i == 7 {
			other = append(other, NewGold(fmt.Sprintf("gr%d", i), i))
		}
		if i == 7 || i == 8 || i == 9 {
			for j := 0; j < d.cfg.HighValueGoldCards; j++ {
				other = append(other, NewGold(fmt.Sprintf("g%d_x%d", i, j+1), i))
			}
		}
	}

	return other
}
