package game

import (
	"errors"
	"strings"
	"testing"

	"github.com/alelopezbcn/thecampaign/internal/domain/board"
	"github.com/alelopezbcn/thecampaign/internal/domain/gamestatus"
	"github.com/alelopezbcn/thecampaign/internal/domain/types"
	"github.com/alelopezbcn/thecampaign/test/mocks"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

func TestSpecialPowerAction_PlayerName(t *testing.T) {
	action := NewSpecialPowerAction("Player1", "K1", "EK1", "SP1")
	assert.Equal(t, "Player1", action.PlayerName())
}

func TestSpecialPowerAction_NextPhase(t *testing.T) {
	action := NewSpecialPowerAction("Player1", "K1", "EK1", "SP1")
	assert.Equal(t, types.PhaseTypeSpySteal, action.NextPhase())
}

func TestSpecialPowerAction_Validate(t *testing.T) {
	t.Run("Error when not in Attack phase", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockPlayer1 := mocks.NewMockPlayer(ctrl)
		mockPlayer2 := mocks.NewMockPlayer(ctrl)

		mockPlayer1.EXPECT().Name().Return("Player1").AnyTimes()

		g := &Game{
			players:       []board.Player{mockPlayer1, mockPlayer2},
			CurrentTurn:   0,
			currentAction: types.PhaseTypeBuy,
		}

		action := NewSpecialPowerAction("Player1", "K1", "EK1", "SP1")
		err := action.Validate(g)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "cannot use special power in the")
	})

	t.Run("Error when warrior not in field", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockPlayer1 := mocks.NewMockPlayer(ctrl)
		mockPlayer2 := mocks.NewMockPlayer(ctrl)

		mockPlayer1.EXPECT().Name().Return("Player1").AnyTimes()
		mockPlayer1.EXPECT().GetCardFromField("K1").Return(nil, false)

		g := &Game{
			players:       []board.Player{mockPlayer1, mockPlayer2},
			CurrentTurn:   0,
			currentAction: types.PhaseTypeAttack,
		}

		action := NewSpecialPowerAction("Player1", "K1", "EK1", "SP1")
		err := action.Validate(g)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "warrior card not in field")
	})

	t.Run("Error when user card is not a warrior", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockPlayer1 := mocks.NewMockPlayer(ctrl)
		mockPlayer2 := mocks.NewMockPlayer(ctrl)
		mockCard := mocks.NewMockCard(ctrl) // Not a Warrior

		mockPlayer1.EXPECT().Name().Return("Player1").AnyTimes()
		mockPlayer1.EXPECT().GetCardFromField("K1").Return(mockCard, true)

		g := &Game{
			players:       []board.Player{mockPlayer1, mockPlayer2},
			CurrentTurn:   0,
			currentAction: types.PhaseTypeAttack,
		}

		action := NewSpecialPowerAction("Player1", "K1", "EK1", "SP1")
		err := action.Validate(g)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "the attacking card is not a warrior")
	})

	t.Run("Error when target not in any field", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockPlayer1 := mocks.NewMockPlayer(ctrl)
		mockPlayer2 := mocks.NewMockPlayer(ctrl)
		mockWarrior := mocks.NewMockWarrior(ctrl)

		mockPlayer1.EXPECT().Name().Return("Player1").AnyTimes()
		mockPlayer1.EXPECT().GetCardFromField("K1").Return(mockWarrior, true)
		mockWarrior.EXPECT().Type().Return(types.ArcherWarriorType)
		mockPlayer1.EXPECT().GetCardFromField("EK1").Return(nil, false)
		mockPlayer2.EXPECT().GetCardFromField("EK1").Return(nil, false)

		g := &Game{
			players:       []board.Player{mockPlayer1, mockPlayer2},
			CurrentTurn:   0,
			currentAction: types.PhaseTypeAttack,
		}

		action := NewSpecialPowerAction("Player1", "K1", "EK1", "SP1")
		err := action.Validate(g)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "target card not valid")
	})

	t.Run("Error when archer targets ally", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockPlayer1 := mocks.NewMockPlayer(ctrl)
		mockPlayer2 := mocks.NewMockPlayer(ctrl)
		mockWarrior := mocks.NewMockWarrior(ctrl)
		mockTarget := mocks.NewMockWarrior(ctrl)

		mockPlayer1.EXPECT().Name().Return("Player1").AnyTimes()
		mockPlayer1.EXPECT().GetCardFromField("A1").Return(mockWarrior, true)
		mockWarrior.EXPECT().Type().Return(types.ArcherWarriorType)
		// Target found in own field (ally)
		mockPlayer1.EXPECT().GetCardFromField("T1").Return(mockTarget, true)

		g := &Game{
			players:       []board.Player{mockPlayer1, mockPlayer2},
			CurrentTurn:   0,
			currentAction: types.PhaseTypeAttack,
		}

		action := NewSpecialPowerAction("Player1", "A1", "T1", "SP1")
		err := action.Validate(g)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "archer instant kill can only target enemies")
	})

	t.Run("Error when knight targets enemy", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockPlayer1 := mocks.NewMockPlayer(ctrl)
		mockPlayer2 := mocks.NewMockPlayer(ctrl)
		mockWarrior := mocks.NewMockWarrior(ctrl)
		mockTarget := mocks.NewMockWarrior(ctrl)

		mockPlayer1.EXPECT().Name().Return("Player1").AnyTimes()
		mockPlayer1.EXPECT().GetCardFromField("K1").Return(mockWarrior, true)
		mockWarrior.EXPECT().Type().Return(types.KnightWarriorType)
		// Target not in own field
		mockPlayer1.EXPECT().GetCardFromField("EK1").Return(nil, false)
		// Target found in enemy field
		mockPlayer2.EXPECT().GetCardFromField("EK1").Return(mockTarget, true)

		g := &Game{
			players:       []board.Player{mockPlayer1, mockPlayer2},
			CurrentTurn:   0,
			currentAction: types.PhaseTypeAttack,
		}

		action := NewSpecialPowerAction("Player1", "K1", "EK1", "SP1")
		err := action.Validate(g)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "knight/mage special power can only target allies")
	})

	t.Run("Error when weapon not in hand", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockPlayer1 := mocks.NewMockPlayer(ctrl)
		mockPlayer2 := mocks.NewMockPlayer(ctrl)
		mockWarrior := mocks.NewMockWarrior(ctrl)
		mockTarget := mocks.NewMockWarrior(ctrl)

		mockPlayer1.EXPECT().Name().Return("Player1").AnyTimes()
		mockPlayer1.EXPECT().GetCardFromField("K1").Return(mockWarrior, true)
		mockWarrior.EXPECT().Type().Return(types.ArcherWarriorType)
		mockPlayer1.EXPECT().GetCardFromField("EK1").Return(nil, false)
		mockPlayer2.EXPECT().GetCardFromField("EK1").Return(mockTarget, true)
		mockPlayer1.EXPECT().GetCardFromHand("SP1").Return(nil, false)

		g := &Game{
			players:       []board.Player{mockPlayer1, mockPlayer2},
			CurrentTurn:   0,
			currentAction: types.PhaseTypeAttack,
		}

		action := NewSpecialPowerAction("Player1", "K1", "EK1", "SP1")
		err := action.Validate(g)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "weapon card not in hand")
	})

	t.Run("Error when card is not a special power", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockPlayer1 := mocks.NewMockPlayer(ctrl)
		mockPlayer2 := mocks.NewMockPlayer(ctrl)
		mockWarrior := mocks.NewMockWarrior(ctrl)
		mockTarget := mocks.NewMockWarrior(ctrl)
		mockResource := mocks.NewMockResource(ctrl)

		mockPlayer1.EXPECT().Name().Return("Player1").AnyTimes()
		mockPlayer1.EXPECT().GetCardFromField("K1").Return(mockWarrior, true)
		mockWarrior.EXPECT().Type().Return(types.ArcherWarriorType)
		mockPlayer1.EXPECT().GetCardFromField("EK1").Return(nil, false)
		mockPlayer2.EXPECT().GetCardFromField("EK1").Return(mockTarget, true)
		mockPlayer1.EXPECT().GetCardFromHand("SP1").Return(mockResource, true)

		g := &Game{
			players:       []board.Player{mockPlayer1, mockPlayer2},
			CurrentTurn:   0,
			currentAction: types.PhaseTypeAttack,
		}

		action := NewSpecialPowerAction("Player1", "K1", "EK1", "SP1")
		err := action.Validate(g)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "the card is not a special power")
	})

	t.Run("Error when target card is not a warrior", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockPlayer1 := mocks.NewMockPlayer(ctrl)
		mockPlayer2 := mocks.NewMockPlayer(ctrl)
		mockWarrior := mocks.NewMockWarrior(ctrl)
		mockTargetCard := mocks.NewMockCard(ctrl) // Not a Warrior
		mockSP := mocks.NewMockSpecialPower(ctrl)

		mockPlayer1.EXPECT().Name().Return("Player1").AnyTimes()
		mockPlayer1.EXPECT().GetCardFromField("A1").Return(mockWarrior, true)
		mockWarrior.EXPECT().Type().Return(types.ArcherWarriorType)
		mockPlayer1.EXPECT().GetCardFromField("EK1").Return(nil, false)
		mockPlayer2.EXPECT().GetCardFromField("EK1").Return(mockTargetCard, true)
		mockPlayer1.EXPECT().GetCardFromHand("SP1").Return(mockSP, true)

		g := &Game{
			players:       []board.Player{mockPlayer1, mockPlayer2},
			CurrentTurn:   0,
			currentAction: types.PhaseTypeAttack,
		}

		action := NewSpecialPowerAction("Player1", "A1", "EK1", "SP1")
		err := action.Validate(g)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "the target card is not a warrior")
	})

	t.Run("Success on enemy target stores fields", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockPlayer1 := mocks.NewMockPlayer(ctrl)
		mockPlayer2 := mocks.NewMockPlayer(ctrl)
		mockWarrior := mocks.NewMockWarrior(ctrl)
		mockTarget := mocks.NewMockWarrior(ctrl)
		mockSP := mocks.NewMockSpecialPower(ctrl)

		mockPlayer1.EXPECT().Name().Return("Player1").AnyTimes()
		mockPlayer1.EXPECT().GetCardFromField("A1").Return(mockWarrior, true)
		mockWarrior.EXPECT().Type().Return(types.ArcherWarriorType)
		mockPlayer1.EXPECT().GetCardFromField("EK1").Return(nil, false)
		mockPlayer2.EXPECT().GetCardFromField("EK1").Return(mockTarget, true)
		mockPlayer1.EXPECT().GetCardFromHand("SP1").Return(mockSP, true)

		g := &Game{
			players:       []board.Player{mockPlayer1, mockPlayer2},
			CurrentTurn:   0,
			currentAction: types.PhaseTypeAttack,
		}

		action := NewSpecialPowerAction("Player1", "A1", "EK1", "SP1")
		err := action.Validate(g)

		assert.NoError(t, err)
		assert.Equal(t, mockWarrior, action.usedBy)
		assert.Equal(t, mockTarget, action.usedOn)
		assert.Equal(t, mockSP, action.specialPower)
	})

	t.Run("Success on own target (knight protect/heal)", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockPlayer1 := mocks.NewMockPlayer(ctrl)
		mockPlayer2 := mocks.NewMockPlayer(ctrl)
		mockWarrior := mocks.NewMockWarrior(ctrl)
		mockTarget := mocks.NewMockWarrior(ctrl)
		mockSP := mocks.NewMockSpecialPower(ctrl)

		mockPlayer1.EXPECT().Name().Return("Player1").AnyTimes()
		mockPlayer1.EXPECT().GetCardFromField("K1").Return(mockWarrior, true)
		mockWarrior.EXPECT().Type().Return(types.KnightWarriorType)
		// Target found in own field
		mockPlayer1.EXPECT().GetCardFromField("A1").Return(mockTarget, true)
		mockPlayer1.EXPECT().GetCardFromHand("SP1").Return(mockSP, true)

		g := &Game{
			players:       []board.Player{mockPlayer1, mockPlayer2},
			CurrentTurn:   0,
			currentAction: types.PhaseTypeAttack,
		}

		action := NewSpecialPowerAction("Player1", "K1", "A1", "SP1")
		err := action.Validate(g)

		assert.NoError(t, err)
		assert.Equal(t, mockWarrior, action.usedBy)
		assert.Equal(t, mockTarget, action.usedOn)
		assert.Equal(t, mockSP, action.specialPower)
	})
}

func TestSpecialPowerAction_Execute(t *testing.T) {
	t.Run("Error when special power fails", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockPlayer1 := mocks.NewMockPlayer(ctrl)
		mockPlayer2 := mocks.NewMockPlayer(ctrl)
		mockWarrior := mocks.NewMockWarrior(ctrl)
		mockTarget := mocks.NewMockWarrior(ctrl)
		mockSP := mocks.NewMockSpecialPower(ctrl)

		mockPlayer1.EXPECT().Name().Return("Player1").AnyTimes()
		mockPlayer1.EXPECT().UseSpecialPower(mockWarrior, mockTarget, mockSP).Return(errors.New("power failed"))

		g := &Game{
			players:       []board.Player{mockPlayer1, mockPlayer2},
			CurrentTurn:   0,
			currentAction: types.PhaseTypeAttack,
		}

		action := NewSpecialPowerAction("Player1", "K1", "EK1", "SP1")
		action.usedBy = mockWarrior
		action.usedOn = mockTarget
		action.specialPower = mockSP

		result, _, err := action.Execute(g)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "special power action failed")
		assert.NotNil(t, result)
	})

	t.Run("Success returns result", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockPlayer1 := mocks.NewMockPlayer(ctrl)
		mockPlayer2 := mocks.NewMockPlayer(ctrl)
		mockProvider := NewMockGameStatusProvider(ctrl)
		mockWarrior := mocks.NewMockWarrior(ctrl)
		mockTarget := mocks.NewMockWarrior(ctrl)
		mockSP := mocks.NewMockSpecialPower(ctrl)

		expectedStatus := gamestatus.GameStatus{CurrentPlayer: "Player1"}

		mockPlayer1.EXPECT().Name().Return("Player1").AnyTimes()
		mockPlayer1.EXPECT().UseSpecialPower(mockWarrior, mockTarget, mockSP).Return(nil)
		mockTarget.EXPECT().String().Return("Knight (20)")

		g := &Game{
			players:            []board.Player{mockPlayer1, mockPlayer2},
			CurrentTurn:        0,
			currentAction:      types.PhaseTypeAttack,
			gameStatusProvider: mockProvider,
			history:            []types.HistoryLine{},
		}

		mockProvider.EXPECT().Get(mockPlayer1, g).Return(expectedStatus)

		action := NewSpecialPowerAction("Player1", "K1", "EK1", "SP1")
		action.usedBy = mockWarrior
		action.usedOn = mockTarget
		action.specialPower = mockSP

		result, statusFn, err := action.Execute(g)

		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Equal(t, types.LastActionSpecialPower, result.Action)
		assert.Equal(t, expectedStatus, statusFn())
	})

	t.Run("History is updated on success", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockPlayer1 := mocks.NewMockPlayer(ctrl)
		mockPlayer2 := mocks.NewMockPlayer(ctrl)
		mockWarrior := mocks.NewMockWarrior(ctrl)
		mockTarget := mocks.NewMockWarrior(ctrl)
		mockSP := mocks.NewMockSpecialPower(ctrl)

		mockPlayer1.EXPECT().Name().Return("Player1").AnyTimes()
		mockPlayer1.EXPECT().UseSpecialPower(mockWarrior, mockTarget, mockSP).Return(nil)
		mockTarget.EXPECT().String().Return("Knight (20)")

		g := &Game{
			players:       []board.Player{mockPlayer1, mockPlayer2},
			CurrentTurn:   0,
			currentAction: types.PhaseTypeAttack,
			history:       []types.HistoryLine{},
		}

		action := NewSpecialPowerAction("Player1", "K1", "EK1", "SP1")
		action.usedBy = mockWarrior
		action.usedOn = mockTarget
		action.specialPower = mockSP

		_, statusFn, err := action.Execute(g)

		assert.NoError(t, err)
		assert.NotNil(t, statusFn)
		found := false
		for _, h := range g.history {
			if strings.Contains(h.Msg, "Player1") && strings.Contains(h.Msg, "special power") {
				found = true
				break
			}
		}
		assert.True(t, found, "History should contain special power action")
	})
}
