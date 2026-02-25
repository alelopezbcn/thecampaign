package board_test

import (
	"testing"

	"github.com/alelopezbcn/thecampaign/internal/domain/board"
	"github.com/alelopezbcn/thecampaign/internal/domain/cards"
	"github.com/alelopezbcn/thecampaign/internal/domain/types"
	"github.com/alelopezbcn/thecampaign/test/mocks"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

func newTestPlayer(t *testing.T, ctrl *gomock.Controller) (
	board.Player,
	*mocks.MockCardMovedToPileObserver,
	*mocks.MockWarriorMovedToCemeteryObserver,
	*mocks.MockCastleCompletionObserver,
	*mocks.MockFieldWithoutWarriorsObserver,
) {
	pileObs := mocks.NewMockCardMovedToPileObserver(ctrl)
	cemeteryObs := mocks.NewMockWarriorMovedToCemeteryObserver(ctrl)
	castleObs := mocks.NewMockCastleCompletionObserver(ctrl)
	fieldObs := mocks.NewMockFieldWithoutWarriorsObserver(ctrl)

	p := board.NewPlayer("TestPlayer", 0, pileObs, cemeteryObs, castleObs, fieldObs, 10)
	return p, pileObs, cemeteryObs, castleObs, fieldObs
}

func mockCardWithObserver(ctrl *gomock.Controller, id string) *mocks.MockCard {
	c := mocks.NewMockCard(ctrl)
	c.EXPECT().GetID().Return(id).AnyTimes()
	c.EXPECT().AddCardMovedToPileObserver(gomock.Any()).AnyTimes()
	return c
}

func mockWarriorCard(ctrl *gomock.Controller, id string, wType types.WarriorType) *mocks.MockWarrior {
	w := mocks.NewMockWarrior(ctrl)
	w.EXPECT().GetID().Return(id).AnyTimes()
	w.EXPECT().AddCardMovedToPileObserver(gomock.Any()).AnyTimes()
	w.EXPECT().AddWarriorDeadObserver(gomock.Any()).AnyTimes()
	w.EXPECT().Type().Return(wType).AnyTimes()
	return w
}

func TestPlayer_NewPlayer(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	p, _, _, _, _ := newTestPlayer(t, ctrl)

	assert.Equal(t, "TestPlayer", p.Name())
	assert.Equal(t, 0, p.Idx())
	assert.NotNil(t, p.Hand())
	assert.NotNil(t, p.Field())
	assert.NotNil(t, p.Castle())
	assert.Equal(t, 0, p.CardsInHand())
}

func TestPlayer_TakeCards(t *testing.T) {
	t.Run("Takes cards successfully", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()
		p, _, _, _, _ := newTestPlayer(t, ctrl)

		card := mockCardWithObserver(ctrl, "c1")

		ok := p.TakeCards(card)
		assert.True(t, ok)
		assert.Equal(t, 1, p.CardsInHand())
	})

	t.Run("Takes warrior cards and registers observer", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()
		p, _, _, _, _ := newTestPlayer(t, ctrl)

		warrior := mockWarriorCard(ctrl, "w1", types.KnightWarriorType)

		ok := p.TakeCards(warrior)
		assert.True(t, ok)
		assert.Equal(t, 1, p.CardsInHand())
	})

	t.Run("Fails when hand limit exceeded", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()
		p, _, _, _, _ := newTestPlayer(t, ctrl)

		// Fill hand to max (7 cards)
		for i := 0; i < 7; i++ {
			c := mockCardWithObserver(ctrl, "c"+string(rune('0'+i)))
			p.TakeCards(c)
		}

		extra := mockCardWithObserver(ctrl, "extra")
		ok := p.TakeCards(extra)
		assert.False(t, ok)
		assert.Equal(t, 7, p.CardsInHand())
	})
}

func TestPlayer_CanTakeCards(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	p, _, _, _, _ := newTestPlayer(t, ctrl)

	assert.True(t, p.CanTakeCards(7))
	assert.False(t, p.CanTakeCards(8))

	card := mockCardWithObserver(ctrl, "c1")
	p.TakeCards(card)

	assert.True(t, p.CanTakeCards(6))
	assert.False(t, p.CanTakeCards(7))
}

func TestPlayer_GiveCards(t *testing.T) {
	t.Run("Gives cards successfully", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()
		p, _, _, _, _ := newTestPlayer(t, ctrl)

		card := mockCardWithObserver(ctrl, "c1")
		p.TakeCards(card)

		given, err := p.RemoveFromHand("c1")
		assert.NoError(t, err)
		assert.Len(t, given, 1)
		assert.Equal(t, 0, p.CardsInHand())
	})

	t.Run("Error when card not found", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()
		p, _, _, _, _ := newTestPlayer(t, ctrl)

		_, err := p.RemoveFromHand("nonexistent")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "card with ID nonexistent not found in hand")
	})

	t.Run("Gives multiple cards", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()
		p, _, _, _, _ := newTestPlayer(t, ctrl)

		c1 := mockCardWithObserver(ctrl, "c1")
		c2 := mockCardWithObserver(ctrl, "c2")
		c3 := mockCardWithObserver(ctrl, "c3")
		p.TakeCards(c1, c2, c3)

		given, err := p.RemoveFromHand("c1", "c3")
		assert.NoError(t, err)
		assert.Len(t, given, 2)
		assert.Equal(t, 1, p.CardsInHand())
	})
}

func TestPlayer_GetCardFromHand(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	p, _, _, _, _ := newTestPlayer(t, ctrl)

	card := mockCardWithObserver(ctrl, "c1")
	p.TakeCards(card)

	found, ok := p.GetCardFromHand("c1")
	assert.True(t, ok)
	assert.Equal(t, card, found)

	_, ok = p.GetCardFromHand("nonexistent")
	assert.False(t, ok)
}

func TestPlayer_GetCardFromField(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	p, _, _, _, _ := newTestPlayer(t, ctrl)

	warrior := mockWarriorCard(ctrl, "w1", types.KnightWarriorType)
	p.TakeCards(warrior)
	err := p.MoveCardToField("w1")
	assert.NoError(t, err)

	found, ok := p.GetCardFromField("w1")
	assert.True(t, ok)
	assert.Equal(t, warrior, found)

	_, ok = p.GetCardFromField("nonexistent")
	assert.False(t, ok)
}

func TestPlayer_MoveCardToField(t *testing.T) {
	t.Run("Moves warrior to field", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()
		p, _, _, _, _ := newTestPlayer(t, ctrl)

		warrior := mockWarriorCard(ctrl, "w1", types.KnightWarriorType)
		p.TakeCards(warrior)

		err := p.MoveCardToField("w1")
		assert.NoError(t, err)
		assert.Equal(t, 0, p.CardsInHand())

		_, ok := p.GetCardFromField("w1")
		assert.True(t, ok)
	})

	t.Run("Error when card not in hand", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()
		p, _, _, _, _ := newTestPlayer(t, ctrl)

		err := p.MoveCardToField("nonexistent")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "card with ID nonexistent not found in hand")
	})

	t.Run("Error when card is not a warrior", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()
		p, _, _, _, _ := newTestPlayer(t, ctrl)

		card := mockCardWithObserver(ctrl, "c1")
		p.TakeCards(card)

		err := p.MoveCardToField("c1")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "warrior or dragon cards can be moved to field")
	})
}


func TestPlayer_CanAttack(t *testing.T) {
	t.Run("True with sword and knight", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()
		p, _, _, _, _ := newTestPlayer(t, ctrl)

		knight := mockWarriorCard(ctrl, "k1", types.KnightWarriorType)
		p.TakeCards(knight)
		p.MoveCardToField("k1")

		weapon := mocks.NewMockWeapon(ctrl)
		weapon.EXPECT().GetID().Return("sw1").AnyTimes()
		weapon.EXPECT().AddCardMovedToPileObserver(gomock.Any()).AnyTimes()
		weapon.EXPECT().Type().Return(types.SwordWeaponType).AnyTimes()
		p.TakeCards(weapon)

		assert.True(t, p.CanAttack())
	})

	t.Run("True with arrow and archer", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()
		p, _, _, _, _ := newTestPlayer(t, ctrl)

		archer := mockWarriorCard(ctrl, "a1", types.ArcherWarriorType)
		p.TakeCards(archer)
		p.MoveCardToField("a1")

		weapon := mocks.NewMockWeapon(ctrl)
		weapon.EXPECT().GetID().Return("ar1").AnyTimes()
		weapon.EXPECT().AddCardMovedToPileObserver(gomock.Any()).AnyTimes()
		weapon.EXPECT().Type().Return(types.ArrowWeaponType).AnyTimes()
		p.TakeCards(weapon)

		assert.True(t, p.CanAttack())
	})

	t.Run("True with poison and mage", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()
		p, _, _, _, _ := newTestPlayer(t, ctrl)

		mage := mockWarriorCard(ctrl, "m1", types.MageWarriorType)
		p.TakeCards(mage)
		p.MoveCardToField("m1")

		weapon := mocks.NewMockWeapon(ctrl)
		weapon.EXPECT().GetID().Return("po1").AnyTimes()
		weapon.EXPECT().AddCardMovedToPileObserver(gomock.Any()).AnyTimes()
		weapon.EXPECT().Type().Return(types.PoisonWeaponType).AnyTimes()
		p.TakeCards(weapon)

		assert.True(t, p.CanAttack())
	})

	t.Run("True with dragon and any weapon", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()
		p, _, _, _, _ := newTestPlayer(t, ctrl)

		dragon := mockWarriorCard(ctrl, "d1", types.DragonWarriorType)
		p.TakeCards(dragon)
		p.MoveCardToField("d1")

		weapon := mocks.NewMockWeapon(ctrl)
		weapon.EXPECT().GetID().Return("sw1").AnyTimes()
		weapon.EXPECT().AddCardMovedToPileObserver(gomock.Any()).AnyTimes()
		weapon.EXPECT().Type().Return(types.SwordWeaponType).AnyTimes()
		p.TakeCards(weapon)

		assert.True(t, p.CanAttack())
	})

	t.Run("True with special power and archer", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()
		p, _, _, _, _ := newTestPlayer(t, ctrl)

		archer := mockWarriorCard(ctrl, "a1", types.ArcherWarriorType)
		p.TakeCards(archer)
		p.MoveCardToField("a1")

		weapon := mocks.NewMockWeapon(ctrl)
		weapon.EXPECT().GetID().Return("sp1").AnyTimes()
		weapon.EXPECT().AddCardMovedToPileObserver(gomock.Any()).AnyTimes()
		weapon.EXPECT().Type().Return(types.SpecialPowerWeaponType).AnyTimes()
		p.TakeCards(weapon)

		assert.True(t, p.CanAttack())
	})

	t.Run("False with no weapons", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()
		p, _, _, _, _ := newTestPlayer(t, ctrl)

		knight := mockWarriorCard(ctrl, "k1", types.KnightWarriorType)
		p.TakeCards(knight)
		p.MoveCardToField("k1")

		assert.False(t, p.CanAttack())
	})

	t.Run("False with wrong weapon for warrior type", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()
		p, _, _, _, _ := newTestPlayer(t, ctrl)

		archer := mockWarriorCard(ctrl, "a1", types.ArcherWarriorType)
		p.TakeCards(archer)
		p.MoveCardToField("a1")

		weapon := mocks.NewMockWeapon(ctrl)
		weapon.EXPECT().GetID().Return("sw1").AnyTimes()
		weapon.EXPECT().AddCardMovedToPileObserver(gomock.Any()).AnyTimes()
		weapon.EXPECT().Type().Return(types.SwordWeaponType).AnyTimes()
		p.TakeCards(weapon)

		assert.False(t, p.CanAttack())
	})
}

func TestPlayer_CanBuy(t *testing.T) {
	t.Run("True with buyable resource", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()
		p, _, _, _, _ := newTestPlayer(t, ctrl)

		resource := mocks.NewMockResource(ctrl)
		resource.EXPECT().GetID().Return("r1").AnyTimes()
		resource.EXPECT().AddCardMovedToPileObserver(gomock.Any()).AnyTimes()
		resource.EXPECT().CanConstruct().Return(false).AnyTimes()
		resource.EXPECT().Value().Return(2).AnyTimes()
		p.TakeCards(resource)

		assert.True(t, p.CanBuy())
	})

	t.Run("False with construct-only resource", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()
		p, _, _, _, _ := newTestPlayer(t, ctrl)

		resource := mocks.NewMockResource(ctrl)
		resource.EXPECT().GetID().Return("r1").AnyTimes()
		resource.EXPECT().AddCardMovedToPileObserver(gomock.Any()).AnyTimes()
		resource.EXPECT().CanConstruct().Return(true).AnyTimes()
		p.TakeCards(resource)

		assert.False(t, p.CanBuy())
	})

	t.Run("False when hand would exceed limit", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()
		p, _, _, _, _ := newTestPlayer(t, ctrl)

		// Fill hand to 7 cards: 6 non-resources + 1 resource with value 6 (buys 3)
		for i := 0; i < 6; i++ {
			c := mockCardWithObserver(ctrl, "c"+string(rune('0'+i)))
			p.TakeCards(c)
		}
		resource := mocks.NewMockResource(ctrl)
		resource.EXPECT().GetID().Return("r1").AnyTimes()
		resource.EXPECT().AddCardMovedToPileObserver(gomock.Any()).AnyTimes()
		resource.EXPECT().CanConstruct().Return(false).AnyTimes()
		resource.EXPECT().Value().Return(6).AnyTimes() // buys 3, but 7+3-1=9 > 7
		p.TakeCards(resource)

		assert.False(t, p.CanBuy())
	})

	t.Run("False with no resources", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()
		p, _, _, _, _ := newTestPlayer(t, ctrl)

		card := mockCardWithObserver(ctrl, "c1")
		p.TakeCards(card)

		assert.False(t, p.CanBuy())
	})
}

func TestPlayer_CanBuyWith(t *testing.T) {
	t.Run("True for valid resource", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()
		p, _, _, _, _ := newTestPlayer(t, ctrl)

		resource := mocks.NewMockResource(ctrl)
		resource.EXPECT().CanConstruct().Return(false)
		resource.EXPECT().Value().Return(2) // buys 1 card: 0+1-1=0 <= 7

		assert.True(t, p.CanBuyWith(resource))
	})

	t.Run("False for construct-only resource", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()
		p, _, _, _, _ := newTestPlayer(t, ctrl)

		resource := mocks.NewMockResource(ctrl)
		resource.EXPECT().CanConstruct().Return(true)

		assert.False(t, p.CanBuyWith(resource))
	})
}

func TestPlayer_CanConstruct(t *testing.T) {
	t.Run("True with constructable resource (castle not started)", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()
		p, _, _, _, _ := newTestPlayer(t, ctrl)

		resource := mocks.NewMockResource(ctrl)
		resource.EXPECT().GetID().Return("r1").AnyTimes()
		resource.EXPECT().AddCardMovedToPileObserver(gomock.Any()).AnyTimes()
		resource.EXPECT().CanConstruct().Return(true).AnyTimes()
		p.TakeCards(resource)

		assert.True(t, p.CanConstruct())
	})

	t.Run("False with non-constructable resource (castle not started)", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()
		p, _, _, _, _ := newTestPlayer(t, ctrl)

		resource := mocks.NewMockResource(ctrl)
		resource.EXPECT().GetID().Return("r1").AnyTimes()
		resource.EXPECT().AddCardMovedToPileObserver(gomock.Any()).AnyTimes()
		resource.EXPECT().CanConstruct().Return(false).AnyTimes()
		resource.EXPECT().Value().Return(4).AnyTimes()
		p.TakeCards(resource)

		assert.False(t, p.CanConstruct())
	})

	t.Run("True with constructable weapon (castle not started)", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()
		p, _, _, _, _ := newTestPlayer(t, ctrl)

		weapon := mocks.NewMockWeapon(ctrl)
		weapon.EXPECT().GetID().Return("w1").AnyTimes()
		weapon.EXPECT().AddCardMovedToPileObserver(gomock.Any()).AnyTimes()
		weapon.EXPECT().CanConstruct().Return(true).AnyTimes()
		p.TakeCards(weapon)

		assert.True(t, p.CanConstruct())
	})

	t.Run("False with no constructable cards", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()
		p, _, _, _, _ := newTestPlayer(t, ctrl)

		card := mockCardWithObserver(ctrl, "c1")
		p.TakeCards(card)

		assert.False(t, p.CanConstruct())
	})
}

func TestPlayer_HasWarriorsInHand(t *testing.T) {
	t.Run("True with warrior in hand", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()
		p, _, _, _, _ := newTestPlayer(t, ctrl)

		warrior := mockWarriorCard(ctrl, "w1", types.KnightWarriorType)
		p.TakeCards(warrior)

		assert.True(t, p.HasWarriorsInHand())
	})

	t.Run("False after moving warrior to field", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()
		p, _, _, _, _ := newTestPlayer(t, ctrl)

		warrior := mockWarriorCard(ctrl, "w1", types.KnightWarriorType)
		p.TakeCards(warrior)
		p.MoveCardToField("w1")

		assert.False(t, p.HasWarriorsInHand())
	})
}

func TestPlayer_CanTradeCards(t *testing.T) {
	t.Run("True with 3+ tradeable weapons", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()
		p, _, _, _, _ := newTestPlayer(t, ctrl)

		for i := 0; i < 3; i++ {
			w := mocks.NewMockWeapon(ctrl)
			w.EXPECT().GetID().Return("w" + string(rune('0'+i))).AnyTimes()
			w.EXPECT().AddCardMovedToPileObserver(gomock.Any()).AnyTimes()
			w.EXPECT().Type().Return(types.SwordWeaponType).AnyTimes()
			p.TakeCards(w)
		}

		assert.True(t, p.CanTradeCards())
	})

	t.Run("False with less than 3 tradeable weapons", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()
		p, _, _, _, _ := newTestPlayer(t, ctrl)

		for i := 0; i < 2; i++ {
			w := mocks.NewMockWeapon(ctrl)
			w.EXPECT().GetID().Return("w" + string(rune('0'+i))).AnyTimes()
			w.EXPECT().AddCardMovedToPileObserver(gomock.Any()).AnyTimes()
			w.EXPECT().Type().Return(types.SwordWeaponType).AnyTimes()
			p.TakeCards(w)
		}

		assert.False(t, p.CanTradeCards())
	})

	t.Run("False with 2 tradeable weapons and 1 special power", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()
		p, _, _, _, _ := newTestPlayer(t, ctrl)

		for i := 0; i < 2; i++ {
			w := mocks.NewMockWeapon(ctrl)
			w.EXPECT().GetID().Return("w" + string(rune('0'+i))).AnyTimes()
			w.EXPECT().AddCardMovedToPileObserver(gomock.Any()).AnyTimes()
			w.EXPECT().Type().Return(types.SwordWeaponType).AnyTimes()
			p.TakeCards(w)
		}
		sp := mocks.NewMockWeapon(ctrl)
		sp.EXPECT().GetID().Return("sp1").AnyTimes()
		sp.EXPECT().AddCardMovedToPileObserver(gomock.Any()).AnyTimes()
		sp.EXPECT().Type().Return(types.SpecialPowerWeaponType).AnyTimes()
		p.TakeCards(sp)

		assert.False(t, p.CanTradeCards())
	})
}


func TestPlayer_OnCardMovedToPile(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	p, pileObs, _, _, _ := newTestPlayer(t, ctrl)
	pp := p.(cards.CardMovedToPileObserver)

	card := mockCardWithObserver(ctrl, "c1")
	pileObs.EXPECT().OnCardMovedToPile(card)

	pp.OnCardMovedToPile(card)
}

func TestPlayer_OnWarriorDead(t *testing.T) {
	t.Run("Removes warrior from field and notifies cemetery", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		p, _, cemeteryObs, _, fieldObs := newTestPlayer(t, ctrl)
		pp := p.(cards.WarriorDeadObserver)

		warrior := mockWarriorCard(ctrl, "w1", types.KnightWarriorType)
		// Need a second warrior so field doesn't become empty (which triggers fieldObs)
		warrior2 := mockWarriorCard(ctrl, "w2", types.ArcherWarriorType)
		p.TakeCards(warrior, warrior2)
		p.MoveCardToField("w1")
		p.MoveCardToField("w2")

		cemeteryObs.EXPECT().OnWarriorMovedToCemetery(warrior)
		_ = fieldObs // not called because field still has warrior2

		pp.OnWarriorDead(warrior)

		_, ok := p.GetCardFromField("w1")
		assert.False(t, ok)
	})

	t.Run("Triggers field empty observer when last warrior dies", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		p, _, cemeteryObs, _, fieldObs := newTestPlayer(t, ctrl)
		pp := p.(cards.WarriorDeadObserver)

		warrior := mockWarriorCard(ctrl, "w1", types.KnightWarriorType)
		p.TakeCards(warrior)
		p.MoveCardToField("w1")

		cemeteryObs.EXPECT().OnWarriorMovedToCemetery(warrior)
		fieldObs.EXPECT().OnFieldWithoutWarriors("TestPlayer")

		pp.OnWarriorDead(warrior)
	})
}
