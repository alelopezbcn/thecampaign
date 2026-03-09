package gameactions_test

import (
	"errors"
	"testing"

	"github.com/alelopezbcn/thecampaign/internal/domain/cards"
	"github.com/alelopezbcn/thecampaign/internal/domain/gameactions"
	"github.com/alelopezbcn/thecampaign/internal/domain/gamestatus"
	"github.com/alelopezbcn/thecampaign/internal/domain/types"
	"github.com/alelopezbcn/thecampaign/test/mocks"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

func TestForgeAction_PlayerName(t *testing.T) {
	action := gameactions.NewForgeAction("Player1", "w1", "w2")
	assert.Equal(t, "Player1", action.PlayerName())
}

func TestForgeAction_Validate(t *testing.T) {
	t.Run("Error when already forged this turn", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockGame := mocks.NewMockGame(ctrl)
		mockGame.EXPECT().TurnState().Return(types.TurnState{HasForged: true})

		action := gameactions.NewForgeAction("Player1", "w1", "w2")
		err := action.Validate(mockGame)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "already forged this turn")
	})

	t.Run("Error when card1 not in hand", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockGame := mocks.NewMockGame(ctrl)
		mockPlayer := mocks.NewMockPlayer(ctrl)
		mockGame.EXPECT().TurnState().Return(types.TurnState{})
		mockGame.EXPECT().CurrentPlayer().Return(mockPlayer)
		mockPlayer.EXPECT().GetCardFromHand("w1").Return(nil, false)

		action := gameactions.NewForgeAction("Player1", "w1", "w2")
		err := action.Validate(mockGame)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "w1")
	})

	t.Run("Error when card2 not in hand", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockGame := mocks.NewMockGame(ctrl)
		mockPlayer := mocks.NewMockPlayer(ctrl)
		mockWeapon1 := mocks.NewMockWeapon(ctrl)
		mockGame.EXPECT().TurnState().Return(types.TurnState{})
		mockGame.EXPECT().CurrentPlayer().Return(mockPlayer)
		mockPlayer.EXPECT().GetCardFromHand("w1").Return(mockWeapon1, true)
		mockPlayer.EXPECT().GetCardFromHand("w2").Return(nil, false)

		action := gameactions.NewForgeAction("Player1", "w1", "w2")
		err := action.Validate(mockGame)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "w2")
	})

	t.Run("Error when card1 is not a weapon", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockGame := mocks.NewMockGame(ctrl)
		mockPlayer := mocks.NewMockPlayer(ctrl)
		mockCard := mocks.NewMockCard(ctrl)
		mockGame.EXPECT().TurnState().Return(types.TurnState{})
		mockGame.EXPECT().CurrentPlayer().Return(mockPlayer)
		mockPlayer.EXPECT().GetCardFromHand("w1").Return(mockCard, true)
		mockPlayer.EXPECT().GetCardFromHand("w2").Return(mockCard, true)

		action := gameactions.NewForgeAction("Player1", "w1", "w2")
		err := action.Validate(mockGame)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "not a weapon")
	})

	t.Run("Error when weapons are different types", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockGame := mocks.NewMockGame(ctrl)
		mockPlayer := mocks.NewMockPlayer(ctrl)
		mockSword := mocks.NewMockWeapon(ctrl)
		mockArrow := mocks.NewMockWeapon(ctrl)
		mockGame.EXPECT().TurnState().Return(types.TurnState{})
		mockGame.EXPECT().CurrentPlayer().Return(mockPlayer)
		mockPlayer.EXPECT().GetCardFromHand("w1").Return(mockSword, true)
		mockPlayer.EXPECT().GetCardFromHand("w2").Return(mockArrow, true)
		mockSword.EXPECT().Type().Return(types.SwordWeaponType).AnyTimes()
		mockArrow.EXPECT().Type().Return(types.ArrowWeaponType).AnyTimes()

		action := gameactions.NewForgeAction("Player1", "w1", "w2")
		err := action.Validate(mockGame)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "cannot forge different weapon types")
	})

	t.Run("Error when weapon type is not forgeable (harpoon)", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockGame := mocks.NewMockGame(ctrl)
		mockPlayer := mocks.NewMockPlayer(ctrl)
		mockHarpoon1 := mocks.NewMockWeapon(ctrl)
		mockHarpoon2 := mocks.NewMockWeapon(ctrl)
		mockGame.EXPECT().TurnState().Return(types.TurnState{})
		mockGame.EXPECT().CurrentPlayer().Return(mockPlayer)
		mockPlayer.EXPECT().GetCardFromHand("h1").Return(mockHarpoon1, true)
		mockPlayer.EXPECT().GetCardFromHand("h2").Return(mockHarpoon2, true)
		mockHarpoon1.EXPECT().Type().Return(types.HarpoonWeaponType).AnyTimes()
		mockHarpoon2.EXPECT().Type().Return(types.HarpoonWeaponType).AnyTimes()

		action := gameactions.NewForgeAction("Player1", "h1", "h2")
		err := action.Validate(mockGame)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "cannot be forged")
	})

	t.Run("Success with 2 swords", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockGame := mocks.NewMockGame(ctrl)
		mockPlayer := mocks.NewMockPlayer(ctrl)
		mockSword1 := mocks.NewMockWeapon(ctrl)
		mockSword2 := mocks.NewMockWeapon(ctrl)
		mockGame.EXPECT().TurnState().Return(types.TurnState{})
		mockGame.EXPECT().CurrentPlayer().Return(mockPlayer)
		mockPlayer.EXPECT().GetCardFromHand("w1").Return(mockSword1, true)
		mockPlayer.EXPECT().GetCardFromHand("w2").Return(mockSword2, true)
		mockSword1.EXPECT().Type().Return(types.SwordWeaponType).AnyTimes()
		mockSword2.EXPECT().Type().Return(types.SwordWeaponType).AnyTimes()

		action := gameactions.NewForgeAction("Player1", "w1", "w2")
		err := action.Validate(mockGame)

		assert.NoError(t, err)
	})
}

func TestForgeAction_Execute(t *testing.T) {
	t.Run("Error when RemoveFromHand fails", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockGame := mocks.NewMockGame(ctrl)
		mockPlayer := mocks.NewMockPlayer(ctrl)
		mockSword1 := mocks.NewMockWeapon(ctrl)
		mockSword2 := mocks.NewMockWeapon(ctrl)

		// Setup Validate first
		mockGame.EXPECT().TurnState().Return(types.TurnState{})
		mockGame.EXPECT().CurrentPlayer().Return(mockPlayer).AnyTimes()
		mockPlayer.EXPECT().GetCardFromHand("w1").Return(mockSword1, true)
		mockPlayer.EXPECT().GetCardFromHand("w2").Return(mockSword2, true)
		mockSword1.EXPECT().Type().Return(types.SwordWeaponType).AnyTimes()
		mockSword2.EXPECT().Type().Return(types.SwordWeaponType).AnyTimes()

		mockPlayer.EXPECT().RemoveFromHand("w1", "w2").Return(nil, errors.New("card not found"))

		action := gameactions.NewForgeAction("Player1", "w1", "w2")
		_ = action.Validate(mockGame)
		result, _, err := action.Execute(mockGame)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "removing weapons for forging failed")
		assert.NotNil(t, result)
	})

	t.Run("Success forging 2 swords into 1", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockGame := mocks.NewMockGame(ctrl)
		mockPlayer := mocks.NewMockPlayer(ctrl)
		mockSword1 := mocks.NewMockWeapon(ctrl)
		mockSword2 := mocks.NewMockWeapon(ctrl)
		expectedStatus := gamestatus.GameStatus{CurrentPlayer: "Player1"}

		// Setup Validate
		mockGame.EXPECT().TurnState().Return(types.TurnState{})
		mockGame.EXPECT().CurrentAction().Return(types.PhaseTypeBuy).AnyTimes()
		mockGame.EXPECT().CurrentPlayer().Return(mockPlayer).AnyTimes()
		mockPlayer.EXPECT().GetCardFromHand("w1").Return(mockSword1, true)
		mockPlayer.EXPECT().GetCardFromHand("w2").Return(mockSword2, true)
		mockSword1.EXPECT().Type().Return(types.SwordWeaponType).AnyTimes()
		mockSword2.EXPECT().Type().Return(types.SwordWeaponType).AnyTimes()

		// Setup Execute
		mockPlayer.EXPECT().RemoveFromHand("w1", "w2").Return(
			[]cards.Card{mockSword1, mockSword2}, nil)
		mockSword1.EXPECT().DamageAmount().Return(3)
		mockSword2.EXPECT().DamageAmount().Return(4)
		mockPlayer.EXPECT().Name().Return("Player1")
		mockPlayer.EXPECT().TakeCards(gomock.Any())
		mockGame.EXPECT().AddHistory(gomock.Any(), gomock.Any())
		mockGame.EXPECT().SetHasForged(true)
		mockGame.EXPECT().SetCanForge(false)
		mockGame.EXPECT().Status(mockPlayer, gomock.Any()).Return(expectedStatus)

		action := gameactions.NewForgeAction("Player1", "w1", "w2")
		_ = action.Validate(mockGame)
		result, statusFn, err := action.Execute(mockGame)

		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Equal(t, types.LastActionForge, result.Action)
		assert.Equal(t, types.PhaseTypeBuy, action.NextPhase())
		assert.Equal(t, expectedStatus, statusFn())
	})

	t.Run("History updated with forged weapon info", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockGame := mocks.NewMockGame(ctrl)
		mockPlayer := mocks.NewMockPlayer(ctrl)
		mockSword1 := mocks.NewMockWeapon(ctrl)
		mockSword2 := mocks.NewMockWeapon(ctrl)

		// Setup Validate
		mockGame.EXPECT().TurnState().Return(types.TurnState{})
		mockGame.EXPECT().CurrentAction().Return(types.PhaseTypeBuy).AnyTimes()
		mockGame.EXPECT().CurrentPlayer().Return(mockPlayer).AnyTimes()
		mockPlayer.EXPECT().GetCardFromHand("w1").Return(mockSword1, true)
		mockPlayer.EXPECT().GetCardFromHand("w2").Return(mockSword2, true)
		mockSword1.EXPECT().Type().Return(types.SwordWeaponType).AnyTimes()
		mockSword2.EXPECT().Type().Return(types.SwordWeaponType).AnyTimes()

		// Setup Execute
		mockPlayer.EXPECT().RemoveFromHand("w1", "w2").Return(
			[]cards.Card{mockSword1, mockSword2}, nil)
		mockSword1.EXPECT().DamageAmount().Return(3)
		mockSword2.EXPECT().DamageAmount().Return(4)
		mockPlayer.EXPECT().Name().Return("Player1")
		mockPlayer.EXPECT().TakeCards(gomock.Any())

		var capturedMsg string
		mockGame.EXPECT().AddHistory(gomock.Any(), gomock.Any()).Do(func(msg string, _ types.Category) {
			capturedMsg = msg
		})
		mockGame.EXPECT().SetHasForged(true)
		mockGame.EXPECT().SetCanForge(false)

		action := gameactions.NewForgeAction("Player1", "w1", "w2")
		_ = action.Validate(mockGame)
		_, statusFn, err := action.Execute(mockGame)

		assert.NoError(t, err)
		assert.NotNil(t, statusFn)
		assert.Contains(t, capturedMsg, "Player1")
		assert.Contains(t, capturedMsg, "forged")
		assert.Contains(t, capturedMsg, "7")
	})

	t.Run("Discarding forged weapon unforges into original components", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockGame := mocks.NewMockGame(ctrl)
		mockPlayer := mocks.NewMockPlayer(ctrl)
		mockSword1 := mocks.NewMockWeapon(ctrl)
		mockSword2 := mocks.NewMockWeapon(ctrl)
		mockObs1 := mocks.NewMockCardMovedToPileObserver(ctrl)
		mockObs2 := mocks.NewMockCardMovedToPileObserver(ctrl)

		// Setup Validate
		mockGame.EXPECT().TurnState().Return(types.TurnState{})
		mockGame.EXPECT().CurrentAction().Return(types.PhaseTypeBuy).AnyTimes()
		mockGame.EXPECT().CurrentPlayer().Return(mockPlayer).AnyTimes()
		mockPlayer.EXPECT().GetCardFromHand("w1").Return(mockSword1, true)
		mockPlayer.EXPECT().GetCardFromHand("w2").Return(mockSword2, true)
		mockSword1.EXPECT().Type().Return(types.SwordWeaponType).AnyTimes()
		mockSword2.EXPECT().Type().Return(types.SwordWeaponType).AnyTimes()

		// Setup Execute
		mockPlayer.EXPECT().RemoveFromHand("w1", "w2").Return(
			[]cards.Card{mockSword1, mockSword2}, nil)
		mockSword1.EXPECT().DamageAmount().Return(3)
		mockSword2.EXPECT().DamageAmount().Return(4)
		mockPlayer.EXPECT().Name().Return("Player1")
		mockGame.EXPECT().AddHistory(gomock.Any(), gomock.Any())
		mockGame.EXPECT().SetHasForged(true)
		mockGame.EXPECT().SetCanForge(false)

		// Capture the forged weapon passed to TakeCards
		var capturedWeapon cards.Card
		mockPlayer.EXPECT().TakeCards(gomock.Any()).Do(func(cs ...cards.Card) {
			if len(cs) > 0 {
				capturedWeapon = cs[0]
			}
		})

		action := gameactions.NewForgeAction("Player1", "w1", "w2")
		_ = action.Validate(mockGame)
		_, _, err := action.Execute(mockGame)
		assert.NoError(t, err)
		assert.NotNil(t, capturedWeapon)

		// When the forged weapon is discarded, each original component must be
		// discarded via its own observer — not the forged weapon itself.
		mockSword1.EXPECT().GetCardMovedToPileObserver().Return(mockObs1)
		mockSword2.EXPECT().GetCardMovedToPileObserver().Return(mockObs2)
		mockObs1.EXPECT().OnCardMovedToPile(mockSword1)
		mockObs2.EXPECT().OnCardMovedToPile(mockSword2)

		capturedWeapon.GetCardMovedToPileObserver().OnCardMovedToPile(capturedWeapon)
	})
}
