package domain

import (
	"testing"

	"github.com/alelopezbcn/thecampaign/internal/domain/cards"
	"github.com/alelopezbcn/thecampaign/internal/domain/ports"
	"github.com/stretchr/testify/assert"
)

func TestAttacks(t *testing.T) {
	t.Run("Knight attacks Archer causing double damage", func(t *testing.T) {
		dmgAmnt := 4
		k := cards.NewKnight("k1")
		a := cards.NewArcher("a1")
		sword := cards.NewSword("s1", dmgAmnt)

		p1 := newPlayerWithCards("Player1",
			[]ports.Card{sword},
			[]ports.Warrior{k},
		)
		p2 := newPlayerWithCards("Player2",
			[]ports.Card{},
			[]ports.Warrior{a},
		)
		g := &Game{
			Players: []ports.Player{p1, p2},
		}

		err := g.Attack(p1.Name(), a.GetID(), sword.GetID())
		assert.NoError(t, err)
		assert.Equal(t, cards.WarriorMaxHealth-dmgAmnt*2, a.Health())
		assert.NotContains(t, p1.Hand().ShowCards(), sword)
	})
	t.Run("Knight attacks Mage causing normal damage", func(t *testing.T) {
		dmgAmnt := 4
		k := cards.NewKnight("k1")
		m := cards.NewMage("m1")
		sword := cards.NewSword("s1", dmgAmnt)

		p1 := newPlayerWithCards("Player1",
			[]ports.Card{sword},
			[]ports.Warrior{k},
		)
		p2 := newPlayerWithCards("Player2",
			[]ports.Card{},
			[]ports.Warrior{m},
		)
		g := &Game{
			Players: []ports.Player{p1, p2},
		}

		err := g.Attack(p1.Name(), m.GetID(), sword.GetID())
		assert.NoError(t, err)
		assert.Equal(t, cards.WarriorMaxHealth-dmgAmnt*1, m.Health())
		assert.NotContains(t, p1.Hand().ShowCards(), sword)
	})
	t.Run("Knight attacks Knight causing normal damage", func(t *testing.T) {
		dmgAmnt := 4
		k := cards.NewKnight("k1")
		k2 := cards.NewKnight("k2")
		sword := cards.NewSword("s1", dmgAmnt)

		p1 := newPlayerWithCards("Player1",
			[]ports.Card{sword},
			[]ports.Warrior{k},
		)
		p2 := newPlayerWithCards("Player2",
			[]ports.Card{},
			[]ports.Warrior{k2},
		)
		g := &Game{
			Players: []ports.Player{p1, p2},
		}

		err := g.Attack(p1.Name(), k2.GetID(), sword.GetID())
		assert.NoError(t, err)
		assert.Equal(t, cards.WarriorMaxHealth-dmgAmnt*1, k2.Health())
		assert.NotContains(t, p1.Hand().ShowCards(), sword)
	})
	t.Run("Knight attacks Dragon causing normal damage", func(t *testing.T) {
		dmgAmnt := 4
		k := cards.NewKnight("k1")
		d := cards.NewDragon("d1")
		sword := cards.NewSword("s1", dmgAmnt)

		p1 := newPlayerWithCards("Player1",
			[]ports.Card{sword},
			[]ports.Warrior{k},
		)
		p2 := newPlayerWithCards("Player2",
			[]ports.Card{},
			[]ports.Warrior{d},
		)
		g := &Game{
			Players: []ports.Player{p1, p2},
		}

		err := g.Attack(p1.Name(), d.GetID(), sword.GetID())
		assert.NoError(t, err)
		assert.Equal(t, cards.DragonMaxHealth-dmgAmnt*1, d.Health())
		assert.NotContains(t, p1.Hand().ShowCards(), sword)
	})
	t.Run("Knight cant attack with wrong weapon", func(t *testing.T) {
		dmgAmnt := 4
		k := cards.NewKnight("k1")
		a := cards.NewArcher("a1")
		poison := cards.NewPoison("s1", dmgAmnt)

		p1 := newPlayerWithCards("Player1",
			[]ports.Card{poison},
			[]ports.Warrior{k},
		)
		p2 := newPlayerWithCards("Player2",
			[]ports.Card{},
			[]ports.Warrior{a},
		)
		g := &Game{
			Players: []ports.Player{p1, p2},
		}

		err := g.Attack(p1.Name(), a.GetID(), poison.GetID())
		assert.Error(t, err)
		assert.Equal(t, cards.WarriorMaxHealth, a.Health())
		assert.Contains(t, p1.Hand().ShowCards(), poison)
	})

	t.Run("Archer attacks Mage causing double damage", func(t *testing.T) {
		dmgAmnt := 4
		attacker := cards.NewArcher("a1")
		target := cards.NewMage("a1")
		weapon := cards.NewArrow("s1", dmgAmnt)

		p1 := newPlayerWithCards("Player1",
			[]ports.Card{weapon},
			[]ports.Warrior{attacker},
		)
		p2 := newPlayerWithCards("Player2",
			[]ports.Card{},
			[]ports.Warrior{target},
		)
		g := &Game{
			Players: []ports.Player{p1, p2},
		}

		err := g.Attack(p1.Name(), target.GetID(), weapon.GetID())
		assert.NoError(t, err)
		assert.Equal(t, cards.WarriorMaxHealth-dmgAmnt*2, target.Health())
		assert.NotContains(t, p1.Hand().ShowCards(), weapon)
	})
	t.Run("Archer attacks Knight causing normal damage", func(t *testing.T) {
		dmgAmnt := 4
		attacker := cards.NewArcher("a1")
		target := cards.NewKnight("k1")
		weapon := cards.NewArrow("s1", dmgAmnt)

		p1 := newPlayerWithCards("Player1",
			[]ports.Card{weapon},
			[]ports.Warrior{attacker},
		)
		p2 := newPlayerWithCards("Player2",
			[]ports.Card{},
			[]ports.Warrior{target},
		)
		g := &Game{
			Players: []ports.Player{p1, p2},
		}

		err := g.Attack(p1.Name(), target.GetID(), weapon.GetID())
		assert.NoError(t, err)
		assert.Equal(t, cards.WarriorMaxHealth-dmgAmnt*1, target.Health())
		assert.NotContains(t, p1.Hand().ShowCards(), weapon)
	})
	t.Run("Archer attacks Archer causing normal damage", func(t *testing.T) {
		dmgAmnt := 4
		attacker := cards.NewArcher("a1")
		target := cards.NewArcher("a2")
		weapon := cards.NewArrow("s1", dmgAmnt)

		p1 := newPlayerWithCards("Player1",
			[]ports.Card{weapon},
			[]ports.Warrior{attacker},
		)
		p2 := newPlayerWithCards("Player2",
			[]ports.Card{},
			[]ports.Warrior{target},
		)
		g := &Game{
			Players: []ports.Player{p1, p2},
		}

		err := g.Attack(p1.Name(), target.GetID(), weapon.GetID())
		assert.NoError(t, err)
		assert.Equal(t, cards.WarriorMaxHealth-dmgAmnt*1, target.Health())
		assert.NotContains(t, p1.Hand().ShowCards(), weapon)
	})
	t.Run("Archer attacks Dragon causing normal damage", func(t *testing.T) {
		dmgAmnt := 4
		attacker := cards.NewArcher("a1")
		target := cards.NewDragon("d1")
		weapon := cards.NewArrow("s1", dmgAmnt)

		p1 := newPlayerWithCards("Player1",
			[]ports.Card{weapon},
			[]ports.Warrior{attacker},
		)
		p2 := newPlayerWithCards("Player2",
			[]ports.Card{},
			[]ports.Warrior{target},
		)
		g := &Game{
			Players: []ports.Player{p1, p2},
		}

		err := g.Attack(p1.Name(), target.GetID(), weapon.GetID())
		assert.NoError(t, err)
		assert.Equal(t, cards.DragonMaxHealth-dmgAmnt*1, target.Health())
		assert.NotContains(t, p1.Hand().ShowCards(), weapon)
	})
	t.Run("Archer cant attack with wrong weapon", func(t *testing.T) {
		dmgAmnt := 4
		attacker := cards.NewArcher("a1")
		target := cards.NewMage("m1")
		weapon := cards.NewSword("s1", dmgAmnt)

		p1 := newPlayerWithCards("Player1",
			[]ports.Card{weapon},
			[]ports.Warrior{attacker},
		)
		p2 := newPlayerWithCards("Player2",
			[]ports.Card{},
			[]ports.Warrior{target},
		)
		g := &Game{
			Players: []ports.Player{p1, p2},
		}

		err := g.Attack(p1.Name(), target.GetID(), weapon.GetID())
		assert.Error(t, err)
		assert.Equal(t, cards.WarriorMaxHealth, target.Health())
		assert.Contains(t, p1.Hand().ShowCards(), weapon)
	})

	t.Run("Mage attacks Knight causing double damage", func(t *testing.T) {
		dmgAmnt := 4
		attacker := cards.NewMage("m1")
		target := cards.NewKnight("k1")
		weapon := cards.NewPoison("s1", dmgAmnt)

		p1 := newPlayerWithCards("Player1",
			[]ports.Card{weapon},
			[]ports.Warrior{attacker},
		)
		p2 := newPlayerWithCards("Player2",
			[]ports.Card{},
			[]ports.Warrior{target},
		)
		g := &Game{
			Players: []ports.Player{p1, p2},
		}

		err := g.Attack(p1.Name(), target.GetID(), weapon.GetID())
		assert.NoError(t, err)
		assert.Equal(t, cards.WarriorMaxHealth-dmgAmnt*2, target.Health())
		assert.NotContains(t, p1.Hand().ShowCards(), weapon)
	})
	t.Run("Mage attacks Archer causing normal damage", func(t *testing.T) {
		dmgAmnt := 4
		attacker := cards.NewMage("m1")
		target := cards.NewArcher("a1")
		weapon := cards.NewPoison("s1", dmgAmnt)

		p1 := newPlayerWithCards("Player1",
			[]ports.Card{weapon},
			[]ports.Warrior{attacker},
		)
		p2 := newPlayerWithCards("Player2",
			[]ports.Card{},
			[]ports.Warrior{target},
		)
		g := &Game{
			Players: []ports.Player{p1, p2},
		}

		err := g.Attack(p1.Name(), target.GetID(), weapon.GetID())
		assert.NoError(t, err)
		assert.Equal(t, cards.WarriorMaxHealth-dmgAmnt*1, target.Health())
		assert.NotContains(t, p1.Hand().ShowCards(), weapon)
	})
	t.Run("Mage attacks Mage causing normal damage", func(t *testing.T) {
		dmgAmnt := 4
		attacker := cards.NewMage("m1")
		target := cards.NewMage("m2")
		weapon := cards.NewPoison("s1", dmgAmnt)

		p1 := newPlayerWithCards("Player1",
			[]ports.Card{weapon},
			[]ports.Warrior{attacker},
		)
		p2 := newPlayerWithCards("Player2",
			[]ports.Card{},
			[]ports.Warrior{target},
		)
		g := &Game{
			Players: []ports.Player{p1, p2},
		}

		err := g.Attack(p1.Name(), target.GetID(), weapon.GetID())
		assert.NoError(t, err)
		assert.Equal(t, cards.WarriorMaxHealth-dmgAmnt*1, target.Health())
		assert.NotContains(t, p1.Hand().ShowCards(), weapon)
	})
	t.Run("Mage attacks Dragon causing normal damage", func(t *testing.T) {
		dmgAmnt := 4
		attacker := cards.NewMage("m1")
		target := cards.NewDragon("d1")
		weapon := cards.NewPoison("s1", dmgAmnt)

		p1 := newPlayerWithCards("Player1",
			[]ports.Card{weapon},
			[]ports.Warrior{attacker},
		)
		p2 := newPlayerWithCards("Player2",
			[]ports.Card{},
			[]ports.Warrior{target},
		)
		g := &Game{
			Players: []ports.Player{p1, p2},
		}

		err := g.Attack(p1.Name(), target.GetID(), weapon.GetID())
		assert.NoError(t, err)
		assert.Equal(t, cards.DragonMaxHealth-dmgAmnt*1, target.Health())
		assert.NotContains(t, p1.Hand().ShowCards(), weapon)
	})
	t.Run("Mage cant attack with wrong weapon", func(t *testing.T) {
		dmgAmnt := 4
		attacker := cards.NewMage("m1")
		target := cards.NewKnight("k1")
		weapon := cards.NewArrow("s1", dmgAmnt)

		p1 := newPlayerWithCards("Player1",
			[]ports.Card{weapon},
			[]ports.Warrior{attacker},
		)
		p2 := newPlayerWithCards("Player2",
			[]ports.Card{},
			[]ports.Warrior{target},
		)
		g := &Game{
			Players: []ports.Player{p1, p2},
		}

		err := g.Attack(p1.Name(), target.GetID(), weapon.GetID())
		assert.Error(t, err)
		assert.Equal(t, cards.WarriorMaxHealth, target.Health())
		assert.Contains(t, p1.Hand().ShowCards(), weapon)
	})

	t.Run("Player cant attack with non existing cards", func(t *testing.T) {
		k := cards.NewKnight("k1")
		a := cards.NewArcher("a1")
		sword := cards.NewSword("s1", 4)

		p1 := newPlayerWithCards("Player1",
			[]ports.Card{sword},
			[]ports.Warrior{k},
		)
		p2 := newPlayerWithCards("Player2",
			[]ports.Card{},
			[]ports.Warrior{a},
		)
		g := &Game{
			Players: []ports.Player{p1, p2},
		}

		err := g.Attack(p1.Name(), "non-existent-target", sword.GetID())
		assert.Error(t, err)

		err = g.Attack(p1.Name(), a.GetID(), "non-existent-weapon")
		assert.Error(t, err)
	})

	t.Run("Dragon attacks Knight with Sword causing normal damage", func(t *testing.T) {
		dmgAmnt := 6
		attacker := cards.NewDragon("d1")
		target := cards.NewKnight("k1")
		weapon := cards.NewSword("s1", dmgAmnt)

		p1 := newPlayerWithCards("Player1",
			[]ports.Card{weapon},
			[]ports.Warrior{attacker},
		)
		p2 := newPlayerWithCards("Player2",
			[]ports.Card{},
			[]ports.Warrior{target},
		)
		g := &Game{
			Players: []ports.Player{p1, p2},
		}

		err := g.Attack(p1.Name(), target.GetID(), weapon.GetID())
		assert.NoError(t, err)
		assert.Equal(t, cards.WarriorMaxHealth-dmgAmnt*1, target.Health())
		assert.NotContains(t, p1.Hand().ShowCards(), weapon)
	})
	t.Run("Dragon attacks Knight with Arrow causing normal damage", func(t *testing.T) {
		dmgAmnt := 6
		attacker := cards.NewDragon("d1")
		target := cards.NewKnight("k1")
		weapon := cards.NewArrow("s1", dmgAmnt)

		p1 := newPlayerWithCards("Player1",
			[]ports.Card{weapon},
			[]ports.Warrior{attacker},
		)
		p2 := newPlayerWithCards("Player2",
			[]ports.Card{},
			[]ports.Warrior{target},
		)
		g := &Game{
			Players: []ports.Player{p1, p2},
		}

		err := g.Attack(p1.Name(), target.GetID(), weapon.GetID())
		assert.NoError(t, err)
		assert.Equal(t, cards.WarriorMaxHealth-dmgAmnt*1, target.Health())
		assert.NotContains(t, p1.Hand().ShowCards(), weapon)
	})
	t.Run("Dragon attacks Knight with Poison causing double damage", func(t *testing.T) {
		dmgAmnt := 6
		attacker := cards.NewDragon("d1")
		target := cards.NewKnight("k1")
		weapon := cards.NewPoison("s1", dmgAmnt)

		p1 := newPlayerWithCards("Player1",
			[]ports.Card{weapon},
			[]ports.Warrior{attacker},
		)
		p2 := newPlayerWithCards("Player2",
			[]ports.Card{},
			[]ports.Warrior{target},
		)
		g := &Game{
			Players: []ports.Player{p1, p2},
		}

		err := g.Attack(p1.Name(), target.GetID(), weapon.GetID())
		assert.NoError(t, err)
		assert.Equal(t, cards.WarriorMaxHealth-dmgAmnt*2, target.Health())
		assert.NotContains(t, p1.Hand().ShowCards(), weapon)
	})
	t.Run("Dragon attacks Archer with Sword causing double damage", func(t *testing.T) {
		dmgAmnt := 6
		attacker := cards.NewDragon("d1")
		target := cards.NewArcher("a1")
		weapon := cards.NewSword("s1", dmgAmnt)

		p1 := newPlayerWithCards("Player1",
			[]ports.Card{weapon},
			[]ports.Warrior{attacker},
		)
		p2 := newPlayerWithCards("Player2",
			[]ports.Card{},
			[]ports.Warrior{target},
		)
		g := &Game{
			Players: []ports.Player{p1, p2},
		}

		err := g.Attack(p1.Name(), target.GetID(), weapon.GetID())
		assert.NoError(t, err)
		assert.Equal(t, cards.WarriorMaxHealth-dmgAmnt*2, target.Health())
		assert.NotContains(t, p1.Hand().ShowCards(), weapon)
	})
	t.Run("Dragon attacks Archer with Arrow causing normal damage", func(t *testing.T) {
		dmgAmnt := 6
		attacker := cards.NewDragon("d1")
		target := cards.NewArcher("a1")
		weapon := cards.NewArrow("s1", dmgAmnt)

		p1 := newPlayerWithCards("Player1",
			[]ports.Card{weapon},
			[]ports.Warrior{attacker},
		)
		p2 := newPlayerWithCards("Player2",
			[]ports.Card{},
			[]ports.Warrior{target},
		)
		g := &Game{
			Players: []ports.Player{p1, p2},
		}

		err := g.Attack(p1.Name(), target.GetID(), weapon.GetID())
		assert.NoError(t, err)
		assert.Equal(t, cards.WarriorMaxHealth-dmgAmnt*1, target.Health())
		assert.NotContains(t, p1.Hand().ShowCards(), weapon)
	})
	t.Run("Dragon attacks Archer with Poison causing normal damage", func(t *testing.T) {
		dmgAmnt := 6
		attacker := cards.NewDragon("d1")
		target := cards.NewArcher("a1")
		weapon := cards.NewPoison("s1", dmgAmnt)

		p1 := newPlayerWithCards("Player1",
			[]ports.Card{weapon},
			[]ports.Warrior{attacker},
		)
		p2 := newPlayerWithCards("Player2",
			[]ports.Card{},
			[]ports.Warrior{target},
		)
		g := &Game{
			Players: []ports.Player{p1, p2},
		}

		err := g.Attack(p1.Name(), target.GetID(), weapon.GetID())
		assert.NoError(t, err)
		assert.Equal(t, cards.WarriorMaxHealth-dmgAmnt*1, target.Health())
		assert.NotContains(t, p1.Hand().ShowCards(), weapon)
	})
	t.Run("Dragon attacks Mage with Sword causing normal damage", func(t *testing.T) {
		dmgAmnt := 6
		attacker := cards.NewDragon("d1")
		target := cards.NewMage("m1")
		weapon := cards.NewSword("s1", dmgAmnt)

		p1 := newPlayerWithCards("Player1",
			[]ports.Card{weapon},
			[]ports.Warrior{attacker},
		)
		p2 := newPlayerWithCards("Player2",
			[]ports.Card{},
			[]ports.Warrior{target},
		)
		g := &Game{
			Players: []ports.Player{p1, p2},
		}

		err := g.Attack(p1.Name(), target.GetID(), weapon.GetID())
		assert.NoError(t, err)
		assert.Equal(t, cards.WarriorMaxHealth-dmgAmnt*1, target.Health())
		assert.NotContains(t, p1.Hand().ShowCards(), weapon)
	})
	t.Run("Dragon attacks Mage with Arrow causing double damage", func(t *testing.T) {
		dmgAmnt := 6
		attacker := cards.NewDragon("d1")
		target := cards.NewMage("m1")
		weapon := cards.NewArrow("s1", dmgAmnt)

		p1 := newPlayerWithCards("Player1",
			[]ports.Card{weapon},
			[]ports.Warrior{attacker},
		)
		p2 := newPlayerWithCards("Player2",
			[]ports.Card{},
			[]ports.Warrior{target},
		)
		g := &Game{
			Players: []ports.Player{p1, p2},
		}

		err := g.Attack(p1.Name(), target.GetID(), weapon.GetID())
		assert.NoError(t, err)
		assert.Equal(t, cards.WarriorMaxHealth-dmgAmnt*2, target.Health())
		assert.NotContains(t, p1.Hand().ShowCards(), weapon)
	})
	t.Run("Dragon attacks Mage with Poison causing normal damage", func(t *testing.T) {
		dmgAmnt := 6
		attacker := cards.NewDragon("d1")
		target := cards.NewMage("m1")
		weapon := cards.NewPoison("s1", dmgAmnt)

		p1 := newPlayerWithCards("Player1",
			[]ports.Card{weapon},
			[]ports.Warrior{attacker},
		)
		p2 := newPlayerWithCards("Player2",
			[]ports.Card{},
			[]ports.Warrior{target},
		)
		g := &Game{
			Players: []ports.Player{p1, p2},
		}

		err := g.Attack(p1.Name(), target.GetID(), weapon.GetID())
		assert.NoError(t, err)
		assert.Equal(t, cards.WarriorMaxHealth-dmgAmnt*1, target.Health())
		assert.NotContains(t, p1.Hand().ShowCards(), weapon)
	})
	t.Run("Warrior dead on second attack", func(t *testing.T) {
		dmgAmnt := 5
		k := cards.NewKnight("k1")
		a := cards.NewArcher("a1")
		a2 := cards.NewArcher("a2")
		sword1 := cards.NewSword("s1", dmgAmnt)
		sword2 := cards.NewSword("s2", dmgAmnt)
		g := &Game{}
		p1 := newPlayerWithCardAndObserver("Player1",
			[]ports.Card{sword1, sword2},
			[]ports.Warrior{k},
			g,
		)
		p2 := newPlayerWithCardAndObserver("Player2",
			[]ports.Card{},
			[]ports.Warrior{a, a2},
			g,
		)

		g.Players = []ports.Player{p1, p2}

		err := g.Attack(p1.Name(), a.GetID(), sword1.GetID())
		assert.NoError(t, err)
		assert.NotContains(t, p1.Hand().ShowCards(), sword1)

		err = g.Attack(p1.Name(), a.GetID(), sword2.GetID())
		assert.NoError(t, err)
		assert.NotContains(t, p1.Hand().ShowCards(), sword2)

		assert.Equal(t, 0, a.Health())
		_, ok := p2.GetCardFromField(a.GetID())
		assert.False(t, ok, "Archer should have been removed from field after death")
		_, ok = p2.GetCardFromField(a2.GetID())
		assert.True(t, ok, "Second Archer should still be on the field")
		assert.True(t, foundInCemetery(g, a), "Cemetery should contain the dead archer")
		assert.True(t, foundInDiscardPile(g, sword1), "Discard pile should contain the used sword")
		assert.True(t, foundInDiscardPile(g, sword2), "Discard pile should contain the used sword")
	})
	t.Run("Dragon dead on multiple attacks", func(t *testing.T) {
		dmgAmnt := 5
		m1 := cards.NewMage("m1")
		k2 := cards.NewKnight("k2")
		a3 := cards.NewArcher("a3")

		target := cards.NewDragon("d1")
		a2 := cards.NewArcher("a2")

		poison1 := cards.NewPoison("p1", dmgAmnt)
		sword2 := cards.NewSword("s2", dmgAmnt)
		arrow3 := cards.NewArrow("a3", dmgAmnt)
		sword4 := cards.NewSword("s4", dmgAmnt)

		g := &Game{}
		p1 := newPlayerWithCardAndObserver("Player1",
			[]ports.Card{poison1, sword2, arrow3, sword4},
			[]ports.Warrior{m1, k2, a3},
			g,
		)
		p2 := newPlayerWithCardAndObserver("Player2",
			[]ports.Card{},
			[]ports.Warrior{target, a2},
			g,
		)

		g.Players = []ports.Player{p1, p2}

		err := g.Attack(p1.Name(), target.GetID(), poison1.GetID())
		assert.NoError(t, err)
		err = g.Attack(p1.Name(), target.GetID(), sword2.GetID())
		assert.NoError(t, err)
		err = g.Attack(p1.Name(), target.GetID(), arrow3.GetID())
		assert.NoError(t, err)
		err = g.Attack(p1.Name(), target.GetID(), sword4.GetID())
		assert.NoError(t, err)

		assert.Equal(t, 0, target.Health())
		_, ok := p1.GetCardFromField(m1.GetID())
		assert.True(t, ok, "Mage should still be on the field")
		_, ok = p1.GetCardFromField(k2.GetID())
		assert.True(t, ok, "Knight should still be on the field")
		_, ok = p1.GetCardFromField(a3.GetID())
		assert.True(t, ok, "Archer should still be on the field")

		_, ok = p2.GetCardFromField(target.GetID())
		assert.False(t, ok, "Dragon should have been removed from field after death")
		_, ok = p2.GetCardFromField(a2.GetID())
		assert.True(t, ok, "Archer should still be on the field")

		_, ok = p1.GetCardFromHand(poison1.GetID())
		assert.False(t, ok, "Poison should have been discarded after attack")
		_, ok = p1.GetCardFromHand(sword2.GetID())
		assert.False(t, ok, "Sword should have been discarded after attack")
		_, ok = p1.GetCardFromHand(arrow3.GetID())
		assert.False(t, ok, "Arrow should have been discarded after attack")
		_, ok = p1.GetCardFromHand(sword4.GetID())
		assert.False(t, ok, "Sword should have been discarded after attack")

		assert.True(t, foundInCemetery(g, target), "Cemetery should contain the dead dragon")
		assert.True(t, foundInDiscardPile(g, poison1), "Discard pile should contain the used poison")
		assert.True(t, foundInDiscardPile(g, sword2), "Discard pile should contain the used sword")
		assert.True(t, foundInDiscardPile(g, arrow3), "Discard pile should contain the used arrow")
		assert.True(t, foundInDiscardPile(g, sword4), "Discard pile should contain the used sword")
	})
}

func TestGame_SpecialPower(t *testing.T) {
	t.Run("Use special power of Archer (Instant Kill) on warrior", func(t *testing.T) {
		a := cards.NewArcher("a1")
		target := cards.NewArcher("a2")
		sp := cards.NewSpecialPower("sp")
		g := &Game{}
		p1 := newPlayerWithCardAndObserver("Player1",
			[]ports.Card{sp},
			[]ports.Warrior{a},
			g,
		)
		p2 := newPlayerWithCardAndObserver("Player2",
			[]ports.Card{},
			[]ports.Warrior{target},
			g,
		)

		g.Players = []ports.Player{p1, p2}

		err := g.SpecialPower(p1.Name(), a.GetID(), target.GetID(), sp.GetID())
		assert.NoError(t, err)

		assert.Equal(t, 0, target.Health())
		_, ok := p2.GetCardFromField(target.GetID())
		assert.False(t, ok, "Target should have been removed from field after death")
		_, ok = p1.GetCardFromHand(sp.GetID())
		assert.False(t, ok, "Special Power should have been discarded after attack")
		assert.True(t, foundInCemetery(g, target), "Cemetery should contain the dead target")
		assert.True(t, foundInDiscardPile(g, sp), "Discard pile should contain the used special power")
	})
	t.Run("Use special power of Archer (Instant Kill) on dragon", func(t *testing.T) {
		a := cards.NewArcher("a1")
		target := cards.NewDragon("dr")
		sp := cards.NewSpecialPower("sp")
		g := &Game{}
		p1 := newPlayerWithCardAndObserver("Player1",
			[]ports.Card{sp},
			[]ports.Warrior{a},
			g,
		)
		p2 := newPlayerWithCardAndObserver("Player2",
			[]ports.Card{},
			[]ports.Warrior{target},
			g,
		)

		g.Players = []ports.Player{p1, p2}

		err := g.SpecialPower(p1.Name(), a.GetID(), target.GetID(), sp.GetID())
		assert.NoError(t, err)

		assert.Equal(t, cards.DragonMaxHealth-cards.SpecialPowerDamage, target.Health())
		_, ok := p2.GetCardFromField(target.GetID())
		assert.True(t, ok, "Dragon should still be on the field")

		assert.True(t, findInAttackedBy(target.AttackedBy(), sp.GetID()), "Target should have been marked as attacked by special power")
	})
	t.Run("Use special power of Mage (Heal) on warrior", func(t *testing.T) {
		m := cards.NewMage("m1")
		target := cards.NewKnight("k1")
		attacker := cards.NewArcher("a1")
		arrow := cards.NewArrow("s1", 4)

		sp := cards.NewSpecialPower("sp")
		g := &Game{}
		p1 := newPlayerWithCardAndObserver("Player1",
			[]ports.Card{arrow},
			[]ports.Warrior{attacker},
			g,
		)
		p2 := newPlayerWithCardAndObserver("Player2",
			[]ports.Card{sp},
			[]ports.Warrior{m, target},
			g,
		)

		g.Players = []ports.Player{p1, p2}

		_ = g.Attack(p1.Name(), target.GetID(), arrow.GetID())
		assert.Equal(t, cards.WarriorMaxHealth-4, target.Health())
		err := g.EndTurn(p1.Name())
		assert.NoError(t, err)
		err = g.SpecialPower(p2.Name(), m.GetID(), target.GetID(), sp.GetID())
		assert.NoError(t, err)

		assert.Equal(t, cards.WarriorMaxHealth, target.Health())
		_, ok := p2.GetCardFromHand(sp.GetID())
		assert.False(t, ok, "Special Power should have been discarded after use")
		assert.True(t, foundInDiscardPile(g, sp), "Discard pile should contain the used special power")

	})
	t.Run("Use special power of Knight (Protection) on warrior", func(t *testing.T) {
		user := cards.NewKnight("k1")
		target := cards.NewKnight("k2")
		attacker := cards.NewArcher("a1")
		arrow := cards.NewArrow("a1", 4)
		arrow2 := cards.NewArrow("a2", 8)

		sp := cards.NewSpecialPower("sp")
		g := &Game{}
		p1 := newPlayerWithCardAndObserver("Player1",
			[]ports.Card{sp},
			[]ports.Warrior{user, target},
			g,
		)
		p2 := newPlayerWithCardAndObserver("Player2",
			[]ports.Card{arrow, arrow2},
			[]ports.Warrior{attacker},
			g,
		)

		g.Players = []ports.Player{p1, p2}

		err := g.SpecialPower(p1.Name(), user.GetID(), target.GetID(), sp.GetID())
		assert.NoError(t, err)
		assert.NotContains(t, p1.Hand().ShowCards(), sp)
		isProtected, card := target.IsProtected()
		assert.True(t, isProtected)
		assert.Equal(t, card, sp)
		_ = g.EndTurn(p1.Name())

		_ = g.Attack(p2.Name(), target.GetID(), arrow.GetID())
		assert.Equal(t, cards.SpecialPowerMaxHealth-4, sp.Health())
		assert.Equal(t, cards.WarriorMaxHealth, target.Health())

		_ = g.Attack(p2.Name(), target.GetID(), arrow2.GetID())
		assert.Equal(t, cards.SpecialPowerMaxHealth-4-8, sp.Health())
		assert.Equal(t, cards.WarriorMaxHealth, target.Health())

		assert.True(t, foundInDiscardPile(g, sp), "Discard pile should contain the used special power")

	})
}

func TestDrawCards(t *testing.T) {
	t.Run("Take card when deck is empty", func(t *testing.T) {
		p := newPlayerWithCards("Player1", []ports.Card{}, []ports.Warrior{})
		g := &Game{
			Players: []ports.Player{p},
			deck:    NewDeck([]ports.Card{}),
			discardPile: []ports.Card{
				cards.NewSword("s1", 4),
				cards.NewArrow("a1", 3),
				cards.NewPoison("p1", 4),
			},
			cemetery: []ports.Warrior{},
		}

		err := g.DrawCards(p.Name(), 1)
		assert.NoError(t, err)
		assert.Equal(t, 1, p.CardsInHand(), "Player should have drawn one card from reshuffled deck")
		assert.Equal(t, 2, len(g.deck.(*deck).cards), "Deck should have two cards remaining after drawing one")
		assert.Equal(t, 0, len(g.discardPile), "Discard pile should be empty after reshuffling into deck")
	})
	t.Run("Take card from deck to hand", func(t *testing.T) {
		p := newPlayerWithCards("Player1",
			[]ports.Card{cards.NewGold("g1", 5)},
			[]ports.Warrior{})
		g := &Game{
			Players: []ports.Player{p},
			deck: NewDeck([]ports.Card{
				cards.NewSword("s1", 4),
				cards.NewArrow("a1", 3),
				cards.NewPoison("p1", 4),
			}),
			discardPile: []ports.Card{},
			cemetery:    []ports.Warrior{},
		}

		err := g.DrawCards(p.Name(), 1)
		assert.NoError(t, err)
		assert.Equal(t, 2, p.CardsInHand(), "Player should have drawn two cards from deck")
		assert.Equal(t, 2, len(g.deck.(*deck).cards), "Deck should have one card remaining after drawing two")
	})
}

func TestNewGame(t *testing.T) {
	t.Run("Create new game with two players getting expected number of cards", func(t *testing.T) {
		p1 := "Alice"
		p2 := "Bob"
		g := NewGame(p1, p2, cards.NewDealer())

		assert.Equal(t, 2, len(g.Players), "Game should have two players")
		assert.Equal(t, 7, g.Players[0].CardsInHand(), "Each player should start with 7 cards in hand")
		assert.Equal(t, 7, g.Players[1].CardsInHand(), "Each player should start with 7 cards in hand")
		assert.Equal(t, 46, len(g.deck.(*deck).cards), "Deck should start with 40 cards")
		assert.Equal(t, g.state, StateSettingInitialWarriors)
	})
	t.Run("Set initial warriors for players", func(t *testing.T) {
		p1 := "Alice"
		p2 := "Bob"
		g := NewGame(p1, p2, cards.NewDealer())

		current, _ := g.WhoIsCurrent()
		cont := 0
		var warriors1 []string
		for _, card := range current.Hand().ShowCards() {
			if _, ok := card.(ports.Warrior); ok {
				cont++
				warriors1 = append(warriors1, card.GetID())
				if cont == 3 {
					break
				}
			}
		}

		err := g.SetInitialWarriors(current.Name(), warriors1)
		assert.NoError(t, err)
		assert.Equal(t, len(current.Field().Warriors()), len(warriors1))
		assert.True(t, fieldContainsCardWithID(current.Field(), warriors1[0]), "Field should contain the warrior with the given ID")
		assert.True(t, fieldContainsCardWithID(current.Field(), warriors1[1]), "Field should contain the warrior with the given ID")
		assert.True(t, fieldContainsCardWithID(current.Field(), warriors1[2]), "Field should contain the warrior with the given ID")
		assert.False(t, handContainsCardWithID(current.Hand(), warriors1[0]), "Hand should not contain the warrior with the given ID")
		assert.False(t, handContainsCardWithID(current.Hand(), warriors1[1]), "Hand should not contain the warrior with the given ID")
		assert.False(t, handContainsCardWithID(current.Hand(), warriors1[2]), "Hand should not contain the warrior with the given ID")
		assert.Equal(t, 4, current.CardsInHand(), "Player should have 4 cards left in hand after setting 3 warriors")

		current, _ = g.WhoIsCurrent()
		cont = 0
		var warriors2 []string
		for _, card := range current.Hand().ShowCards() {
			if _, ok := card.(ports.Warrior); ok {
				cont++
				warriors2 = append(warriors2, card.GetID())
				if cont == 2 {
					break
				}
			}
		}

		err = g.SetInitialWarriors(current.Name(), warriors2)
		assert.NoError(t, err)
		assert.Equal(t, len(current.Field().Warriors()), len(warriors2))
		assert.True(t, fieldContainsCardWithID(current.Field(), warriors2[0]), "Field should contain the warrior with the given ID")
		assert.True(t, fieldContainsCardWithID(current.Field(), warriors2[1]), "Field should contain the warrior with the given ID")
		assert.False(t, handContainsCardWithID(current.Hand(), warriors2[0]), "Hand should not contain the warrior with the given ID")
		assert.False(t, handContainsCardWithID(current.Hand(), warriors2[1]), "Hand should not contain the warrior with the given ID")
		assert.Equal(t, 5, current.CardsInHand(), "Player should have 5 cards left in hand after setting 2 warriors")
		assert.Equal(t, StateWaitingDraw, g.state)
	})
}

func TestGame_Spy(t *testing.T) {
	t.Run("Spy reveals top cards of deck", func(t *testing.T) {
		spy := cards.NewSpy("s1")

		p1 := newPlayerWithCards("Player1",
			[]ports.Card{spy},
			[]ports.Warrior{},
		)
		p2 := newPlayerWithCards("Player2", nil, nil)
		deckCards := []ports.Card{
			cards.NewGold("g1", 5),
			cards.NewSword("sw1", 4),
			cards.NewArrow("a1", 3),
			cards.NewArrow("a2", 5),
			cards.NewDragon("d1"),
		}
		g := &Game{
			Players: []ports.Player{p1, p2},
			deck:    NewDeck(deckCards),
		}

		revealedCards, err := g.Spy(p1.Name(), 1)
		assert.NoError(t, err)
		assert.Len(t, revealedCards, 5, "Spy should reveal five cards from the top of the deck")
		assert.Contains(t, revealedCards, deckCards[0], "Revealed cards should contain the first top card of the deck")
		assert.Contains(t, revealedCards, deckCards[1], "Revealed cards should contain the second top card of the deck")
		assert.Contains(t, revealedCards, deckCards[2], "Revealed cards should contain the third top card of the deck")
		assert.Contains(t, revealedCards, deckCards[3], "Revealed cards should contain the fourth top card of the deck")
		assert.Contains(t, revealedCards, deckCards[4], "Revealed cards should contain the fifth top card of the deck")
		_, hasSpy := p1.Hand().GetCard(spy.GetID())
		assert.False(t, hasSpy)
	})
	t.Run("Spy reveals opponent's hand", func(t *testing.T) {
		spy := cards.NewSpy("s1")
		enemyCard1 := cards.NewGold("g1", 5)
		enemyCard2 := cards.NewSword("sw1", 4)

		p1 := newPlayerWithCards("Player1",
			[]ports.Card{spy},
			[]ports.Warrior{},
		)
		p2 := newPlayerWithCards("Player2",
			[]ports.Card{enemyCard1, enemyCard2},
			[]ports.Warrior{},
		)
		g := &Game{
			Players: []ports.Player{p1, p2},
		}

		revealedCards, err := g.Spy(p1.Name(), 2)
		assert.NoError(t, err)
		assert.Len(t, revealedCards, 2, "Spy should reveal two cards from opponent's hand")
		assert.Contains(t, revealedCards, enemyCard1, "Revealed cards should contain the first enemy card")
		assert.Contains(t, revealedCards, enemyCard2, "Revealed cards should contain the second enemy card")
		_, hasSpy := p1.Hand().GetCard(spy.GetID())
		assert.False(t, hasSpy)
	})
}

func TestGame_Steal(t *testing.T) {
	t.Run("Steal a card from opponent's hand", func(t *testing.T) {
		stealCard := cards.NewThief("t1")
		enemyCard1 := cards.NewGold("g1", 5)
		enemyCard2 := cards.NewSword("sw1", 4)
		enemyCard3 := cards.NewSword("sw2", 7)

		p1 := newPlayerWithCards("Player1",
			[]ports.Card{stealCard},
			[]ports.Warrior{},
		)
		p2 := newPlayerWithCards("Player2",
			[]ports.Card{enemyCard1, enemyCard2, enemyCard3},
			[]ports.Warrior{},
		)
		g := &Game{
			Players: []ports.Player{p1, p2},
		}

		err := g.Steal(p1.Name(), 2)
		assert.NoError(t, err)

		_, hasSteal := p1.Hand().GetCard(stealCard.GetID())
		assert.False(t, hasSteal, "Steal card should be discarded after use")
		assert.Len(t, p1.Hand().ShowCards(), 1, "Player should have one more card in hand after stealing")
		assert.Len(t, p2.Hand().ShowCards(), 2, "Enemy should have one less card in hand after being stolen from")
	})
}

func findInAttackedBy(cards []ports.Weapon, id string) bool {
	for _, c := range cards {
		if c != nil && c.GetID() == id {
			return true
		}
	}
	return false
}

func foundInCemetery(g *Game, a ports.Warrior) bool {
	for _, w := range g.cemetery {
		if w == a || (w != nil && w.GetID() == a.GetID()) {
			return true
		}
	}
	return false
}

func foundInDiscardPile(g *Game, a ports.Card) bool {
	for _, w := range g.discardPile {
		if w == a || (w != nil && w.GetID() == a.GetID()) {
			return true
		}
	}
	return false
}

func fieldContainsCardWithID(field ports.Field, id string) bool {
	for _, c := range field.Warriors() {
		if c != nil && c.GetID() == id {
			return true
		}
	}
	return false
}

func handContainsCardWithID(hand ports.Hand, id string) bool {
	for _, c := range hand.ShowCards() {
		if c != nil && c.GetID() == id {
			return true
		}
	}
	return false
}

func newPlayerWithCards(name string, cardsInHand []ports.Card,
	cardsInField []ports.Warrior) ports.Player {
	return newPlayerWithCardAndObserver(name, cardsInHand, cardsInField, nil)
}

func newPlayerWithCardAndObserver(name string, cardsInHand []ports.Card,
	cardsInField []ports.Warrior, game *Game) ports.Player {
	p := &player{
		name:                           name,
		cardMovedToPileObserver:        game,
		warriorMovedToCemeteryObserver: game,
		hand: &hand{
			cards: cardsInHand,
		},
		field: &field{
			cards: cardsInField,
		},
		castle: &castle{},
	}

	for _, card := range cardsInField {
		card.AddCardMovedToPileObserver(p)
		targ, ok := card.(ports.Warrior)
		if ok {
			targ.AddWarriorDeadObserver(p)
		}
	}
	for _, card := range cardsInHand {
		card.AddCardMovedToPileObserver(p)
	}

	return p
}
