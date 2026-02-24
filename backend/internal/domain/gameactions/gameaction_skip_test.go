package gameactions_test

import (
	"testing"

	"github.com/alelopezbcn/thecampaign/internal/domain/gameactions"
	"github.com/alelopezbcn/thecampaign/internal/domain/gamestatus"
	"github.com/alelopezbcn/thecampaign/internal/domain/types"
	"github.com/alelopezbcn/thecampaign/test/mocks"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

func TestSkipPhaseAction_PlayerName(t *testing.T) {
	action := gameactions.NewSkipPhaseAction("Player1")
	assert.Equal(t, "Player1", action.PlayerName())
}

func TestSkipPhaseAction_Validate(t *testing.T) {
	phases := []struct {
		current types.PhaseType
		next    types.PhaseType
		wantErr bool
	}{
		{types.PhaseTypeDrawCard, "", true},
		{types.PhaseTypeEndTurn, "", true},
		{types.PhaseTypeAttack, types.PhaseTypeSpySteal, false},
		{types.PhaseTypeSpySteal, types.PhaseTypeBuy, false},
		{types.PhaseTypeBuy, types.PhaseTypeConstruct, false},
		{types.PhaseTypeConstruct, types.PhaseTypeEndTurn, false},
	}

	for _, tt := range phases {
		tt := tt
		t.Run(string(tt.current), func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockGame := mocks.NewMockGame(ctrl)
			mockGame.EXPECT().CurrentAction().Return(tt.current)

			action := gameactions.NewSkipPhaseAction("Player1")
			err := action.Validate(mockGame)

			if tt.wantErr {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), "cannot skip this phase")
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.next, action.NextPhase())
			}
		})
	}
}

func TestSkipPhaseAction_Execute(t *testing.T) {
	t.Run("Returns skip result and status function", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockGame := mocks.NewMockGame(ctrl)
		mockPlayer1 := mocks.NewMockPlayer(ctrl)
		expectedStatus := gamestatus.GameStatus{CurrentPlayer: "Player1"}

		mockGame.EXPECT().CurrentPlayer().Return(mockPlayer1)
		mockGame.EXPECT().Status(mockPlayer1).Return(expectedStatus)

		action := gameactions.NewSkipPhaseAction("Player1")
		result, statusFn, err := action.Execute(mockGame)

		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Equal(t, types.LastActionSkip, result.Action)
		assert.Equal(t, expectedStatus, statusFn())
	})
}
