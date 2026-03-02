package domain

import (
	"errors"
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

// testBoardImpl is a minimal board.Board implementation for use in tests.
type testBoardImpl struct {
	players     []board.Player
	discardPile board.DiscardPile
	cemetery    board.Cemetery
	deck        board.Deck
}

func (b *testBoardImpl) Deck() board.Deck               { return b.deck }
func (b *testBoardImpl) DiscardPile() board.DiscardPile { return b.discardPile }
func (b *testBoardImpl) Cemetery() board.Cemetery       { return b.cemetery }
func (b *testBoardImpl) Players() []board.Player        { return b.players }

func TestGame_ExecuteAction(t *testing.T) {
	t.Run("Error when not current player's turn", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockPlayer1 := mocks.NewMockPlayer(ctrl)
		mockPlayer2 := mocks.NewMockPlayer(ctrl)
		mockAction := mocks.NewMockGameAction(ctrl)

		mockPlayer1.EXPECT().Name().Return("Player1").AnyTimes()
		mockAction.EXPECT().PlayerName().Return("Player2").AnyTimes()

		g := &game{
			board:       &testBoardImpl{players: []board.Player{mockPlayer1, mockPlayer2}},
			currentTurn: 0,
		}

		status, err := g.ExecuteAction(mockAction)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "Player2 not your turn")
		assert.Equal(t, gamestatus.GameStatus{}, status)
	})

	t.Run("Error when Validate fails", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockPlayer1 := mocks.NewMockPlayer(ctrl)
		mockPlayer2 := mocks.NewMockPlayer(ctrl)
		mockAction := mocks.NewMockGameAction(ctrl)

		mockPlayer1.EXPECT().Name().Return("Player1").AnyTimes()
		mockAction.EXPECT().PlayerName().Return("Player1").AnyTimes()
		mockAction.EXPECT().Validate(gomock.Any()).Return(errors.New("validation failed"))

		g := &game{
			board:       &testBoardImpl{players: []board.Player{mockPlayer1, mockPlayer2}},
			currentTurn: 0,
		}

		status, err := g.ExecuteAction(mockAction)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "validation failed")
		assert.Equal(t, gamestatus.GameStatus{}, status)
	})

	t.Run("Error when Execute fails", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockPlayer1 := mocks.NewMockPlayer(ctrl)
		mockPlayer2 := mocks.NewMockPlayer(ctrl)
		mockAction := mocks.NewMockGameAction(ctrl)

		mockPlayer1.EXPECT().Name().Return("Player1").AnyTimes()
		mockAction.EXPECT().PlayerName().Return("Player1").AnyTimes()
		mockAction.EXPECT().Validate(gomock.Any()).Return(nil)
		mockAction.EXPECT().Execute(gomock.Any()).Return(nil, nil, errors.New("execute failed"))

		g := &game{
			board:       &testBoardImpl{players: []board.Player{mockPlayer1, mockPlayer2}},
			currentTurn: 0,
		}

		status, err := g.ExecuteAction(mockAction)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "execute failed")
		assert.Equal(t, gamestatus.GameStatus{}, status)
	})

	t.Run("Success stores lastResult and advances phase", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockPlayer1 := mocks.NewMockPlayer(ctrl)
		mockPlayer2 := mocks.NewMockPlayer(ctrl)
		mockAction := mocks.NewMockGameAction(ctrl)
		mockHand := mocks.NewMockHand(ctrl)

		expectedStatus := gamestatus.GameStatus{CurrentPlayer: "Player1"}
		actionResult := &gameactions.Result{Action: types.LastActionDraw}

		mockPlayer1.EXPECT().Name().Return("Player1").AnyTimes()
		mockAction.EXPECT().PlayerName().Return("Player1").AnyTimes()
		mockAction.EXPECT().Validate(gomock.Any()).Return(nil)
		mockAction.EXPECT().Execute(gomock.Any()).Return(actionResult, func() gamestatus.GameStatus {
			return expectedStatus
		}, nil)
		mockAction.EXPECT().NextPhase().Return(types.PhaseTypeAttack)

		// nextAction expectations for PhaseTypeAttack
		mockPlayer1.EXPECT().HasWarriorsInHand().Return(false)
		mockPlayer1.EXPECT().CanTradeCards().Return(false)
		// HasCardTypeInHand[Catapult/BloodRain/Harpoon] each call Hand().ShowCards()
		mockPlayer1.EXPECT().Hand().Return(mockHand).AnyTimes()
		mockHand.EXPECT().ShowCards().Return([]cards.Card{}).AnyTimes()
		mockPlayer1.EXPECT().CanAttack().Return(true)

		g := &game{
			board:       &testBoardImpl{players: []board.Player{mockPlayer1, mockPlayer2}},
			currentTurn: 0,
		}

		status, err := g.ExecuteAction(mockAction)

		assert.NoError(t, err)
		assert.Equal(t, expectedStatus, status)
		assert.Equal(t, types.LastActionDraw, g.lastResult.Action)
		assert.Equal(t, types.PhaseTypeAttack, g.currentAction)
	})

	t.Run("Success skips phases when player has no capabilities", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockPlayer1 := mocks.NewMockPlayer(ctrl)
		mockPlayer2 := mocks.NewMockPlayer(ctrl)
		mockAction := mocks.NewMockGameAction(ctrl)
		mockHand := mocks.NewMockHand(ctrl)

		expectedStatus := gamestatus.GameStatus{CurrentPlayer: "Player1"}
		actionResult := &gameactions.Result{Action: types.LastActionDraw}

		mockPlayer1.EXPECT().Name().Return("Player1").AnyTimes()
		mockAction.EXPECT().PlayerName().Return("Player1").AnyTimes()
		mockAction.EXPECT().Validate(gomock.Any()).Return(nil)
		mockAction.EXPECT().Execute(gomock.Any()).Return(actionResult, func() gamestatus.GameStatus {
			return expectedStatus
		}, nil)
		mockAction.EXPECT().NextPhase().Return(types.PhaseTypeAttack)

		// nextAction: no attack, no spy/steal, no buy -> skips to construct
		mockPlayer1.EXPECT().HasWarriorsInHand().Return(false)
		mockPlayer1.EXPECT().CanTradeCards().Return(false)
		// HasCardTypeInHand[Catapult/BloodRain/Harpoon/Spy/Thief] each call Hand().ShowCards()
		mockPlayer1.EXPECT().Hand().Return(mockHand).AnyTimes()
		mockHand.EXPECT().ShowCards().Return([]cards.Card{}).AnyTimes()
		mockPlayer1.EXPECT().CanAttack().Return(false)
		mockPlayer1.EXPECT().CanBuy().Return(false)
		mockPlayer1.EXPECT().CanConstruct().Return(true)

		g := &game{
			board:       &testBoardImpl{players: []board.Player{mockPlayer1, mockPlayer2}},
			currentTurn: 0,
		}

		status, err := g.ExecuteAction(mockAction)

		assert.NoError(t, err)
		assert.Equal(t, expectedStatus, status)
		assert.Equal(t, types.PhaseTypeConstruct, g.currentAction)
	})
}

func TestGame_OnCastleCompletion(t *testing.T) {
	t.Run("1v1 sets individual winner", func(t *testing.T) {
		g := &game{
			mode:  types.GameMode1v1,
			board: &testBoardImpl{},
		}

		g.OnCastleCompletion("Player1")

		assert.True(t, g.winState.GameOver)
		assert.Equal(t, "Player1", g.winState.Winner)
	})

	t.Run("2v2 sets team winner", func(t *testing.T) {
		g := &game{
			mode:  types.GameMode2v2,
			board: &testBoardImpl{},
		}

		g.OnCastleCompletion("Player1")

		assert.True(t, g.winState.GameOver)
		assert.Equal(t, "Player1's team", g.winState.Winner)
	})

	t.Run("FFA3 sets individual winner", func(t *testing.T) {
		g := &game{
			mode:  types.GameModeFFA3,
			board: &testBoardImpl{},
		}

		g.OnCastleCompletion("Player1")

		assert.True(t, g.winState.GameOver)
		assert.Equal(t, "Player1", g.winState.Winner)
	})
}

func TestGame_OnFieldWithoutWarriors(t *testing.T) {
	t.Run("1v1 current player wins immediately", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockPlayer1 := mocks.NewMockPlayer(ctrl)
		mockPlayer2 := mocks.NewMockPlayer(ctrl)
		mockPlayer1.EXPECT().Name().Return("Player1").AnyTimes()
		mockPlayer2.EXPECT().Name().Return("Player2").AnyTimes()

		g := &game{
			board:             &testBoardImpl{players: []board.Player{mockPlayer1, mockPlayer2}},
			currentTurn:       0,
			mode:              types.GameMode1v1,
			eliminatedPlayers: make(map[int]bool),
			history:           []types.HistoryLine{},
		}

		g.OnFieldWithoutWarriors("Player2")

		assert.True(t, g.winState.GameOver)
		assert.Equal(t, "Player1", g.winState.Winner)
	})

	t.Run("FFA3 eliminates player, game continues with 2 remaining", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockPlayer1 := mocks.NewMockPlayer(ctrl)
		mockPlayer2 := mocks.NewMockPlayer(ctrl)
		mockPlayer3 := mocks.NewMockPlayer(ctrl)
		mockPlayer1.EXPECT().Name().Return("Player1").AnyTimes()
		mockPlayer2.EXPECT().Name().Return("Player2").AnyTimes()
		mockPlayer3.EXPECT().Name().Return("Player3").AnyTimes()

		mockHand2 := mocks.NewMockHand(ctrl)
		mockCastle2 := mocks.NewMockCastle(ctrl)
		mockPlayer2.EXPECT().Hand().Return(mockHand2)
		mockHand2.EXPECT().ShowCards().Return([]cards.Card{})
		mockPlayer2.EXPECT().Castle().Return(mockCastle2)
		mockCastle2.EXPECT().ResourceCards().Return([]cards.Resource{})

		g := &game{
			board:             &testBoardImpl{players: []board.Player{mockPlayer1, mockPlayer2, mockPlayer3}},
			currentTurn:       0,
			mode:              types.GameModeFFA3,
			eliminatedPlayers: make(map[int]bool),
			history:           []types.HistoryLine{},
		}

		g.OnFieldWithoutWarriors("Player2")

		assert.False(t, g.winState.GameOver)
		assert.True(t, g.eliminatedPlayers[1])
		assert.Contains(t, g.history, types.HistoryLine{Msg: "Player2 has been eliminated!", Category: types.CategoryElimination})
	})

	t.Run("FFA3 last player standing wins", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockPlayer1 := mocks.NewMockPlayer(ctrl)
		mockPlayer2 := mocks.NewMockPlayer(ctrl)
		mockPlayer3 := mocks.NewMockPlayer(ctrl)
		mockPlayer1.EXPECT().Name().Return("Player1").AnyTimes()
		mockPlayer2.EXPECT().Name().Return("Player2").AnyTimes()
		mockPlayer3.EXPECT().Name().Return("Player3").AnyTimes()

		mockHand3 := mocks.NewMockHand(ctrl)
		mockCastle3 := mocks.NewMockCastle(ctrl)
		mockPlayer3.EXPECT().Hand().Return(mockHand3)
		mockHand3.EXPECT().ShowCards().Return([]cards.Card{})
		mockPlayer3.EXPECT().Castle().Return(mockCastle3)
		mockCastle3.EXPECT().ResourceCards().Return([]cards.Resource{})

		g := &game{
			board:             &testBoardImpl{players: []board.Player{mockPlayer1, mockPlayer2, mockPlayer3}},
			currentTurn:       0,
			mode:              types.GameModeFFA3,
			eliminatedPlayers: map[int]bool{1: true}, // Player2 already eliminated
			history:           []types.HistoryLine{},
		}

		g.OnFieldWithoutWarriors("Player3")

		assert.True(t, g.winState.GameOver)
		assert.Equal(t, "Player1", g.winState.Winner)
		assert.True(t, g.eliminatedPlayers[2])
	})

	t.Run("FFA5 eliminates player, game continues", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockPlayers := make([]*mocks.MockPlayer, 5)
		players := make([]board.Player, 5)
		for i := 0; i < 5; i++ {
			mp := mocks.NewMockPlayer(ctrl)
			mp.EXPECT().Name().Return(
				"Player" + string(rune('1'+i))).AnyTimes()
			mockPlayers[i] = mp
			players[i] = mp
		}

		// Mock Hand/Castle for eliminated player (Player2 = index 1)
		mockHand := mocks.NewMockHand(ctrl)
		mockCastle := mocks.NewMockCastle(ctrl)
		mockPlayers[1].EXPECT().Hand().Return(mockHand)
		mockHand.EXPECT().ShowCards().Return([]cards.Card{})
		mockPlayers[1].EXPECT().Castle().Return(mockCastle)
		mockCastle.EXPECT().ResourceCards().Return([]cards.Resource{})

		g := &game{
			board:             &testBoardImpl{players: players},
			currentTurn:       0,
			mode:              types.GameModeFFA5,
			eliminatedPlayers: make(map[int]bool),
			history:           []types.HistoryLine{},
		}

		g.OnFieldWithoutWarriors("Player2")

		assert.False(t, g.winState.GameOver)
		assert.True(t, g.eliminatedPlayers[1])
	})

	t.Run("2v2 eliminates one enemy, game continues", func(t *testing.T) {
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

		mockHand2 := mocks.NewMockHand(ctrl)
		mockCastle2 := mocks.NewMockCastle(ctrl)
		mockPlayer2.EXPECT().Hand().Return(mockHand2)
		mockHand2.EXPECT().ShowCards().Return([]cards.Card{})
		mockPlayer2.EXPECT().Castle().Return(mockCastle2)
		mockCastle2.EXPECT().ResourceCards().Return([]cards.Resource{})

		g := &game{
			board:             &testBoardImpl{players: []board.Player{mockPlayer1, mockPlayer2, mockPlayer3, mockPlayer4}},
			currentTurn:       0, // Player1's turn (Team 1)
			mode:              types.GameMode2v2,
			teams:             map[int][]int{1: {0, 2}, 2: {1, 3}},
			eliminatedPlayers: make(map[int]bool),
			history:           []types.HistoryLine{},
		}

		// Player2 (Team 2) loses warriors, but Player4 (Team 2) is still alive
		g.OnFieldWithoutWarriors("Player2")

		assert.False(t, g.winState.GameOver)
		assert.True(t, g.eliminatedPlayers[1])
		assert.Contains(t, g.history, types.HistoryLine{Msg: "Player2 has been eliminated!", Category: types.CategoryElimination})
	})

	t.Run("2v2 both enemies eliminated, team wins", func(t *testing.T) {
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

		mockHand4 := mocks.NewMockHand(ctrl)
		mockCastle4 := mocks.NewMockCastle(ctrl)
		mockPlayer4.EXPECT().Hand().Return(mockHand4)
		mockHand4.EXPECT().ShowCards().Return([]cards.Card{})
		mockPlayer4.EXPECT().Castle().Return(mockCastle4)
		mockCastle4.EXPECT().ResourceCards().Return([]cards.Resource{})

		g := &game{
			board:             &testBoardImpl{players: []board.Player{mockPlayer1, mockPlayer2, mockPlayer3, mockPlayer4}},
			currentTurn:       0, // Player1's turn (Team 1)
			mode:              types.GameMode2v2,
			teams:             map[int][]int{1: {0, 2}, 2: {1, 3}},
			eliminatedPlayers: map[int]bool{1: true}, // Player2 already eliminated
			history:           []types.HistoryLine{},
		}

		// Player4 (last of Team 2) loses warriors
		g.OnFieldWithoutWarriors("Player4")

		assert.True(t, g.winState.GameOver)
		assert.Equal(t, "Player1's team", g.winState.Winner)
		assert.True(t, g.eliminatedPlayers[3])
	})
}

func TestGame_IsGameOver(t *testing.T) {
	t.Run("Returns false initially", func(t *testing.T) {
		g := &game{}

		gameOver, winner := g.IsGameOver()

		assert.False(t, gameOver)
		assert.Empty(t, winner)
	})

	t.Run("Returns true after game ends", func(t *testing.T) {
		g := &game{
			winState: winState{GameOver: true, Winner: "Player1"},
		}

		gameOver, winner := g.IsGameOver()

		assert.True(t, gameOver)
		assert.Equal(t, "Player1", winner)
	})
}

func TestGame_GetHistory(t *testing.T) {
	t.Run("Returns all history on first call", func(t *testing.T) {
		g := &game{
			history: []types.HistoryLine{
				{Msg: "msg1", Category: types.CategoryInfo},
				{Msg: "msg2", Category: types.CategoryInfo},
				{Msg: "msg3", Category: types.CategoryInfo},
			},
		}

		result := g.getHistory()

		assert.Len(t, result, 3)
		assert.Equal(t, "msg1", result[0].Msg)
		assert.Equal(t, "msg3", result[2].Msg)
	})

	t.Run("Returns only new messages on subsequent calls", func(t *testing.T) {
		g := &game{
			history: []types.HistoryLine{
				{Msg: "msg1", Category: types.CategoryInfo},
				{Msg: "msg2", Category: types.CategoryInfo},
			},
		}

		_ = g.getHistory() // First call reads all

		g.history = append(g.history,
			types.HistoryLine{Msg: "msg3", Category: types.CategoryInfo},
			types.HistoryLine{Msg: "msg4", Category: types.CategoryInfo},
		)
		result := g.getHistory()

		assert.Len(t, result, 2)
		assert.Equal(t, "msg3", result[0].Msg)
		assert.Equal(t, "msg4", result[1].Msg)
	})

	t.Run("Returns empty slice when no new messages", func(t *testing.T) {
		g := &game{
			history: []types.HistoryLine{
				{Msg: "msg1", Category: types.CategoryInfo},
			},
		}

		_ = g.getHistory()
		result := g.getHistory()

		assert.Empty(t, result)
	})
}

func TestGame_OnWarriorMovedToCemetery(t *testing.T) {
	t.Run("Adds warrior to cemetery and records history", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockCemetery := mocks.NewMockCemetery(ctrl)
		mockWarrior := mocks.NewMockWarrior(ctrl)
		mockCemetery.EXPECT().AddCorp(mockWarrior)

		g := &game{
			board:   &testBoardImpl{cemetery: mockCemetery},
			history: []types.HistoryLine{},
		}

		g.OnWarriorMovedToCemetery(mockWarrior)

		assert.Contains(t, g.history, types.HistoryLine{Msg: "warrior buried in cemetery", Category: types.CategoryInfo})
	})
}

func TestGame_AutoMoveWarriorsToField(t *testing.T) {
	t.Run("Returns nil for player with empty hand", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockHand := mocks.NewMockHand(ctrl)
		mockPlayer1 := mocks.NewMockPlayer(ctrl)
		mockPlayer1.EXPECT().Name().Return("Player1").AnyTimes()
		mockPlayer1.EXPECT().Hand().Return(mockHand).Times(2)
		mockHand.EXPECT().ShowCards().Return([]cards.Card{}).Times(2)

		g := &game{board: &testBoardImpl{players: []board.Player{mockPlayer1}}}

		err := g.AutoMoveWarriorsToField("Player1")

		assert.NoError(t, err)
	})

	t.Run("Error when player not found", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockPlayer1 := mocks.NewMockPlayer(ctrl)
		mockPlayer1.EXPECT().Name().Return("Player1").AnyTimes()

		g := &game{board: &testBoardImpl{players: []board.Player{mockPlayer1}}}

		err := g.AutoMoveWarriorsToField("Unknown")

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "player Unknown not found")
	})
}

func TestGame_SameTeam(t *testing.T) {
	t.Run("Returns false for non-2v2 mode", func(t *testing.T) {
		g := &game{mode: types.GameMode1v1}
		assert.False(t, g.SameTeam(0, 1))
	})

	t.Run("Returns true for same team in 2v2", func(t *testing.T) {
		g := &game{
			mode:  types.GameMode2v2,
			teams: map[int][]int{1: {0, 2}, 2: {1, 3}},
		}

		assert.True(t, g.SameTeam(0, 2)) // Team 1
		assert.True(t, g.SameTeam(2, 0)) // Symmetric
		assert.True(t, g.SameTeam(1, 3)) // Team 2
		assert.True(t, g.SameTeam(3, 1)) // Symmetric
	})

	t.Run("Returns false for different teams in 2v2", func(t *testing.T) {
		g := &game{
			mode:  types.GameMode2v2,
			teams: map[int][]int{1: {0, 2}, 2: {1, 3}},
		}

		assert.False(t, g.SameTeam(0, 1))
		assert.False(t, g.SameTeam(0, 3))
		assert.False(t, g.SameTeam(2, 1))
		assert.False(t, g.SameTeam(2, 3))
	})
}

func TestGame_Allies(t *testing.T) {
	t.Run("Returns nil for 1v1", func(t *testing.T) {
		g := &game{mode: types.GameMode1v1}
		assert.Nil(t, g.Allies(0))
	})

	t.Run("Returns nil for FFA3", func(t *testing.T) {
		g := &game{mode: types.GameModeFFA3}
		assert.Nil(t, g.Allies(0))
	})

	t.Run("Returns teammate for 2v2", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockPlayer1 := mocks.NewMockPlayer(ctrl)
		mockPlayer2 := mocks.NewMockPlayer(ctrl)
		mockPlayer3 := mocks.NewMockPlayer(ctrl)
		mockPlayer4 := mocks.NewMockPlayer(ctrl)

		g := &game{
			board: &testBoardImpl{players: []board.Player{mockPlayer1, mockPlayer2, mockPlayer3, mockPlayer4}},
			mode:  types.GameMode2v2,
			teams: map[int][]int{1: {0, 2}, 2: {1, 3}},
		}

		allies0 := g.Allies(0)
		assert.Len(t, allies0, 1)
		assert.Equal(t, mockPlayer3, allies0[0])

		allies1 := g.Allies(1)
		assert.Len(t, allies1, 1)
		assert.Equal(t, mockPlayer4, allies1[0])
	})
}

func TestGame_Enemies(t *testing.T) {
	t.Run("1v1 returns opponent", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockPlayer1 := mocks.NewMockPlayer(ctrl)
		mockPlayer2 := mocks.NewMockPlayer(ctrl)

		g := &game{
			board:             &testBoardImpl{players: []board.Player{mockPlayer1, mockPlayer2}},
			mode:              types.GameMode1v1,
			eliminatedPlayers: make(map[int]bool),
		}

		enemies := g.Enemies(0)
		assert.Len(t, enemies, 1)
		assert.Equal(t, mockPlayer2, enemies[0])
	})

	t.Run("2v2 excludes teammates", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockPlayer1 := mocks.NewMockPlayer(ctrl)
		mockPlayer2 := mocks.NewMockPlayer(ctrl)
		mockPlayer3 := mocks.NewMockPlayer(ctrl)
		mockPlayer4 := mocks.NewMockPlayer(ctrl)

		g := &game{
			board:             &testBoardImpl{players: []board.Player{mockPlayer1, mockPlayer2, mockPlayer3, mockPlayer4}},
			mode:              types.GameMode2v2,
			teams:             map[int][]int{1: {0, 2}, 2: {1, 3}},
			eliminatedPlayers: make(map[int]bool),
		}

		enemies := g.Enemies(0) // Player1 (Team 1)
		assert.Len(t, enemies, 2)
		assert.Equal(t, mockPlayer2, enemies[0])
		assert.Equal(t, mockPlayer4, enemies[1])
	})

	t.Run("Excludes eliminated players", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockPlayer1 := mocks.NewMockPlayer(ctrl)
		mockPlayer2 := mocks.NewMockPlayer(ctrl)
		mockPlayer3 := mocks.NewMockPlayer(ctrl)

		g := &game{
			board:             &testBoardImpl{players: []board.Player{mockPlayer1, mockPlayer2, mockPlayer3}},
			mode:              types.GameModeFFA3,
			eliminatedPlayers: map[int]bool{1: true},
		}

		enemies := g.Enemies(0)
		assert.Len(t, enemies, 1)
		assert.Equal(t, mockPlayer3, enemies[0])
	})
}

func TestGame_getTargetPlayer(t *testing.T) {
	t.Run("Error when target not found", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockPlayer1 := mocks.NewMockPlayer(ctrl)
		mockPlayer1.EXPECT().Name().Return("Player1").AnyTimes()

		g := &game{
			board: &testBoardImpl{players: []board.Player{mockPlayer1}},
		}

		_, err := g.GetTargetPlayer("Player1", "Unknown")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "target player Unknown not found")
	})

	t.Run("Error when targeting self", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockPlayer1 := mocks.NewMockPlayer(ctrl)
		mockPlayer1.EXPECT().Name().Return("Player1").AnyTimes()

		g := &game{
			board:             &testBoardImpl{players: []board.Player{mockPlayer1}},
			eliminatedPlayers: make(map[int]bool),
		}

		_, err := g.GetTargetPlayer("Player1", "Player1")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "cannot attack yourself")
	})

	t.Run("Error when targeting ally in 2v2", func(t *testing.T) {
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
			board:             &testBoardImpl{players: []board.Player{mockPlayer1, mockPlayer2, mockPlayer3, mockPlayer4}},
			mode:              types.GameMode2v2,
			teams:             map[int][]int{1: {0, 2}, 2: {1, 3}},
			eliminatedPlayers: make(map[int]bool),
		}

		_, err := g.GetTargetPlayer("Player1", "Player3") // Player3 is teammate
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "cannot attack your ally")
	})

	t.Run("Error when targeting eliminated player", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockPlayer1 := mocks.NewMockPlayer(ctrl)
		mockPlayer2 := mocks.NewMockPlayer(ctrl)
		mockPlayer3 := mocks.NewMockPlayer(ctrl)
		mockPlayer1.EXPECT().Name().Return("Player1").AnyTimes()
		mockPlayer2.EXPECT().Name().Return("Player2").AnyTimes()
		mockPlayer3.EXPECT().Name().Return("Player3").AnyTimes()

		g := &game{
			board:             &testBoardImpl{players: []board.Player{mockPlayer1, mockPlayer2, mockPlayer3}},
			mode:              types.GameModeFFA3,
			eliminatedPlayers: map[int]bool{1: true},
		}

		_, err := g.GetTargetPlayer("Player1", "Player2")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "cannot attack eliminated player")
	})

	t.Run("Success targeting valid enemy", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockPlayer1 := mocks.NewMockPlayer(ctrl)
		mockPlayer2 := mocks.NewMockPlayer(ctrl)
		mockPlayer1.EXPECT().Name().Return("Player1").AnyTimes()
		mockPlayer2.EXPECT().Name().Return("Player2").AnyTimes()

		g := &game{
			board:             &testBoardImpl{players: []board.Player{mockPlayer1, mockPlayer2}},
			mode:              types.GameMode1v1,
			eliminatedPlayers: make(map[int]bool),
		}

		target, err := g.GetTargetPlayer("Player1", "Player2")
		assert.NoError(t, err)
		assert.Equal(t, mockPlayer2, target)
	})

	t.Run("Success targeting valid enemy in 2v2", func(t *testing.T) {
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
			board:             &testBoardImpl{players: []board.Player{mockPlayer1, mockPlayer2, mockPlayer3, mockPlayer4}},
			mode:              types.GameMode2v2,
			teams:             map[int][]int{1: {0, 2}, 2: {1, 3}},
			eliminatedPlayers: make(map[int]bool),
		}

		target, err := g.GetTargetPlayer("Player1", "Player2") // Player2 is enemy
		assert.NoError(t, err)
		assert.Equal(t, mockPlayer2, target)
	})
}

func TestGame_switchTurn(t *testing.T) {
	t.Run("Switches to next player", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockPlayer1 := mocks.NewMockPlayer(ctrl)
		mockPlayer2 := mocks.NewMockPlayer(ctrl)

		g := &game{
			board:             &testBoardImpl{players: []board.Player{mockPlayer1, mockPlayer2}},
			currentTurn:       0,
			turnState:         types.TurnState{HasMovedWarrior: true, HasTraded: true},
			currentAction:     types.PhaseTypeEndTurn,
			eliminatedPlayers: make(map[int]bool),
		}

		g.SwitchTurn()

		assert.Equal(t, 1, g.currentTurn)
		assert.False(t, g.turnState.HasMovedWarrior)
		assert.False(t, g.turnState.HasTraded)
		assert.Equal(t, types.PhaseTypeDrawCard, g.currentAction)
	})

	t.Run("Wraps around to first player", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockPlayer1 := mocks.NewMockPlayer(ctrl)
		mockPlayer2 := mocks.NewMockPlayer(ctrl)

		g := &game{
			board:             &testBoardImpl{players: []board.Player{mockPlayer1, mockPlayer2}},
			currentTurn:       1,
			eliminatedPlayers: make(map[int]bool),
		}

		g.SwitchTurn()

		assert.Equal(t, 0, g.currentTurn)
	})

	t.Run("Skips eliminated players", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockPlayer1 := mocks.NewMockPlayer(ctrl)
		mockPlayer2 := mocks.NewMockPlayer(ctrl)
		mockPlayer3 := mocks.NewMockPlayer(ctrl)

		g := &game{
			board:             &testBoardImpl{players: []board.Player{mockPlayer1, mockPlayer2, mockPlayer3}},
			currentTurn:       0,
			eliminatedPlayers: map[int]bool{1: true}, // Player2 eliminated
		}

		g.SwitchTurn()

		assert.Equal(t, 2, g.currentTurn) // Skips Player2
	})

	t.Run("Skips multiple eliminated players", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		players := make([]board.Player, 5)
		for i := 0; i < 5; i++ {
			players[i] = mocks.NewMockPlayer(ctrl)
		}

		g := &game{
			board:             &testBoardImpl{players: players},
			currentTurn:       0,
			eliminatedPlayers: map[int]bool{1: true, 2: true, 3: true},
		}

		g.SwitchTurn()

		assert.Equal(t, 4, g.currentTurn)
	})
}

func TestGame_PlayerIndex(t *testing.T) {
	t.Run("Returns correct index", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockPlayer1 := mocks.NewMockPlayer(ctrl)
		mockPlayer2 := mocks.NewMockPlayer(ctrl)
		mockPlayer1.EXPECT().Name().Return("Player1").AnyTimes()
		mockPlayer2.EXPECT().Name().Return("Player2").AnyTimes()

		g := &game{
			board: &testBoardImpl{players: []board.Player{mockPlayer1, mockPlayer2}},
		}

		assert.Equal(t, 0, g.PlayerIndex("Player1"))
		assert.Equal(t, 1, g.PlayerIndex("Player2"))
	})

	t.Run("Returns -1 for unknown player", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockPlayer1 := mocks.NewMockPlayer(ctrl)
		mockPlayer1.EXPECT().Name().Return("Player1").AnyTimes()

		g := &game{
			board: &testBoardImpl{players: []board.Player{mockPlayer1}},
		}

		assert.Equal(t, -1, g.PlayerIndex("Unknown"))
	})
}

func TestGame_GetPlayer(t *testing.T) {
	t.Run("Returns player by name", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockPlayer1 := mocks.NewMockPlayer(ctrl)
		mockPlayer2 := mocks.NewMockPlayer(ctrl)
		mockPlayer1.EXPECT().Name().Return("Player1").AnyTimes()
		mockPlayer2.EXPECT().Name().Return("Player2").AnyTimes()

		g := &game{
			board: &testBoardImpl{players: []board.Player{mockPlayer1, mockPlayer2}},
		}

		assert.Equal(t, mockPlayer1, g.GetPlayer("Player1"))
		assert.Equal(t, mockPlayer2, g.GetPlayer("Player2"))
	})

	t.Run("Returns nil for unknown player", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockPlayer1 := mocks.NewMockPlayer(ctrl)
		mockPlayer1.EXPECT().Name().Return("Player1").AnyTimes()

		g := &game{
			board: &testBoardImpl{players: []board.Player{mockPlayer1}},
		}

		assert.Nil(t, g.GetPlayer("Unknown"))
	})
}

func TestGame_OnCardMovedToPile(t *testing.T) {
	t.Run("Discards card to pile", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockDiscardPile := mocks.NewMockDiscardPile(ctrl)
		mockCard := mocks.NewMockCard(ctrl)
		mockDiscardPile.EXPECT().Discard(mockCard)

		g := &game{
			board: &testBoardImpl{discardPile: mockDiscardPile},
		}

		g.OnCardMovedToPile(mockCard)
	})
}

func TestGame_addToHistory(t *testing.T) {
	t.Run("Adds message to history", func(t *testing.T) {
		g := &game{history: []types.HistoryLine{}}
		g.AddHistory("test message", types.CategoryInfo)
		assert.Len(t, g.history, 1)
		assert.Equal(t, "test message", g.history[0].Msg)
		assert.Equal(t, types.CategoryInfo, g.history[0].Category)
	})

	t.Run("Does not add empty message", func(t *testing.T) {
		g := &game{history: []types.HistoryLine{}}
		g.AddHistory("", types.CategoryInfo)
		assert.Empty(t, g.history)
	})
}

// minDeckConfig returns a DeckConfig with just enough cards for a game to start.
func minDeckConfig() cards.DeckConfig {
	return cards.DeckConfig{Warriors: 2, ConstructionCards: 1}
}

func TestNewGame_CastleGoal_ZeroFallsBackToDefault(t *testing.T) {
	dealer := cards.NewDealer(minDeckConfig())
	g, err := NewGame([]string{"Alice", "Bob"}, types.GameMode1v1, dealer, 0)
	assert.NoError(t, err)
	assert.NotNil(t, g)
}

func TestNewGame_CastleGoal_NegativeFallsBackToDefault(t *testing.T) {
	dealer := cards.NewDealer(minDeckConfig())
	g, err := NewGame([]string{"Alice", "Bob"}, types.GameMode1v1, dealer, -5)
	assert.NoError(t, err)
	assert.NotNil(t, g)
}

func TestNewGame_CastleGoal_CustomGoalIsAccepted(t *testing.T) {
	dealer := cards.NewDealer(minDeckConfig())
	g, err := NewGame([]string{"Alice", "Bob"}, types.GameMode1v1, dealer, 15)
	assert.NoError(t, err)
	assert.NotNil(t, g)
}

func TestGame_validatePlayers(t *testing.T) {
	t.Run("1v1 requires 2 players", func(t *testing.T) {
		assert.NoError(t, validatePlayers(2, types.GameMode1v1))
		assert.Error(t, validatePlayers(1, types.GameMode1v1))
		assert.Error(t, validatePlayers(3, types.GameMode1v1))
	})

	t.Run("2v2 requires 4 players", func(t *testing.T) {
		assert.NoError(t, validatePlayers(4, types.GameMode2v2))
		assert.Error(t, validatePlayers(2, types.GameMode2v2))
		assert.Error(t, validatePlayers(5, types.GameMode2v2))
	})

	t.Run("FFA3 requires 3 players", func(t *testing.T) {
		assert.NoError(t, validatePlayers(3, types.GameModeFFA3))
		assert.Error(t, validatePlayers(2, types.GameModeFFA3))
	})

	t.Run("FFA5 requires 5 players", func(t *testing.T) {
		assert.NoError(t, validatePlayers(5, types.GameModeFFA5))
		assert.Error(t, validatePlayers(3, types.GameModeFFA5))
	})

	t.Run("Invalid game mode", func(t *testing.T) {
		err := validatePlayers(2, "invalid")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "invalid game mode")
	})
}
