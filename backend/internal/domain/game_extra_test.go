package domain

import (
	"testing"
	"time"

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

	// Regression: when all players disconnect in a 1v1, SwitchTurn must not
	// loop forever. Previously this caused the hub goroutine to deadlock.
	t.Run("Does not infinite-loop when all players are disconnected (1v1)", func(t *testing.T) {
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
			disconnectedPlayers: map[int]bool{0: true}, // Player1 already disconnected
			history:             []types.HistoryLine{},
		}

		// Player2 disconnects while it is Player1's turn (which was already skipped to P2).
		// Pretend currentTurn is now 1 (Player2's turn after Player1 disconnected earlier).
		g.currentTurn = 1
		g.disconnectedPlayers[0] = true

		// This must return without hanging.
		done := make(chan struct{})
		go func() {
			_ = g.DisconnectPlayer("Player2")
			close(done)
		}()

		select {
		case <-done:
			// success
		case <-time.After(2 * time.Second):
			t.Fatal("DisconnectPlayer deadlocked: SwitchTurn did not exit when all players disconnected")
		}

		assert.True(t, g.disconnectedPlayers[1])
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

// ──────────────────────────────────────────────────────────────────────────────
// Simple getters: CurrentAction, Board, TurnState, EventHandler
// ──────────────────────────────────────────────────────────────────────────────

func TestGame_CurrentAction(t *testing.T) {
	g := &game{currentAction: types.PhaseTypeAttack}
	assert.Equal(t, types.PhaseTypeAttack, g.CurrentAction())
}

func TestGame_Board(t *testing.T) {
	b := &testBoardImpl{}
	g := &game{board: b}
	assert.Equal(t, b, g.Board())
}

func TestGame_TurnState(t *testing.T) {
	ts := types.TurnState{HasMovedWarrior: true}
	g := &game{turnState: ts}
	assert.Equal(t, ts, g.TurnState())
}

func TestGame_EventHandler_ReturnsNonNil(t *testing.T) {
	g := &game{}
	handler := g.EventHandler()
	assert.NotNil(t, handler)
}

// ──────────────────────────────────────────────────────────────────────────────
// DrawCards — success path
// ──────────────────────────────────────────────────────────────────────────────

func TestGame_DrawCards_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockPlayer := mocks.NewMockPlayer(ctrl)
	mockDeck := mocks.NewMockDeck(ctrl)
	mockDiscard := mocks.NewMockDiscardPile(ctrl)
	mockCard := mocks.NewMockCard(ctrl)

	mockPlayer.EXPECT().CanTakeCards(1).Return(true)
	mockDeck.EXPECT().DrawCards(1, mockDiscard).Return([]cards.Card{mockCard}, nil)

	g := &game{board: &testBoardImpl{deck: mockDeck, discardPile: mockDiscard}}

	result, err := g.DrawCards(mockPlayer, 1)
	assert.NoError(t, err)
	assert.Len(t, result, 1)
	assert.Contains(t, result, mockCard)
}

// ──────────────────────────────────────────────────────────────────────────────
// AutoMoveWarriorsToField
// ──────────────────────────────────────────────────────────────────────────────

func TestGame_AutoMoveWarriorsToField_PlayerNotFound(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockPlayer1 := mocks.NewMockPlayer(ctrl)
	mockPlayer1.EXPECT().Name().Return("Player1").AnyTimes()

	g := &game{board: &testBoardImpl{players: []board.Player{mockPlayer1}}}

	err := g.AutoMoveWarriorsToField("Unknown")
	assert.Error(t, err)
}

func TestGame_AutoMoveWarriorsToField_MovesWarrior(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockPlayer1 := mocks.NewMockPlayer(ctrl)
	mockPlayer1.EXPECT().Name().Return("Player1").AnyTimes()

	mockWarrior := mocks.NewMockWarrior(ctrl)
	mockWarrior.EXPECT().Type().Return(types.KnightWarriorType).AnyTimes()
	mockWarrior.EXPECT().GetID().Return("w1")

	mockHand := mocks.NewMockHand(ctrl)
	mockHand.EXPECT().ShowCards().Return([]cards.Card{mockWarrior}).AnyTimes()
	mockPlayer1.EXPECT().Hand().Return(mockHand).AnyTimes()
	mockPlayer1.EXPECT().MoveCardToField("w1").Return(nil)

	g := &game{board: &testBoardImpl{players: []board.Player{mockPlayer1}}}

	err := g.AutoMoveWarriorsToField("Player1")
	assert.NoError(t, err)
}

func TestGame_AutoMoveWarriorsToField_SkipsDragon(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockPlayer1 := mocks.NewMockPlayer(ctrl)
	mockPlayer1.EXPECT().Name().Return("Player1").AnyTimes()

	mockDragon := mocks.NewMockWarrior(ctrl)
	mockDragon.EXPECT().Type().Return(types.DragonWarriorType).AnyTimes()

	mockHand := mocks.NewMockHand(ctrl)
	mockHand.EXPECT().ShowCards().Return([]cards.Card{mockDragon}).AnyTimes()
	mockPlayer1.EXPECT().Hand().Return(mockHand).AnyTimes()

	// MoveCardToField must NOT be called for a dragon
	g := &game{board: &testBoardImpl{players: []board.Player{mockPlayer1}}}

	err := g.AutoMoveWarriorsToField("Player1")
	assert.NoError(t, err)
}

func TestGame_AutoMoveWarriorsToField_StopsAfterThree(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockPlayer1 := mocks.NewMockPlayer(ctrl)
	mockPlayer1.EXPECT().Name().Return("Player1").AnyTimes()

	makeWarrior := func(id string) *mocks.MockWarrior {
		w := mocks.NewMockWarrior(ctrl)
		w.EXPECT().Type().Return(types.KnightWarriorType).AnyTimes()
		w.EXPECT().GetID().Return(id).AnyTimes()
		return w
	}
	w1 := makeWarrior("w1")
	w2 := makeWarrior("w2")
	w3 := makeWarrior("w3")
	w4 := makeWarrior("w4")

	mockHand := mocks.NewMockHand(ctrl)
	mockHand.EXPECT().ShowCards().Return([]cards.Card{w1, w2, w3, w4}).AnyTimes()
	mockPlayer1.EXPECT().Hand().Return(mockHand).AnyTimes()

	// Only first 3 should be moved
	mockPlayer1.EXPECT().MoveCardToField("w1").Return(nil)
	mockPlayer1.EXPECT().MoveCardToField("w2").Return(nil)
	mockPlayer1.EXPECT().MoveCardToField("w3").Return(nil)

	g := &game{board: &testBoardImpl{players: []board.Player{mockPlayer1}}}

	err := g.AutoMoveWarriorsToField("Player1")
	assert.NoError(t, err)
}

// ──────────────────────────────────────────────────────────────────────────────
// nextAction — phase branches
// ──────────────────────────────────────────────────────────────────────────────

// setupNextActionPlayer creates a mock player with the 3 always-called expectations.
func setupNextActionPlayer(ctrl *gomock.Controller, name string) *mocks.MockPlayer {
	p := mocks.NewMockPlayer(ctrl)
	p.EXPECT().Name().Return(name).AnyTimes()
	p.EXPECT().HasWarriorsInHand().Return(false)
	p.EXPECT().CanTradeCards().Return(false)
	p.EXPECT().CanForgeWeapons().Return(false)
	return p
}

func TestGame_nextAction_SetsExpectedPhase(t *testing.T) {
	phases := []types.PhaseType{
		types.PhaseTypeAttack,
		types.PhaseTypeBuy,
		types.PhaseTypeConstruct,
		types.PhaseTypeEndTurn,
	}
	for _, phase := range phases {
		t.Run(string(phase), func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockPlayer1 := setupNextActionPlayer(ctrl, "Player1")
			mockPlayer2 := mocks.NewMockPlayer(ctrl)

			expected := gamestatus.GameStatus{}
			g := &game{
				board:             &testBoardImpl{players: []board.Player{mockPlayer1, mockPlayer2}},
				currentTurn:       0,
				eliminatedPlayers: make(map[int]bool),
			}

			g.nextAction(phase, func() gamestatus.GameStatus { return expected })
			assert.Equal(t, phase, g.currentAction)
		})
	}
}

func TestGame_nextAction_SpySteal_HasSpy(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockPlayer1 := setupNextActionPlayer(ctrl, "Player1")
	mockPlayer2 := mocks.NewMockPlayer(ctrl)
	mockHand := mocks.NewMockHand(ctrl)
	mockSpy := mocks.NewMockSpy(ctrl)
	mockPlayer1.EXPECT().Hand().Return(mockHand).AnyTimes()
	mockHand.EXPECT().ShowCards().Return([]cards.Card{mockSpy}).AnyTimes()

	expected := gamestatus.GameStatus{}
	g := &game{
		board:             &testBoardImpl{players: []board.Player{mockPlayer1, mockPlayer2}},
		currentTurn:       0,
		eliminatedPlayers: make(map[int]bool),
	}

	g.nextAction(types.PhaseTypeSpySteal, func() gamestatus.GameStatus { return expected })
	assert.Equal(t, types.PhaseTypeSpySteal, g.currentAction)
}

func TestGame_nextAction_SpySteal_AutoSkipsToByWhenNoCards(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockPlayer1 := setupNextActionPlayer(ctrl, "Player1")
	mockPlayer2 := mocks.NewMockPlayer(ctrl)
	mockHand := mocks.NewMockHand(ctrl)
	mockPlayer1.EXPECT().Hand().Return(mockHand).AnyTimes()
	mockHand.EXPECT().ShowCards().Return([]cards.Card{}).AnyTimes()

	expected := gamestatus.GameStatus{}
	g := &game{
		board:             &testBoardImpl{players: []board.Player{mockPlayer1, mockPlayer2}},
		currentTurn:       0,
		eliminatedPlayers: make(map[int]bool),
	}

	g.nextAction(types.PhaseTypeSpySteal, func() gamestatus.GameStatus { return expected })
	assert.Equal(t, types.PhaseTypeBuy, g.currentAction)
}

// ──────────────────────────────────────────────────────────────────────────────
// Status / StatusWithModal / getStatus / extractField / extractCastle
// (integration tests using a real NewGame)
// ──────────────────────────────────────────────────────────────────────────────

func TestGame_Status_Integration(t *testing.T) {
	dealer := cards.NewDealer(minDeckConfig())
	g, err := NewGame([]string{"Alice", "Bob"}, types.GameMode1v1, dealer, 25)
	assert.NoError(t, err)

	alice := g.board.Players()[0]

	status := g.Status(alice)
	assert.Equal(t, "Alice", status.CurrentPlayer) // viewer name
	assert.Equal(t, "Alice", status.TurnPlayer)    // turn player (fresh game)
}

func TestGame_StatusWithModal_Integration(t *testing.T) {
	dealer := cards.NewDealer(minDeckConfig())
	g, err := NewGame([]string{"Alice", "Bob"}, types.GameMode1v1, dealer, 25)
	assert.NoError(t, err)

	alice := g.board.Players()[0]

	status := g.StatusWithModal(alice, nil)
	assert.Equal(t, "Alice", status.CurrentPlayer)
}

func TestGame_Status_Opponent_HasCorrectOpponentCount(t *testing.T) {
	dealer := cards.NewDealer(minDeckConfig())
	g, err := NewGame([]string{"Alice", "Bob"}, types.GameMode1v1, dealer, 25)
	assert.NoError(t, err)

	alice := g.board.Players()[0]

	status := g.Status(alice)
	assert.Len(t, status.Opponents, 1)
	assert.Equal(t, "Bob", status.Opponents[0].PlayerName)
}

func TestGame_Status_GameOver_PopulatesPlayerStats(t *testing.T) {
	dealer := cards.NewDealer(minDeckConfig())
	g, err := NewGame([]string{"Alice", "Bob"}, types.GameMode1v1, dealer, 25)
	assert.NoError(t, err)

	// Simulate game over
	g.winState = winState{GameOver: true, Winner: "Alice", WinnerIdx: 0}
	g.playerKills = map[string]int{"Alice": 3, "Bob": 1}
	g.playerDamage = map[string]int{"Alice": 10, "Bob": 5}

	alice := g.board.Players()[0]
	status := g.Status(alice)

	assert.NotEmpty(t, status.GameOverMsg)
	assert.Len(t, status.PlayerStats, 2)
}

func TestGame_Status_2v2_CoversAllyFields(t *testing.T) {
	dealer := cards.NewDealer(cards.DeckConfig{Warriors: 4, ConstructionCards: 2})
	g, err := NewGame([]string{"A", "B", "C", "D"}, types.GameMode2v2, dealer, 30)
	assert.NoError(t, err)

	a := g.board.Players()[0]
	status := g.Status(a)
	// In 2v2 mode: A's ally is C (team 1: {0,2}); enemies are B and D
	assert.Equal(t, "A", status.CurrentPlayer)
	assert.Len(t, status.Opponents, 3) // B (enemy), C (ally), D (enemy)
}

// ──────────────────────────────────────────────────────────────────────────────
// nextActiveTurnPlayer — return "" when all players are out
// ──────────────────────────────────────────────────────────────────────────────

func TestGame_nextActiveTurnPlayer_ReturnsEmptyWhenAllOut(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockPlayer1 := mocks.NewMockPlayer(ctrl)
	mockPlayer2 := mocks.NewMockPlayer(ctrl)
	// Both eliminated/disconnected → loop wraps back to currentTurn → return ""
	g := &game{
		board:               &testBoardImpl{players: []board.Player{mockPlayer1, mockPlayer2}},
		currentTurn:         0,
		eliminatedPlayers:   map[int]bool{0: true},
		disconnectedPlayers: map[int]bool{1: true},
	}

	assert.Equal(t, "", g.nextActiveTurnPlayer())
}

// ──────────────────────────────────────────────────────────────────────────────
// getStatus — lastResult optional field branches
// ──────────────────────────────────────────────────────────────────────────────

func TestGame_getStatus_LastResult_AllBranches(t *testing.T) {
	dealer := cards.NewDealer(minDeckConfig())
	g, err := NewGame([]string{"Alice", "Bob"}, types.GameMode1v1, dealer, 25)
	assert.NoError(t, err)

	// Populate all lastResult optional pointer fields to exercise the nil-check branches
	g.lastResult = gameactions.Result{
		Attack:       &gameactions.AttackDetails{WeaponID: "w1", TargetID: "t1", TargetPlayer: "Bob"},
		Steal:        &gameactions.StealDetails{From: "Bob"},
		Sabotage:     &gameactions.SabotageDetails{From: "Bob"},
		Spy:          &types.SpyInfo{Target: types.SpyTargetDeck},
		Treason:      &gameactions.TreasonDetails{FromPlayer: "Bob"},
		Resurrection: &gameactions.ResurrectionDetails{TargetPlayer: "Bob", PlayerName: "Alice"},
		PlaceAmbush:  &gameactions.PlaceAmbushDetails{TargetPlayer: "Bob"},
	}

	alice := g.board.Players()[0]
	// Must not panic; all nil-check branches in getStatus are exercised
	status := g.Status(alice)
	assert.NotEmpty(t, status.CurrentPlayer)
}
