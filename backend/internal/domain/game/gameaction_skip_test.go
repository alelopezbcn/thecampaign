package game

import (
	"testing"

	"github.com/alelopezbcn/thecampaign/internal/domain/gamestatus"
	"github.com/alelopezbcn/thecampaign/internal/domain/ports"
	"github.com/alelopezbcn/thecampaign/internal/domain/types"
	"github.com/alelopezbcn/thecampaign/test/mocks"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

func TestSkipPhaseAction_PlayerName(t *testing.T) {
	action := NewSkipPhaseAction("Player1")
	assert.Equal(t, "Player1", action.PlayerName())
}

func TestSkipPhaseAction_Validate(t *testing.T) {
	t.Run("Error when trying to skip DrawCard phase", func(t *testing.T) {
		g := &Game{currentAction: types.PhaseTypeDrawCard}
		action := NewSkipPhaseAction("Player1")

		err := action.Validate(g)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "cannot skip this phase")
	})

	t.Run("Error when trying to skip EndTurn phase", func(t *testing.T) {
		g := &Game{currentAction: types.PhaseTypeEndTurn}
		action := NewSkipPhaseAction("Player1")

		err := action.Validate(g)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "cannot skip this phase")
	})

	t.Run("Attack phase sets next phase to SpySteal", func(t *testing.T) {
		g := &Game{currentAction: types.PhaseTypeAttack}
		action := NewSkipPhaseAction("Player1")

		err := action.Validate(g)

		assert.NoError(t, err)
		assert.Equal(t, types.PhaseTypeSpySteal, action.NextPhase())
	})

	t.Run("SpySteal phase sets next phase to Buy", func(t *testing.T) {
		g := &Game{currentAction: types.PhaseTypeSpySteal}
		action := NewSkipPhaseAction("Player1")

		err := action.Validate(g)

		assert.NoError(t, err)
		assert.Equal(t, types.PhaseTypeBuy, action.NextPhase())
	})

	t.Run("Buy phase sets next phase to Construct", func(t *testing.T) {
		g := &Game{currentAction: types.PhaseTypeBuy}
		action := NewSkipPhaseAction("Player1")

		err := action.Validate(g)

		assert.NoError(t, err)
		assert.Equal(t, types.PhaseTypeConstruct, action.NextPhase())
	})

	t.Run("Construct phase sets next phase to EndTurn", func(t *testing.T) {
		g := &Game{currentAction: types.PhaseTypeConstruct}
		action := NewSkipPhaseAction("Player1")

		err := action.Validate(g)

		assert.NoError(t, err)
		assert.Equal(t, types.PhaseTypeEndTurn, action.NextPhase())
	})
}

func TestSkipPhaseAction_Execute(t *testing.T) {
	t.Run("Returns skip result and status function", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockPlayer1 := mocks.NewMockPlayer(ctrl)
		mockPlayer2 := mocks.NewMockPlayer(ctrl)
		mockProvider := NewMockGameStatusProvider(ctrl)
		mockDiscardPile := mocks.NewMockDiscardPile(ctrl)

		expectedStatus := gamestatus.GameStatus{CurrentPlayer: "Player1"}

		mockPlayer1.EXPECT().Name().Return("Player1").AnyTimes()

		g := &Game{
			Players:            []ports.Player{mockPlayer1, mockPlayer2},
			CurrentTurn:        0,
			currentAction:      types.PhaseTypeAttack,
			discardPile:        mockDiscardPile,
			gameStatusProvider: mockProvider,
		}

		mockProvider.EXPECT().Get(mockPlayer1, g).Return(expectedStatus)

		action := NewSkipPhaseAction("Player1")
		result, statusFn, err := action.Execute(g)

		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Equal(t, types.LastActionSkip, result.Action)
		assert.Equal(t, expectedStatus, statusFn())
	})
}
