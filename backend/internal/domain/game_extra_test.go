package domain

import (
	"testing"

	"github.com/alelopezbcn/thecampaign/internal/domain/board"
	"github.com/alelopezbcn/thecampaign/internal/domain/cards"
	"github.com/alelopezbcn/thecampaign/internal/domain/gameactions"
	"github.com/alelopezbcn/thecampaign/internal/domain/gamestatus"
	"github.com/alelopezbcn/thecampaign/internal/domain/types"
	"github.com/alelopezbcn/thecampaign/test/mocks"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

// ──────────────────────────────────────────────────────────────────────────────
// availableEventTypes
// ──────────────────────────────────────────────────────────────────────────────

func TestGame_AvailableEventTypes_FFAIncludesChampionsBounty(t *testing.T) {
	for _, mode := range []types.GameMode{types.GameModeFFA3, types.GameModeFFA5} {
		g := &game{mode: mode}
		available := g.availableEventTypes()
		found := false
		for _, et := range available {
			if et == types.EventTypeChampionsBounty {
				found = true
			}
		}
		assert.True(t, found, "FFA mode %s should include ChampionsBounty", mode)
	}
}

func TestGame_AvailableEventTypes_NonFFAExcludesChampionsBounty(t *testing.T) {
	for _, mode := range []types.GameMode{types.GameMode1v1, types.GameMode2v2} {
		g := &game{mode: mode}
		available := g.availableEventTypes()
		for _, et := range available {
			assert.NotEqual(t, types.EventTypeChampionsBounty, et,
				"non-FFA mode %s must not include ChampionsBounty", mode)
		}
	}
}

// ──────────────────────────────────────────────────────────────────────────────
// TurnState setters
// ──────────────────────────────────────────────────────────────────────────────

func TestGame_TurnStateSetters(t *testing.T) {
	g := &game{}

	g.SetHasMovedWarrior(true)
	assert.True(t, g.turnState.HasMovedWarrior)
	g.SetHasMovedWarrior(false)
	assert.False(t, g.turnState.HasMovedWarrior)

	g.SetHasTraded(true)
	assert.True(t, g.turnState.HasTraded)

	g.SetCanMoveWarrior(true)
	assert.True(t, g.turnState.CanMoveWarrior)

	g.SetCanTrade(true)
	assert.True(t, g.turnState.CanTrade)

	g.SetHasForged(true)
	assert.True(t, g.turnState.HasForged)

	g.SetCanForge(true)
	assert.True(t, g.turnState.CanForge)
}

// ──────────────────────────────────────────────────────────────────────────────
// DrawCards
// ──────────────────────────────────────────────────────────────────────────────

func TestGame_DrawCards_HandLimitExceeded(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockPlayer := mocks.NewMockPlayer(ctrl)
	mockPlayer.EXPECT().CanTakeCards(2).Return(false)

	g := &game{board: &testBoardImpl{}}

	result, err := g.DrawCards(mockPlayer, 2)
	assert.ErrorIs(t, err, board.ErrHandLimitExceeded)
	assert.Nil(t, result)
}

// ──────────────────────────────────────────────────────────────────────────────
// SwitchTurn — round boundary draws a new event
// ──────────────────────────────────────────────────────────────────────────────

func TestGame_SwitchTurn_DrawsNewEventOnRoundBoundary(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockPlayer1 := mocks.NewMockPlayer(ctrl)
	mockPlayer2 := mocks.NewMockPlayer(ctrl)

	// 2-player game: turn 1 -> 0 wraps around (round boundary)
	g := &game{
		board:             &testBoardImpl{players: []board.Player{mockPlayer1, mockPlayer2}},
		currentTurn:       1,
		eliminatedPlayers: make(map[int]bool),
		history:           []types.HistoryLine{},
	}

	g.SwitchTurn()

	assert.Equal(t, 0, g.currentTurn)
	found := false
	for _, h := range g.history {
		if len(h.Msg) >= 9 && h.Msg[:9] == "New round" {
			found = true
		}
	}
	assert.True(t, found, "SwitchTurn round boundary should add New round history entry")
}

// ──────────────────────────────────────────────────────────────────────────────
// nextActiveTurnPlayer
// ──────────────────────────────────────────────────────────────────────────────

func TestGame_NextActiveTurnPlayer(t *testing.T) {
	t.Run("Returns next active player", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockPlayer1 := mocks.NewMockPlayer(ctrl)
		mockPlayer2 := mocks.NewMockPlayer(ctrl)
		mockPlayer1.EXPECT().Name().Return("Player1").AnyTimes()
		mockPlayer2.EXPECT().Name().Return("Player2").AnyTimes()

		g := &game{
			board:               &testBoardImpl{players: []board.Player{mockPlayer1, mockPlayer2}},
			currentTurn:         0,
			eliminatedPlayers:   make(map[int]bool),
			disconnectedPlayers: make(map[int]bool),
		}

		assert.Equal(t, "Player2", g.nextActiveTurnPlayer())
	})

	t.Run("Skips eliminated player", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockPlayer1 := mocks.NewMockPlayer(ctrl)
		mockPlayer2 := mocks.NewMockPlayer(ctrl)
		mockPlayer3 := mocks.NewMockPlayer(ctrl)
		mockPlayer1.EXPECT().Name().Return("Player1").AnyTimes()
		mockPlayer2.EXPECT().Name().Return("Player2").AnyTimes()
		mockPlayer3.EXPECT().Name().Return("Player3").AnyTimes()

		g := &game{
			board:               &testBoardImpl{players: []board.Player{mockPlayer1, mockPlayer2, mockPlayer3}},
			currentTurn:         0,
			eliminatedPlayers:   map[int]bool{1: true},
			disconnectedPlayers: make(map[int]bool),
		}

		assert.Equal(t, "Player3", g.nextActiveTurnPlayer())
	})

	t.Run("Returns current player when all others are out", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockPlayer1 := mocks.NewMockPlayer(ctrl)
		mockPlayer2 := mocks.NewMockPlayer(ctrl)
		mockPlayer1.EXPECT().Name().Return("Player1").AnyTimes()
		mockPlayer2.EXPECT().Name().Return("Player2").AnyTimes()

		// Player2 eliminated — wraps back to Player1 (current)
		g := &game{
			board:               &testBoardImpl{players: []board.Player{mockPlayer1, mockPlayer2}},
			currentTurn:         0,
			eliminatedPlayers:   map[int]bool{1: true},
			disconnectedPlayers: make(map[int]bool),
		}

		assert.Equal(t, "Player1", g.nextActiveTurnPlayer())
	})
}

// ──────────────────────────────────────────────────────────────────────────────
// isPlayerWinner
// ──────────────────────────────────────────────────────────────────────────────

func TestGame_isPlayerWinner(t *testing.T) {
	t.Run("Returns false when game not over", func(t *testing.T) {
		g := &game{}
		assert.False(t, g.isPlayerWinner(0))
	})

	t.Run("Returns true for direct winner", func(t *testing.T) {
		g := &game{winState: winState{GameOver: true, WinnerIdx: 0}}
		assert.True(t, g.isPlayerWinner(0))
		assert.False(t, g.isPlayerWinner(1))
	})

	t.Run("Returns true for ally winner in 2v2", func(t *testing.T) {
		g := &game{
			mode:     types.GameMode2v2,
			teams:    map[int][]int{1: {0, 2}, 2: {1, 3}},
			winState: winState{GameOver: true, WinnerIdx: 0},
		}
		assert.True(t, g.isPlayerWinner(2))  // same team as winner
		assert.False(t, g.isPlayerWinner(1)) // opposing team
	})
}

// ──────────────────────────────────────────────────────────────────────────────
// ExecuteAction — kills and damage tracking
// ──────────────────────────────────────────────────────────────────────────────

func TestGame_ExecuteAction_TracksKillsAndDamage(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockPlayer1 := mocks.NewMockPlayer(ctrl)
	mockPlayer2 := mocks.NewMockPlayer(ctrl)
	mockAction := mocks.NewMockGameAction(ctrl)
	mockHand := mocks.NewMockHand(ctrl)

	expectedStatus := gamestatus.GameStatus{CurrentPlayer: "Player1"}
	actionResult := &gameactions.Result{
		Action: types.LastActionAttack,
		Attack: &gameactions.AttackDetails{
			KillsGranted: 2,
			DamageDealt:  7,
		},
	}

	mockPlayer1.EXPECT().Name().Return("Player1").AnyTimes()
	mockAction.EXPECT().PlayerName().Return("Player1").AnyTimes()
	mockAction.EXPECT().Validate(gomock.Any()).Return(nil)
	mockAction.EXPECT().Execute(gomock.Any()).Return(actionResult, func() gamestatus.GameStatus {
		return expectedStatus
	}, nil)
	mockAction.EXPECT().NextPhase().Return(types.PhaseTypeAttack)

	mockPlayer1.EXPECT().HasWarriorsInHand().Return(false)
	mockPlayer1.EXPECT().CanTradeCards().Return(false)
	mockPlayer1.EXPECT().CanForgeWeapons().Return(false)
	mockPlayer1.EXPECT().Hand().Return(mockHand).AnyTimes()
	mockHand.EXPECT().ShowCards().Return([]cards.Card{}).AnyTimes()
	mockPlayer1.EXPECT().CanAttack().Return(true)

	g := &game{
		board:        &testBoardImpl{players: []board.Player{mockPlayer1, mockPlayer2}},
		currentTurn:  0,
		playerKills:  make(map[string]int),
		playerDamage: make(map[string]int),
	}

	_, err := g.ExecuteAction(mockAction)
	assert.NoError(t, err)
	assert.Equal(t, 2, g.playerKills["Player1"])
	assert.Equal(t, 7, g.playerDamage["Player1"])
}

// ──────────────────────────────────────────────────────────────────────────────
// DisconnectPlayer
// ──────────────────────────────────────────────────────────────────────────────

func TestGame_DisconnectPlayer(t *testing.T) {
	t.Run("Error when player not found", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockPlayer1 := mocks.NewMockPlayer(ctrl)
		mockPlayer1.EXPECT().Name().Return("Player1").AnyTimes()

		g := &game{board: &testBoardImpl{players: []board.Player{mockPlayer1}}}

		err := g.DisconnectPlayer("Unknown")
		assert.Error(t, err)
	})

	t.Run("No-op when game is already over", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockPlayer1 := mocks.NewMockPlayer(ctrl)
		mockPlayer1.EXPECT().Name().Return("Player1").AnyTimes()

		g := &game{
			board:               &testBoardImpl{players: []board.Player{mockPlayer1}},
			winState:            winState{GameOver: true},
			eliminatedPlayers:   make(map[int]bool),
			disconnectedPlayers: make(map[int]bool),
			history:             []types.HistoryLine{},
		}

		err := g.DisconnectPlayer("Player1")
		assert.NoError(t, err)
		assert.False(t, g.disconnectedPlayers[0])
	})

	t.Run("No-op when player already eliminated", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockPlayer1 := mocks.NewMockPlayer(ctrl)
		mockPlayer1.EXPECT().Name().Return("Player1").AnyTimes()

		g := &game{
			board:               &testBoardImpl{players: []board.Player{mockPlayer1}},
			eliminatedPlayers:   map[int]bool{0: true},
			disconnectedPlayers: make(map[int]bool),
			history:             []types.HistoryLine{},
		}

		err := g.DisconnectPlayer("Player1")
		assert.NoError(t, err)
		assert.False(t, g.disconnectedPlayers[0])
	})

	t.Run("No-op when player already disconnected", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockPlayer1 := mocks.NewMockPlayer(ctrl)
		mockPlayer2 := mocks.NewMockPlayer(ctrl)
		mockPlayer1.EXPECT().Name().Return("Player1").AnyTimes()
		mockPlayer2.EXPECT().Name().Return("Player2").AnyTimes()

		g := &game{
			board:               &testBoardImpl{players: []board.Player{mockPlayer1, mockPlayer2}},
			eliminatedPlayers:   make(map[int]bool),
			disconnectedPlayers: map[int]bool{0: true},
			history:             []types.HistoryLine{},
		}

		err := g.DisconnectPlayer("Player1")
		assert.NoError(t, err)
		assert.Empty(t, g.history) // second disconnect is a no-op
	})

	t.Run("Marks player as disconnected and adds history when not their turn", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockPlayer1 := mocks.NewMockPlayer(ctrl)
		mockPlayer2 := mocks.NewMockPlayer(ctrl)
		mockPlayer1.EXPECT().Name().Return("Player1").AnyTimes()
		mockPlayer2.EXPECT().Name().Return("Player2").AnyTimes()

		g := &game{
			board:               &testBoardImpl{players: []board.Player{mockPlayer1, mockPlayer2}},
			currentTurn:         1, // Player2's turn
			eliminatedPlayers:   make(map[int]bool),
			disconnectedPlayers: make(map[int]bool),
			history:             []types.HistoryLine{},
		}

		err := g.DisconnectPlayer("Player1")
		assert.NoError(t, err)
		assert.True(t, g.disconnectedPlayers[0])
		assert.Contains(t, g.history, types.HistoryLine{Msg: "Player1 disconnected", Category: types.CategoryElimination})
		assert.Equal(t, 1, g.currentTurn) // turn unchanged
	})

	t.Run("Switches turn when it was the disconnected player turn", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockPlayer1 := mocks.NewMockPlayer(ctrl)
		mockPlayer2 := mocks.NewMockPlayer(ctrl)
		mockPlayer1.EXPECT().Name().Return("Player1").AnyTimes()
		mockPlayer2.EXPECT().Name().Return("Player2").AnyTimes()

		g := &game{
			board:               &testBoardImpl{players: []board.Player{mockPlayer1, mockPlayer2}},
			currentTurn:         0, // Player1's turn
			eliminatedPlayers:   make(map[int]bool),
			disconnectedPlayers: make(map[int]bool),
			history:             []types.HistoryLine{},
		}

		err := g.DisconnectPlayer("Player1")
		assert.NoError(t, err)
		assert.True(t, g.disconnectedPlayers[0])
		assert.Equal(t, 1, g.currentTurn) // advanced to Player2
	})
}

// ──────────────────────────────────────────────────────────────────────────────
// ReconnectPlayer
// ──────────────────────────────────────────────────────────────────────────────

func TestGame_ReconnectPlayer(t *testing.T) {
	t.Run("No-op when player not found", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockPlayer1 := mocks.NewMockPlayer(ctrl)
		mockPlayer1.EXPECT().Name().Return("Player1").AnyTimes()

		g := &game{
			board:               &testBoardImpl{players: []board.Player{mockPlayer1}},
			disconnectedPlayers: make(map[int]bool),
			history:             []types.HistoryLine{},
		}

		g.ReconnectPlayer("Unknown")
		assert.Empty(t, g.history)
	})

	t.Run("No-op when player is not disconnected", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockPlayer1 := mocks.NewMockPlayer(ctrl)
		mockPlayer1.EXPECT().Name().Return("Player1").AnyTimes()

		g := &game{
			board:               &testBoardImpl{players: []board.Player{mockPlayer1}},
			disconnectedPlayers: make(map[int]bool),
			history:             []types.HistoryLine{},
		}

		g.ReconnectPlayer("Player1")
		assert.Empty(t, g.history)
	})

	t.Run("Clears disconnected flag and adds history", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockPlayer1 := mocks.NewMockPlayer(ctrl)
		mockPlayer1.EXPECT().Name().Return("Player1").AnyTimes()

		g := &game{
			board:               &testBoardImpl{players: []board.Player{mockPlayer1}},
			disconnectedPlayers: map[int]bool{0: true},
			history:             []types.HistoryLine{},
		}

		g.ReconnectPlayer("Player1")
		assert.False(t, g.disconnectedPlayers[0])
		assert.Contains(t, g.history, types.HistoryLine{Msg: "Player1 reconnected", Category: types.CategoryElimination})
	})
}

// ──────────────────────────────────────────────────────────────────────────────
// FinalizeDisconnection
// ──────────────────────────────────────────────────────────────────────────────

func TestGame_FinalizeDisconnection(t *testing.T) {
	t.Run("No-op when player not found", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockPlayer1 := mocks.NewMockPlayer(ctrl)
		mockPlayer1.EXPECT().Name().Return("Player1").AnyTimes()

		g := &game{
			board:               &testBoardImpl{players: []board.Player{mockPlayer1}},
			disconnectedPlayers: make(map[int]bool),
		}

		g.FinalizeDisconnection("Unknown")
		assert.False(t, g.winState.GameOver)
	})

	t.Run("No-op when player is not disconnected", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockPlayer1 := mocks.NewMockPlayer(ctrl)
		mockPlayer1.EXPECT().Name().Return("Player1").AnyTimes()

		g := &game{
			board:               &testBoardImpl{players: []board.Player{mockPlayer1}},
			disconnectedPlayers: make(map[int]bool),
		}

		g.FinalizeDisconnection("Player1")
		assert.False(t, g.winState.GameOver)
	})

	t.Run("No-op when game is already over", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockPlayer1 := mocks.NewMockPlayer(ctrl)
		mockPlayer1.EXPECT().Name().Return("Player1").AnyTimes()

		g := &game{
			board:               &testBoardImpl{players: []board.Player{mockPlayer1}},
			disconnectedPlayers: map[int]bool{0: true},
			winState:            winState{GameOver: true, Winner: "Player1"},
		}

		g.FinalizeDisconnection("Player1")
		assert.Equal(t, "Player1", g.winState.Winner) // unchanged
	})

	t.Run("1v1: last active player wins", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockPlayer1 := mocks.NewMockPlayer(ctrl)
		mockPlayer2 := mocks.NewMockPlayer(ctrl)
		mockPlayer1.EXPECT().Name().Return("Player1").AnyTimes()
		mockPlayer2.EXPECT().Name().Return("Player2").AnyTimes()

		g := &game{
			board:               &testBoardImpl{players: []board.Player{mockPlayer1, mockPlayer2}},
			mode:                types.GameMode1v1,
			eliminatedPlayers:   make(map[int]bool),
			disconnectedPlayers: map[int]bool{1: true},
		}

		g.FinalizeDisconnection("Player2")

		assert.True(t, g.winState.GameOver)
		assert.Equal(t, "Player1", g.winState.Winner)
	})

	t.Run("FFA3: last active player wins", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockPlayer1 := mocks.NewMockPlayer(ctrl)
		mockPlayer2 := mocks.NewMockPlayer(ctrl)
		mockPlayer3 := mocks.NewMockPlayer(ctrl)
		mockPlayer1.EXPECT().Name().Return("Player1").AnyTimes()
		mockPlayer2.EXPECT().Name().Return("Player2").AnyTimes()
		mockPlayer3.EXPECT().Name().Return("Player3").AnyTimes()

		g := &game{
			board:               &testBoardImpl{players: []board.Player{mockPlayer1, mockPlayer2, mockPlayer3}},
			mode:                types.GameModeFFA3,
			eliminatedPlayers:   map[int]bool{1: true},
			disconnectedPlayers: map[int]bool{2: true},
		}

		g.FinalizeDisconnection("Player3")

		assert.True(t, g.winState.GameOver)
		assert.Equal(t, "Player1", g.winState.Winner)
	})

	t.Run("All players out: nobody wins", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockPlayer1 := mocks.NewMockPlayer(ctrl)
		mockPlayer2 := mocks.NewMockPlayer(ctrl)
		mockPlayer1.EXPECT().Name().Return("Player1").AnyTimes()
		mockPlayer2.EXPECT().Name().Return("Player2").AnyTimes()

		g := &game{
			board:               &testBoardImpl{players: []board.Player{mockPlayer1, mockPlayer2}},
			mode:                types.GameMode1v1,
			eliminatedPlayers:   make(map[int]bool),
			disconnectedPlayers: map[int]bool{0: true, 1: true},
		}

		g.FinalizeDisconnection("Player1")

		assert.True(t, g.winState.GameOver)
		assert.Equal(t, "nobody", g.winState.Winner)
		assert.Equal(t, -1, g.winState.WinnerIdx)
	})

	t.Run("2v2: entire enemy team disconnected, other team wins", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockPlayer1 := mocks.NewMockPlayer(ctrl)
		mockPlayer2 := mocks.NewMockPlayer(ctrl)
		mockPlayer3 := mocks.NewMockPlayer(ctrl)
		mockPlayer4 := mocks.NewMockPlayer(ctrl)
		mockPlayer1.EXPECT().Name().Return("Player1").AnyTimes()
		mockPlayer2.EXPECT().Name().Return("Player2").AnyTimes()
		mockPlayer3.EXPECT().Name().Return("Player3").AnyTimes()
		mockPlayer4.EXPECT().Name().Return("Player4").AnyTimes()

		// Team 1: indices 0,2  |  Team 2: indices 1,3 (both disconnected)
		g := &game{
			board:               &testBoardImpl{players: []board.Player{mockPlayer1, mockPlayer2, mockPlayer3, mockPlayer4}},
			mode:                types.GameMode2v2,
			teams:               map[int][]int{1: {0, 2}, 2: {1, 3}},
			eliminatedPlayers:   make(map[int]bool),
			disconnectedPlayers: map[int]bool{1: true, 3: true},
		}

		g.FinalizeDisconnection("Player2")

		assert.True(t, g.winState.GameOver)
		assert.Contains(t, g.winState.Winner, "team")
	})

	t.Run("2v2: game continues when only one member of enemy team is out", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockPlayer1 := mocks.NewMockPlayer(ctrl)
		mockPlayer2 := mocks.NewMockPlayer(ctrl)
		mockPlayer3 := mocks.NewMockPlayer(ctrl)
		mockPlayer4 := mocks.NewMockPlayer(ctrl)
		mockPlayer1.EXPECT().Name().Return("Player1").AnyTimes()
		mockPlayer2.EXPECT().Name().Return("Player2").AnyTimes()
		mockPlayer3.EXPECT().Name().Return("Player3").AnyTimes()
		mockPlayer4.EXPECT().Name().Return("Player4").AnyTimes()

		g := &game{
			board:               &testBoardImpl{players: []board.Player{mockPlayer1, mockPlayer2, mockPlayer3, mockPlayer4}},
			mode:                types.GameMode2v2,
			teams:               map[int][]int{1: {0, 2}, 2: {1, 3}},
			eliminatedPlayers:   make(map[int]bool),
			disconnectedPlayers: map[int]bool{1: true}, // Player2 out, Player4 still active
		}

		g.FinalizeDisconnection("Player2")

		assert.False(t, g.winState.GameOver)
	})
}
