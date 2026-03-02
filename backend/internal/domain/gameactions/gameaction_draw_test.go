package gameactions_test

import (
	"errors"
	"testing"

	"github.com/alelopezbcn/thecampaign/internal/domain/board"
	"github.com/alelopezbcn/thecampaign/internal/domain/cards"
	"github.com/alelopezbcn/thecampaign/internal/domain/gameactions"
	"github.com/alelopezbcn/thecampaign/internal/domain/gameevents"
	"github.com/alelopezbcn/thecampaign/internal/domain/gamestatus"
	"github.com/alelopezbcn/thecampaign/internal/domain/types"
	"github.com/alelopezbcn/thecampaign/test/mocks"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

// calmEvent returns an EventHandler with no active event (no extra draws, no modifiers).
func calmEvent() gameevents.EventHandler {
	return gameevents.NewHandler(types.ActiveEvent{})
}

// curseEvent returns an EventHandler for a Curse event with the given excluded weapon and modifier.
func curseEvent(excludedWeapon types.WeaponType, modifier int) gameevents.EventHandler {
	return gameevents.NewHandler(types.ActiveEvent{
		Type:                types.EventTypeCurse,
		CurseExcludedWeapon: excludedWeapon,
		CurseModifier:       modifier,
	})
}

// plagueEvent returns an EventHandler for a Plague event with the given HP modifier.
func plagueEvent(modifier int) gameevents.EventHandler {
	return gameevents.NewHandler(types.ActiveEvent{
		Type:           types.EventTypePlague,
		PlagueModifier: modifier,
	})
}

// harvestEvent returns an EventHandler for a Harvest event with the given construction value modifier.
func harvestEvent(modifier int) gameevents.EventHandler {
	return gameevents.NewHandler(types.ActiveEvent{
		Type:            types.EventTypeHarvest,
		HarvestModifier: modifier,
	})
}

func TestDrawCardAction_PlayerName(t *testing.T) {
	action := gameactions.NewDrawCardAction("Player1")
	assert.Equal(t, "Player1", action.PlayerName())
}

func TestDrawCardAction_Validate(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockGame := mocks.NewMockGame(ctrl)
	action := gameactions.NewDrawCardAction("Player1")
	err := action.Validate(mockGame)
	assert.NoError(t, err)
}

func TestDrawCardAction_NextPhase(t *testing.T) {
	action := gameactions.NewDrawCardAction("Player1")
	assert.Equal(t, types.PhaseTypeAttack, action.NextPhase())
}

func TestDrawCardAction_Execute(t *testing.T) {
	t.Run("Error when drawing fails", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockGame := mocks.NewMockGame(ctrl)
		mockPlayer1 := mocks.NewMockPlayer(ctrl)

		mockGame.EXPECT().CurrentPlayer().Return(mockPlayer1)
		mockGame.EXPECT().EventHandler().Return(calmEvent())
		mockGame.EXPECT().DrawCards(mockPlayer1, 1).Return(nil, errors.New("no cards left to draw"))

		action := gameactions.NewDrawCardAction("Player1")
		result, statusFn, err := action.Execute(mockGame)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "no cards left to draw")
		assert.Nil(t, result)
		assert.Nil(t, statusFn)
	})

	t.Run("Success drawing card", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockGame := mocks.NewMockGame(ctrl)
		mockPlayer1 := mocks.NewMockPlayer(ctrl)
		mockDrawnCard := mocks.NewMockCard(ctrl)
		expectedStatus := gamestatus.GameStatus{CurrentPlayer: "Player1"}

		mockGame.EXPECT().CurrentPlayer().Return(mockPlayer1)
		mockGame.EXPECT().EventHandler().Return(calmEvent())
		mockGame.EXPECT().DrawCards(mockPlayer1, 1).Return([]cards.Card{mockDrawnCard}, nil)
		mockPlayer1.EXPECT().TakeCards(mockDrawnCard).Return(true)
		mockPlayer1.EXPECT().Name().Return("Player1")
		mockGame.EXPECT().AddHistory(gomock.Any(), gomock.Any())
		mockGame.EXPECT().Status(mockPlayer1, mockDrawnCard).Return(expectedStatus)

		action := gameactions.NewDrawCardAction("Player1")
		result, statusFn, err := action.Execute(mockGame)

		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Equal(t, types.LastActionDraw, result.Action)
		assert.Equal(t, expectedStatus, statusFn())
	})

	t.Run("Hand limit exceeded returns result without error", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockGame := mocks.NewMockGame(ctrl)
		mockPlayer1 := mocks.NewMockPlayer(ctrl)
		expectedStatus := gamestatus.GameStatus{CurrentPlayer: "Player1"}

		mockGame.EXPECT().CurrentPlayer().Return(mockPlayer1)
		mockGame.EXPECT().EventHandler().Return(calmEvent())
		mockGame.EXPECT().DrawCards(mockPlayer1, 1).Return(nil, board.ErrHandLimitExceeded)
		mockPlayer1.EXPECT().Name().Return("Player1")
		mockGame.EXPECT().AddHistory(gomock.Any(), gomock.Any())
		mockGame.EXPECT().Status(mockPlayer1).Return(expectedStatus)

		action := gameactions.NewDrawCardAction("Player1")
		result, statusFn, err := action.Execute(mockGame)

		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Equal(t, types.LastActionType(""), result.Action)
		assert.Equal(t, expectedStatus, statusFn())
	})

	t.Run("History updated on successful draw", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockGame := mocks.NewMockGame(ctrl)
		mockPlayer1 := mocks.NewMockPlayer(ctrl)
		mockDrawnCard := mocks.NewMockCard(ctrl)

		mockGame.EXPECT().CurrentPlayer().Return(mockPlayer1)
		mockGame.EXPECT().EventHandler().Return(calmEvent())
		mockGame.EXPECT().DrawCards(mockPlayer1, 1).Return([]cards.Card{mockDrawnCard}, nil)
		mockPlayer1.EXPECT().TakeCards(mockDrawnCard).Return(true)
		mockPlayer1.EXPECT().Name().Return("Player1")

		var capturedMsg string
		mockGame.EXPECT().AddHistory(gomock.Any(), gomock.Any()).Do(func(msg string, _ types.Category) {
			capturedMsg = msg
		})

		action := gameactions.NewDrawCardAction("Player1")
		_, statusFn, err := action.Execute(mockGame)

		assert.NoError(t, err)
		assert.NotNil(t, statusFn)
		assert.Contains(t, capturedMsg, "drew")
		assert.Contains(t, capturedMsg, "Player1")
	})

	t.Run("History updated when hand limit exceeded", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockGame := mocks.NewMockGame(ctrl)
		mockPlayer1 := mocks.NewMockPlayer(ctrl)

		mockGame.EXPECT().CurrentPlayer().Return(mockPlayer1)
		mockGame.EXPECT().EventHandler().Return(calmEvent())
		mockGame.EXPECT().DrawCards(mockPlayer1, 1).Return(nil, board.ErrHandLimitExceeded)
		mockPlayer1.EXPECT().Name().Return("Player1")

		var capturedMsg string
		mockGame.EXPECT().AddHistory(gomock.Any(), gomock.Any()).Do(func(msg string, _ types.Category) {
			capturedMsg = msg
		})

		action := gameactions.NewDrawCardAction("Player1")
		_, statusFn, err := action.Execute(mockGame)

		assert.NoError(t, err)
		assert.NotNil(t, statusFn)
		assert.Contains(t, capturedMsg, "can't take more cards")
	})
}

func TestDrawCardAction_Abundance(t *testing.T) {
	abundanceEvent := gameevents.NewHandler(types.ActiveEvent{Type: types.EventTypeAbundance})

	t.Run("Abundance draws 2 cards", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockGame := mocks.NewMockGame(ctrl)
		mockPlayer1 := mocks.NewMockPlayer(ctrl)
		card1 := mocks.NewMockCard(ctrl)
		card2 := mocks.NewMockCard(ctrl)
		expectedStatus := gamestatus.GameStatus{CurrentPlayer: "Player1"}

		mockGame.EXPECT().CurrentPlayer().Return(mockPlayer1)
		mockGame.EXPECT().EventHandler().Return(abundanceEvent)
		mockGame.EXPECT().DrawCards(mockPlayer1, 2).Return([]cards.Card{card1, card2}, nil)
		mockPlayer1.EXPECT().TakeCards(card1, card2).Return(true)
		mockPlayer1.EXPECT().Name().Return("Player1")
		mockGame.EXPECT().AddHistory(gomock.Any(), gomock.Any())
		mockGame.EXPECT().Status(mockPlayer1, card1, card2).Return(expectedStatus)

		action := gameactions.NewDrawCardAction("Player1")
		result, statusFn, err := action.Execute(mockGame)

		assert.NoError(t, err)
		assert.Equal(t, types.LastActionDraw, result.Action)
		assert.Equal(t, expectedStatus, statusFn())
	})

	t.Run("Abundance falls back to 1 card when hand is nearly full", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockGame := mocks.NewMockGame(ctrl)
		mockPlayer1 := mocks.NewMockPlayer(ctrl)
		card1 := mocks.NewMockCard(ctrl)
		expectedStatus := gamestatus.GameStatus{CurrentPlayer: "Player1"}

		mockGame.EXPECT().CurrentPlayer().Return(mockPlayer1)
		mockGame.EXPECT().EventHandler().Return(abundanceEvent)
		// First attempt (2 cards) fails — hand has room for only 1
		mockGame.EXPECT().DrawCards(mockPlayer1, 2).Return(nil, board.ErrHandLimitExceeded)
		// Fallback to 1 card succeeds
		mockGame.EXPECT().DrawCards(mockPlayer1, 1).Return([]cards.Card{card1}, nil)
		mockPlayer1.EXPECT().TakeCards(card1).Return(true)
		mockPlayer1.EXPECT().Name().Return("Player1")
		mockGame.EXPECT().AddHistory(gomock.Any(), gomock.Any())
		mockGame.EXPECT().Status(mockPlayer1, card1).Return(expectedStatus)

		action := gameactions.NewDrawCardAction("Player1")
		result, statusFn, err := action.Execute(mockGame)

		assert.NoError(t, err)
		assert.Equal(t, types.LastActionDraw, result.Action)
		assert.Equal(t, expectedStatus, statusFn())
	})

	t.Run("Abundance history message mentions extra cards", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockGame := mocks.NewMockGame(ctrl)
		mockPlayer1 := mocks.NewMockPlayer(ctrl)
		card1 := mocks.NewMockCard(ctrl)
		card2 := mocks.NewMockCard(ctrl)

		mockGame.EXPECT().CurrentPlayer().Return(mockPlayer1)
		mockGame.EXPECT().EventHandler().Return(abundanceEvent)
		mockGame.EXPECT().DrawCards(mockPlayer1, 2).Return([]cards.Card{card1, card2}, nil)
		mockPlayer1.EXPECT().TakeCards(card1, card2).Return(true)
		mockPlayer1.EXPECT().Name().Return("Player1")

		var capturedMsg string
		mockGame.EXPECT().AddHistory(gomock.Any(), gomock.Any()).Do(func(msg string, _ types.Category) {
			capturedMsg = msg
		})
		mockGame.EXPECT().Status(mockPlayer1, card1, card2).Return(gamestatus.GameStatus{}).AnyTimes()

		action := gameactions.NewDrawCardAction("Player1")
		_, statusFn, err := action.Execute(mockGame)
		statusFn() // trigger the Status call

		assert.NoError(t, err)
		assert.Contains(t, capturedMsg, "2 cards")
		assert.Contains(t, capturedMsg, "Abundance")
	})
}

func TestDrawCardAction_Plague(t *testing.T) {
	t.Run("Plague damages warriors by modifier amount", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockGame := mocks.NewMockGame(ctrl)
		mockPlayer1 := mocks.NewMockPlayer(ctrl)
		mockWarrior := mocks.NewMockWarrior(ctrl)
		mockField := mocks.NewMockField(ctrl)
		mockDrawnCard := mocks.NewMockCard(ctrl)
		expectedStatus := gamestatus.GameStatus{CurrentPlayer: "Player1"}

		mockGame.EXPECT().CurrentPlayer().Return(mockPlayer1)
		mockGame.EXPECT().EventHandler().Return(plagueEvent(-2))
		// Plague: warrior has 10 HP, mod=-2 → safe (10-2=8 ≥ 1) → HealBy(-2)
		mockPlayer1.EXPECT().Field().Return(mockField)
		mockField.EXPECT().Warriors().Return([]cards.Warrior{mockWarrior})
		mockWarrior.EXPECT().Health().Return(10)
		mockWarrior.EXPECT().HealBy(-2)
		mockGame.EXPECT().AddHistory(gomock.Any(), types.CategoryInfo) // Plague history
		// Draw phase
		mockGame.EXPECT().DrawCards(mockPlayer1, 1).Return([]cards.Card{mockDrawnCard}, nil)
		mockPlayer1.EXPECT().TakeCards(mockDrawnCard).Return(true)
		mockPlayer1.EXPECT().Name().Return("Player1")
		mockGame.EXPECT().AddHistory(gomock.Any(), gomock.Any()) // draw history
		mockGame.EXPECT().Status(mockPlayer1, mockDrawnCard).Return(expectedStatus)

		action := gameactions.NewDrawCardAction("Player1")
		result, statusFn, err := action.Execute(mockGame)

		assert.NoError(t, err)
		assert.Equal(t, types.LastActionDraw, result.Action)
		assert.Equal(t, expectedStatus, statusFn())
	})

	t.Run("Plague cannot reduce warrior HP below 1", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockGame := mocks.NewMockGame(ctrl)
		mockPlayer1 := mocks.NewMockPlayer(ctrl)
		mockWarrior := mocks.NewMockWarrior(ctrl)
		mockField := mocks.NewMockField(ctrl)
		mockDrawnCard := mocks.NewMockCard(ctrl)
		expectedStatus := gamestatus.GameStatus{CurrentPlayer: "Player1"}

		mockGame.EXPECT().CurrentPlayer().Return(mockPlayer1)
		mockGame.EXPECT().EventHandler().Return(plagueEvent(-3))
		// warrior has 2 HP, mod=-3 → would go to -1, clamped → HealBy(-1) → 1 HP
		mockPlayer1.EXPECT().Field().Return(mockField)
		mockField.EXPECT().Warriors().Return([]cards.Warrior{mockWarrior})
		mockWarrior.EXPECT().Health().Return(2)
		mockWarrior.EXPECT().HealBy(-1) // clamped: 1 - 2 = -1
		mockGame.EXPECT().AddHistory(gomock.Any(), types.CategoryInfo)
		mockGame.EXPECT().DrawCards(mockPlayer1, 1).Return([]cards.Card{mockDrawnCard}, nil)
		mockPlayer1.EXPECT().TakeCards(mockDrawnCard).Return(true)
		mockPlayer1.EXPECT().Name().Return("Player1")
		mockGame.EXPECT().AddHistory(gomock.Any(), gomock.Any())
		mockGame.EXPECT().Status(mockPlayer1, mockDrawnCard).Return(expectedStatus)

		action := gameactions.NewDrawCardAction("Player1")
		result, statusFn, err := action.Execute(mockGame)

		assert.NoError(t, err)
		assert.Equal(t, types.LastActionDraw, result.Action)
		assert.Equal(t, expectedStatus, statusFn())
	})

	t.Run("Plague heals warriors with positive modifier", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockGame := mocks.NewMockGame(ctrl)
		mockPlayer1 := mocks.NewMockPlayer(ctrl)
		mockWarrior := mocks.NewMockWarrior(ctrl)
		mockField := mocks.NewMockField(ctrl)
		mockDrawnCard := mocks.NewMockCard(ctrl)
		expectedStatus := gamestatus.GameStatus{CurrentPlayer: "Player1"}

		mockGame.EXPECT().CurrentPlayer().Return(mockPlayer1)
		mockGame.EXPECT().EventHandler().Return(plagueEvent(3))
		// warrior has mod=+3 → HealBy(3) directly (Health not checked for positive mod)
		mockPlayer1.EXPECT().Field().Return(mockField)
		mockField.EXPECT().Warriors().Return([]cards.Warrior{mockWarrior})
		mockWarrior.EXPECT().HealBy(3)
		mockGame.EXPECT().AddHistory(gomock.Any(), types.CategoryInfo)
		mockGame.EXPECT().DrawCards(mockPlayer1, 1).Return([]cards.Card{mockDrawnCard}, nil)
		mockPlayer1.EXPECT().TakeCards(mockDrawnCard).Return(true)
		mockPlayer1.EXPECT().Name().Return("Player1")
		mockGame.EXPECT().AddHistory(gomock.Any(), gomock.Any())
		mockGame.EXPECT().Status(mockPlayer1, mockDrawnCard).Return(expectedStatus)

		action := gameactions.NewDrawCardAction("Player1")
		result, statusFn, err := action.Execute(mockGame)

		assert.NoError(t, err)
		assert.Equal(t, types.LastActionDraw, result.Action)
		assert.Equal(t, expectedStatus, statusFn())
	})

	t.Run("Plague history message is added before draw history", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockGame := mocks.NewMockGame(ctrl)
		mockPlayer1 := mocks.NewMockPlayer(ctrl)
		mockField := mocks.NewMockField(ctrl)
		mockDrawnCard := mocks.NewMockCard(ctrl)

		mockGame.EXPECT().CurrentPlayer().Return(mockPlayer1)
		mockGame.EXPECT().EventHandler().Return(plagueEvent(-2))
		mockPlayer1.EXPECT().Field().Return(mockField)
		mockField.EXPECT().Warriors().Return([]cards.Warrior{}) // no warriors
		var capturedPlague string
		mockGame.EXPECT().AddHistory(gomock.Any(), types.CategoryInfo).Do(func(msg string, _ types.Category) {
			capturedPlague = msg
		})
		mockGame.EXPECT().DrawCards(mockPlayer1, 1).Return([]cards.Card{mockDrawnCard}, nil)
		mockPlayer1.EXPECT().TakeCards(mockDrawnCard).Return(true)
		mockPlayer1.EXPECT().Name().Return("Player1")
		mockGame.EXPECT().AddHistory(gomock.Any(), gomock.Any()) // draw history
		mockGame.EXPECT().Status(mockPlayer1, mockDrawnCard).Return(gamestatus.GameStatus{}).AnyTimes()

		action := gameactions.NewDrawCardAction("Player1")
		_, statusFn, err := action.Execute(mockGame)
		statusFn()

		assert.NoError(t, err)
		assert.Contains(t, capturedPlague, "warriors")
		assert.Contains(t, capturedPlague, "HP")
	})

	t.Run("Plague warrior at exactly 1 HP with negative modifier takes no further damage", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockGame := mocks.NewMockGame(ctrl)
		mockPlayer1 := mocks.NewMockPlayer(ctrl)
		mockWarrior := mocks.NewMockWarrior(ctrl)
		mockField := mocks.NewMockField(ctrl)
		mockDrawnCard := mocks.NewMockCard(ctrl)

		mockGame.EXPECT().CurrentPlayer().Return(mockPlayer1)
		mockGame.EXPECT().EventHandler().Return(plagueEvent(-3))
		// warrior is already at 1 HP, mod=-3 → safeAmount = 1-1 = 0 → HealBy NOT called
		mockPlayer1.EXPECT().Field().Return(mockField)
		mockField.EXPECT().Warriors().Return([]cards.Warrior{mockWarrior})
		mockWarrior.EXPECT().Health().Return(1)
		// HealBy(0) is skipped — no expectation needed
		mockGame.EXPECT().AddHistory(gomock.Any(), types.CategoryInfo)
		mockGame.EXPECT().DrawCards(mockPlayer1, 1).Return([]cards.Card{mockDrawnCard}, nil)
		mockPlayer1.EXPECT().TakeCards(mockDrawnCard).Return(true)
		mockPlayer1.EXPECT().Name().Return("Player1")
		mockGame.EXPECT().AddHistory(gomock.Any(), gomock.Any())
		mockGame.EXPECT().Status(mockPlayer1, mockDrawnCard).Return(gamestatus.GameStatus{})

		action := gameactions.NewDrawCardAction("Player1")
		result, statusFn, err := action.Execute(mockGame)

		assert.NoError(t, err)
		assert.Equal(t, types.LastActionDraw, result.Action)
		statusFn()
	})
}
