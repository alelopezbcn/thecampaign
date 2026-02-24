package cards_test

import (
	"testing"

	"github.com/alelopezbcn/thecampaign/internal/domain/cards"
	"github.com/alelopezbcn/thecampaign/internal/domain/types"
	"github.com/alelopezbcn/thecampaign/test/mocks"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

// testObserver tracks OnCardMovedToPile calls.
type testObserver struct {
	called []cards.Card
}

func (o *testObserver) OnCardMovedToPile(c cards.Card) { o.called = append(o.called, c) }

func TestNewSpecialPower(t *testing.T) {
	sp := cards.NewSpecialPower("sp1")
	assert.Equal(t, "SP1", sp.GetID())
	assert.Equal(t, 10, sp.Health())
	assert.Equal(t, 10, sp.DamageAmount())
	assert.Contains(t, sp.String(), "Special Power")
	assert.Contains(t, sp.String(), "10")
}

func TestSpecialPower_Use_ByKnight(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	sp := cards.NewSpecialPower("sp1")
	knight := cards.NewKnight("k1")
	target := mocks.NewMockWarrior(ctrl)
	target.EXPECT().Type().Return(types.KnightWarriorType)
	target.EXPECT().Protect(sp)

	err := sp.Use(knight, target)
	assert.NoError(t, err)
}

func TestSpecialPower_Use_ByKnight_OnDragon(t *testing.T) {
	sp := cards.NewSpecialPower("sp1")
	knight := cards.NewKnight("k1")
	target := cards.NewDragon("d1")

	err := sp.Use(knight, target)
	assert.EqualError(t, err, "dragon cannot be protected")
}

func TestSpecialPower_Use_ByArcher(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	sp := cards.NewSpecialPower("sp1")
	archer := cards.NewArcher("a1")
	target := mocks.NewMockWarrior(ctrl)
	target.EXPECT().InstantKill(sp)

	err := sp.Use(archer, target)
	assert.NoError(t, err)
}

func TestSpecialPower_Use_ByArcher_OnDragon(t *testing.T) {
	sp := cards.NewSpecialPower("sp1")
	archer := cards.NewArcher("a1")
	target := cards.NewDragon("d1")

	err := sp.Use(archer, target)
	assert.NoError(t, err)
}

func TestSpecialPower_Use_ByMage(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	sp := cards.NewSpecialPower("sp1")
	mage := cards.NewMage("m1")
	target := mocks.NewMockWarrior(ctrl)
	target.EXPECT().Type().Return(types.KnightWarriorType)
	target.EXPECT().Heal(sp)

	err := sp.Use(mage, target)
	assert.NoError(t, err)
}

func TestSpecialPower_Use_ByMage_OnDragon(t *testing.T) {
	sp := cards.NewSpecialPower("sp1")
	mage := cards.NewMage("m1")
	target := cards.NewDragon("d1")

	err := sp.Use(mage, target)
	assert.EqualError(t, err, "dragon cannot be healed")
}

func TestSpecialPower_Use_ByDragon(t *testing.T) {
	sp := cards.NewSpecialPower("sp1")
	dragon := cards.NewDragon("d1")
	target := cards.NewKnight("k1")

	err := sp.Use(dragon, target)
	assert.EqualError(t, err, "special power action not allowed to be used by Dragon")
}

func TestSpecialPower_Use_ByUnknownType(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	sp := cards.NewSpecialPower("sp1")
	// MockWarrior is not *knight/*archer/*mage/*dragon, so it falls to the default case.
	unknown := mocks.NewMockWarrior(ctrl)
	unknown.EXPECT().Type().Return(types.KnightWarriorType)
	target := cards.NewKnight("k1")

	err := sp.Use(unknown, target)
	assert.EqualError(t, err, "special power action not allowed for this warrior type")
}

func TestSpecialPower_Destroyed(t *testing.T) {
	obs := &testObserver{}

	// Use real swords as attacking weapons; register obs as their observer.
	sword1 := cards.NewSword("s1", 4)
	sword1.AddCardMovedToPileObserver(obs)
	sword2 := cards.NewSword("s2", 4)
	sword2.AddCardMovedToPileObserver(obs)

	sp := cards.NewSpecialPower("sp1")
	sp.AddCardMovedToPileObserver(obs)

	// Non-lethal attacks so the swords end up in attackedBy.
	sp.BeAttacked(sword1) // health: 10 -> 6
	sp.BeAttacked(sword2) // health: 6  -> 2

	sp.Destroyed()

	assert.Contains(t, obs.called, cards.Card(sword1))
	assert.Contains(t, obs.called, cards.Card(sword2))
	assert.Empty(t, sp.AttackedBy())
}

func TestSpecialPower_ReceiveDamage_NotDefeated(t *testing.T) {
	sp := cards.NewSpecialPower("sp1")
	weapon := cards.NewSword("w1", 4)

	defeated := sp.ReceiveDamage(weapon, 1)

	assert.False(t, defeated)
	assert.Equal(t, 6, sp.Health())
	assert.Len(t, sp.AttackedBy(), 1)
	assert.Equal(t, cards.Weapon(weapon), sp.AttackedBy()[0])
}

func TestSpecialPower_ReceiveDamage_Defeated(t *testing.T) {
	obs := &testObserver{}

	weapon := cards.NewSword("w1", 15)
	weapon.AddCardMovedToPileObserver(obs)

	sp := cards.NewSpecialPower("sp1")
	sp.AddCardMovedToPileObserver(obs)

	defeated := sp.ReceiveDamage(weapon, 1)

	assert.True(t, defeated)
	assert.LessOrEqual(t, sp.Health(), 0)
	assert.Empty(t, sp.AttackedBy())
}

func TestSpecialPower_String_AliveWithWeapons(t *testing.T) {
	sp := cards.NewSpecialPower("sp1")
	str := sp.String()
	assert.Contains(t, str, "Special Power")
	assert.Contains(t, str, "10")
}
