package gameactions_test

import (
	"testing"

	"github.com/alelopezbcn/thecampaign/internal/domain/board"
	"github.com/alelopezbcn/thecampaign/internal/domain/cards"
	"github.com/alelopezbcn/thecampaign/internal/domain/gamestatus"
	"github.com/alelopezbcn/thecampaign/internal/domain/types"
	"github.com/alelopezbcn/thecampaign/test/mocks"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"

	"github.com/alelopezbcn/thecampaign/internal/domain/gameactions"
)

// validateBloodRain sets up and runs Validate for a blood rain action.
func validateBloodRain(
	t *testing.T, ctrl *gomock.Controller, targets []cards.Warrior,
) (gameactions.GameAction, *mocks.MockGame, *mocks.MockPlayer, *mocks.MockPlayer, *mocks.MockBloodRain) {
	t.Helper()

	mockGame := mocks.NewMockGame(ctrl)
	mockPlayer := mocks.NewMockPlayer(ctrl)
	mockTargetPlayer := mocks.NewMockPlayer(ctrl)
	mockField := mocks.NewMockField(ctrl)
	mockBloodRain := mocks.NewMockBloodRain(ctrl)

	mockGame.EXPECT().CurrentAction().Return(types.PhaseTypeAttack)
	mockGame.EXPECT().GetTargetPlayer("Player1", "Player2").Return(mockTargetPlayer, nil)
	mockTargetPlayer.EXPECT().Field().Return(mockField)
	mockField.EXPECT().Warriors().Return(targets)
	mockGame.EXPECT().CurrentPlayer().Return(mockPlayer)
	mockPlayer.EXPECT().GetCardFromHand("bloodrain-id").Return(mockBloodRain, true)

	action := gameactions.NewBloodRainAction("Player1", "Player2", "bloodrain-id")
	if err := action.Validate(mockGame); err != nil {
		t.Fatalf("validateBloodRain: unexpected error: %v", err)
	}
	return action, mockGame, mockPlayer, mockTargetPlayer, mockBloodRain
}

func TestBloodRainAction_PlayerName(t *testing.T) {
	action := gameactions.NewBloodRainAction("Player1", "Player2", "bloodrain-id")
	assert.Equal(t, "Player1", action.PlayerName())
}

func TestBloodRainAction_NextPhase(t *testing.T) {
	action := gameactions.NewBloodRainAction("Player1", "Player2", "bloodrain-id")
	assert.Equal(t, types.PhaseTypeSpySteal, action.NextPhase())
}

func TestBloodRainAction_Execute_Bloodlust(t *testing.T) {
	t.Run("Field warriors healed for each target killed", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		target1 := mocks.NewMockWarrior(ctrl)
		target2 := mocks.NewMockWarrior(ctrl)
		action, mockGame, mockPlayer, _, mockBloodRain := validateBloodRain(t, ctrl, []cards.Warrior{target1, target2})

		fieldWarrior := mocks.NewMockWarrior(ctrl)
		mockField := mocks.NewMockField(ctrl)
		expectedStatus := gamestatus.GameStatus{CurrentPlayer: "Player1"}

		mockGame.EXPECT().EventHandler().Return(bloodlustEvent())
		mockBloodRain.EXPECT().Attack([]cards.Warrior{target1, target2}).Return(nil)
		target1.EXPECT().Health().Return(0) // killed
		target2.EXPECT().Health().Return(0) // killed
		// 2 kills × 2 heal = 4 total heal
		mockPlayer.EXPECT().Field().Return(mockField)
		mockField.EXPECT().Warriors().Return([]cards.Warrior{fieldWarrior})
		fieldWarrior.EXPECT().HealBy(4)
		mockBloodRain.EXPECT().GetID().Return("bloodrain-id")
		mockPlayer.EXPECT().RemoveFromHand("bloodrain-id").Return(nil, nil)
		mockGame.EXPECT().OnCardMovedToPile(mockBloodRain)
		mockGame.EXPECT().AddHistory(gomock.Any(), gomock.Any())
		mockGame.EXPECT().Status(mockPlayer).Return(expectedStatus)

		result, statusFn, err := action.Execute(mockGame)

		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Equal(t, types.LastActionBloodRain, result.Action)
		assert.Equal(t, expectedStatus, statusFn())
	})

	t.Run("Partial kills: only killed targets contribute to heal", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		target1 := mocks.NewMockWarrior(ctrl)
		target2 := mocks.NewMockWarrior(ctrl)
		action, mockGame, mockPlayer, _, mockBloodRain := validateBloodRain(t, ctrl, []cards.Warrior{target1, target2})

		fieldWarrior := mocks.NewMockWarrior(ctrl)
		mockField := mocks.NewMockField(ctrl)
		expectedStatus := gamestatus.GameStatus{CurrentPlayer: "Player1"}

		mockGame.EXPECT().EventHandler().Return(bloodlustEvent())
		mockBloodRain.EXPECT().Attack([]cards.Warrior{target1, target2}).Return(nil)
		target1.EXPECT().Health().Return(0) // killed
		target2.EXPECT().Health().Return(3) // survived
		// 1 kill × 2 heal = 2
		mockPlayer.EXPECT().Field().Return(mockField)
		mockField.EXPECT().Warriors().Return([]cards.Warrior{fieldWarrior})
		fieldWarrior.EXPECT().HealBy(2)
		mockBloodRain.EXPECT().GetID().Return("bloodrain-id")
		mockPlayer.EXPECT().RemoveFromHand("bloodrain-id").Return(nil, nil)
		mockGame.EXPECT().OnCardMovedToPile(mockBloodRain)
		mockGame.EXPECT().AddHistory(gomock.Any(), gomock.Any())
		mockGame.EXPECT().Status(mockPlayer).Return(expectedStatus)

		result, statusFn, err := action.Execute(mockGame)

		assert.NoError(t, err)
		assert.Equal(t, types.LastActionBloodRain, result.Action)
		assert.Equal(t, expectedStatus, statusFn())
	})

	t.Run("No healing when all targets survive", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		target1 := mocks.NewMockWarrior(ctrl)
		action, mockGame, mockPlayer, _, mockBloodRain := validateBloodRain(t, ctrl, []cards.Warrior{target1})

		expectedStatus := gamestatus.GameStatus{CurrentPlayer: "Player1"}

		mockGame.EXPECT().EventHandler().Return(bloodlustEvent())
		mockBloodRain.EXPECT().Attack([]cards.Warrior{target1}).Return(nil)
		target1.EXPECT().Health().Return(5) // survived — HealBy must NOT be called
		mockBloodRain.EXPECT().GetID().Return("bloodrain-id")
		mockPlayer.EXPECT().RemoveFromHand("bloodrain-id").Return(nil, nil)
		mockGame.EXPECT().OnCardMovedToPile(mockBloodRain)
		mockGame.EXPECT().AddHistory(gomock.Any(), gomock.Any())
		mockGame.EXPECT().Status(mockPlayer).Return(expectedStatus)

		result, statusFn, err := action.Execute(mockGame)

		assert.NoError(t, err)
		assert.Equal(t, types.LastActionBloodRain, result.Action)
		assert.Equal(t, expectedStatus, statusFn())
	})

	t.Run("No healing when no bloodlust event", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		target1 := mocks.NewMockWarrior(ctrl)
		action, mockGame, mockPlayer, _, mockBloodRain := validateBloodRain(t, ctrl, []cards.Warrior{target1})

		expectedStatus := gamestatus.GameStatus{CurrentPlayer: "Player1"}

		mockGame.EXPECT().EventHandler().Return(calmEvent())
		mockBloodRain.EXPECT().Attack([]cards.Warrior{target1}).Return(nil)
		// bountyCards=0 and healAmount=0 → kills not counted → Health must NOT be called
		mockBloodRain.EXPECT().GetID().Return("bloodrain-id")
		mockPlayer.EXPECT().RemoveFromHand("bloodrain-id").Return(nil, nil)
		mockGame.EXPECT().OnCardMovedToPile(mockBloodRain)
		mockGame.EXPECT().AddHistory(gomock.Any(), gomock.Any())
		mockGame.EXPECT().Status(mockPlayer).Return(expectedStatus)

		result, statusFn, err := action.Execute(mockGame)

		assert.NoError(t, err)
		assert.Equal(t, types.LastActionBloodRain, result.Action)
		assert.Equal(t, expectedStatus, statusFn())
	})
}

func TestBloodRainAction_Execute_ChampionsBounty(t *testing.T) {
	t.Run("Bounty drawn when any target is killed and target player is top enemy", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		target1 := mocks.NewMockWarrior(ctrl)
		target2 := mocks.NewMockWarrior(ctrl)
		action, mockGame, mockPlayer, mockTargetPlayer, mockBloodRain := validateBloodRain(t, ctrl, []cards.Warrior{target1, target2})

		mockTargetField := mocks.NewMockField(ctrl)
		mockBountyCard := mocks.NewMockCard(ctrl)
		expectedStatus := gamestatus.GameStatus{CurrentPlayer: "Player1"}

		mockGame.EXPECT().EventHandler().Return(championsBountyEvent())
		// pre-kill snapshot: target has 2 warriors summing to 8 HP
		mockTargetPlayer.EXPECT().Field().Return(mockTargetField)
		mockTargetField.EXPECT().Warriors().Return([]cards.Warrior{target1, target2})
		target1.EXPECT().Health().Return(5) // pre-kill snapshot
		target2.EXPECT().Health().Return(3) // pre-kill snapshot
		mockBloodRain.EXPECT().Attack([]cards.Warrior{target1, target2}).Return(nil)
		// post-attack kills count
		target1.EXPECT().Health().Return(0) // killed
		target2.EXPECT().Health().Return(0) // killed
		mockBloodRain.EXPECT().GetID().Return("bloodrain-id")
		mockPlayer.EXPECT().RemoveFromHand("bloodrain-id").Return(nil, nil)
		mockGame.EXPECT().OnCardMovedToPile(mockBloodRain)
		mockGame.EXPECT().AddHistory(gomock.Any(), gomock.Any()).Times(3) // attack + kill bounty + hit-heal bounty
		// isTopEnemy: only enemy is target — trivially top (called twice: kill bounty + hit-heal)
		mockGame.EXPECT().PlayerIndex("Player1").Return(0).Times(2)
		mockGame.EXPECT().Enemies(0).Return([]board.Player{mockTargetPlayer}).Times(2)
		mockTargetPlayer.EXPECT().Name().Return("Player2").AnyTimes()
		mockGame.EXPECT().DrawCards(mockPlayer, 2).Return([]cards.Card{mockBountyCard}, nil)
		mockPlayer.EXPECT().TakeCards(mockBountyCard)
		mockPlayer.EXPECT().Name().Return("Player1").Times(2) // kill bounty + hit-heal history
		// hit-heal: heal +3 to a random attacker warrior
		mockAttackerField := mocks.NewMockField(ctrl)
		healWarrior := mocks.NewMockWarrior(ctrl)
		mockPlayer.EXPECT().Field().Return(mockAttackerField)
		mockAttackerField.EXPECT().Warriors().Return([]cards.Warrior{healWarrior})
		healWarrior.EXPECT().HealBy(3)
		mockGame.EXPECT().Status(mockPlayer, mockBountyCard).Return(expectedStatus)

		result, statusFn, err := action.Execute(mockGame)

		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Equal(t, types.LastActionBloodRain, result.Action)
		assert.Equal(t, "Player1", result.Attack.ChampionsBountyEarner)
		assert.Equal(t, 1, result.Attack.ChampionsBountyCards)
		assert.Equal(t, 3, result.Attack.ChampionsBountyHeal)
		assert.Equal(t, expectedStatus, statusFn())
	})

	t.Run("No bounty when all targets survive", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		target1 := mocks.NewMockWarrior(ctrl)
		action, mockGame, mockPlayer, mockTargetPlayer, mockBloodRain := validateBloodRain(t, ctrl, []cards.Warrior{target1})

		mockTargetField := mocks.NewMockField(ctrl)
		expectedStatus := gamestatus.GameStatus{CurrentPlayer: "Player1"}

		mockGame.EXPECT().EventHandler().Return(championsBountyEvent())
		// pre-kill snapshot (bountyCards=2, hitHeal=3, both >0)
		mockTargetPlayer.EXPECT().Field().Return(mockTargetField)
		mockTargetField.EXPECT().Warriors().Return([]cards.Warrior{target1})
		target1.EXPECT().Health().Return(5) // pre-kill snapshot
		mockBloodRain.EXPECT().Attack([]cards.Warrior{target1}).Return(nil)
		// post-attack: target survived — no kill bounty
		target1.EXPECT().Health().Return(3)
		mockBloodRain.EXPECT().GetID().Return("bloodrain-id")
		mockPlayer.EXPECT().RemoveFromHand("bloodrain-id").Return(nil, nil)
		mockGame.EXPECT().OnCardMovedToPile(mockBloodRain)
		mockGame.EXPECT().AddHistory(gomock.Any(), gomock.Any()).Times(2) // attack + hit-heal bounty
		// hit-heal: target player is top enemy → heal +3
		mockGame.EXPECT().PlayerIndex("Player1").Return(0)
		mockGame.EXPECT().Enemies(0).Return([]board.Player{mockTargetPlayer})
		mockTargetPlayer.EXPECT().Name().Return("Player2")
		mockAttackerField := mocks.NewMockField(ctrl)
		healWarrior := mocks.NewMockWarrior(ctrl)
		mockPlayer.EXPECT().Field().Return(mockAttackerField)
		mockAttackerField.EXPECT().Warriors().Return([]cards.Warrior{healWarrior})
		healWarrior.EXPECT().HealBy(3)
		mockPlayer.EXPECT().Name().Return("Player1")
		mockGame.EXPECT().Status(mockPlayer).Return(expectedStatus)

		result, statusFn, err := action.Execute(mockGame)
		_ = statusFn()

		assert.NoError(t, err)
		assert.Equal(t, "", result.Attack.ChampionsBountyEarner)
		assert.Equal(t, 0, result.Attack.ChampionsBountyCards)
		assert.Equal(t, 3, result.Attack.ChampionsBountyHeal)
	})
}

func TestBloodRainAction_Execute_ChampionsBounty_HitHeal(t *testing.T) {
	t.Run("Hit-heal triggers even when all targets are killed", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		target1 := mocks.NewMockWarrior(ctrl)
		action, mockGame, mockPlayer, mockTargetPlayer, mockBloodRain := validateBloodRain(t, ctrl, []cards.Warrior{target1})

		mockTargetField := mocks.NewMockField(ctrl)
		mockBountyCard := mocks.NewMockCard(ctrl)
		expectedStatus := gamestatus.GameStatus{CurrentPlayer: "Player1"}

		mockGame.EXPECT().EventHandler().Return(championsBountyEvent())
		// pre-kill snapshot
		mockTargetPlayer.EXPECT().Field().Return(mockTargetField)
		mockTargetField.EXPECT().Warriors().Return([]cards.Warrior{target1})
		target1.EXPECT().Health().Return(5) // pre-kill snapshot
		mockBloodRain.EXPECT().Attack([]cards.Warrior{target1}).Return(nil)
		// post-attack: killed
		target1.EXPECT().Health().Return(0)
		mockBloodRain.EXPECT().GetID().Return("bloodrain-id")
		mockPlayer.EXPECT().RemoveFromHand("bloodrain-id").Return(nil, nil)
		mockGame.EXPECT().OnCardMovedToPile(mockBloodRain)
		mockGame.EXPECT().AddHistory(gomock.Any(), gomock.Any()).Times(3) // attack + kill bounty + hit-heal bounty
		// kill bounty
		mockGame.EXPECT().PlayerIndex("Player1").Return(0).Times(2)
		mockGame.EXPECT().Enemies(0).Return([]board.Player{mockTargetPlayer}).Times(2)
		mockTargetPlayer.EXPECT().Name().Return("Player2").AnyTimes()
		mockGame.EXPECT().DrawCards(mockPlayer, 2).Return([]cards.Card{mockBountyCard}, nil)
		mockPlayer.EXPECT().TakeCards(mockBountyCard)
		mockPlayer.EXPECT().Name().Return("Player1").Times(2)
		// hit-heal
		mockAttackerField := mocks.NewMockField(ctrl)
		healWarrior := mocks.NewMockWarrior(ctrl)
		mockPlayer.EXPECT().Field().Return(mockAttackerField)
		mockAttackerField.EXPECT().Warriors().Return([]cards.Warrior{healWarrior})
		healWarrior.EXPECT().HealBy(3)
		mockGame.EXPECT().Status(mockPlayer, mockBountyCard).Return(expectedStatus)

		result, statusFn, err := action.Execute(mockGame)

		assert.NoError(t, err)
		assert.Equal(t, "Player1", result.Attack.ChampionsBountyEarner)
		assert.Equal(t, 1, result.Attack.ChampionsBountyCards)
		assert.Equal(t, 3, result.Attack.ChampionsBountyHeal)
		assert.Equal(t, expectedStatus, statusFn())
	})

	t.Run("Hit-heal skipped when target player is not top enemy", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		target1 := mocks.NewMockWarrior(ctrl)
		action, mockGame, mockPlayer, mockTargetPlayer, mockBloodRain := validateBloodRain(t, ctrl, []cards.Warrior{target1})

		mockTargetField := mocks.NewMockField(ctrl)
		expectedStatus := gamestatus.GameStatus{CurrentPlayer: "Player1"}

		mockGame.EXPECT().EventHandler().Return(championsBountyEvent())
		// pre-kill snapshot (target HP=3, player3 has 8)
		mockTargetPlayer.EXPECT().Field().Return(mockTargetField)
		mockTargetField.EXPECT().Warriors().Return([]cards.Warrior{target1})
		target1.EXPECT().Health().Return(3) // pre-kill snapshot
		mockBloodRain.EXPECT().Attack([]cards.Warrior{target1}).Return(nil)
		// post-attack: survived
		target1.EXPECT().Health().Return(2)
		mockBloodRain.EXPECT().GetID().Return("bloodrain-id")
		mockPlayer.EXPECT().RemoveFromHand("bloodrain-id").Return(nil, nil)
		mockGame.EXPECT().OnCardMovedToPile(mockBloodRain)
		mockGame.EXPECT().AddHistory(gomock.Any(), gomock.Any()) // attack only
		// hit-heal check: player3 has more HP → not top enemy
		mockGame.EXPECT().PlayerIndex("Player1").Return(0)
		mockPlayer3 := mocks.NewMockPlayer(ctrl)
		mockPlayer3Field := mocks.NewMockField(ctrl)
		mockGame.EXPECT().Enemies(0).Return([]board.Player{mockTargetPlayer, mockPlayer3})
		mockTargetPlayer.EXPECT().Name().Return("Player2")
		mockPlayer3.EXPECT().Name().Return("Player3")
		mockPlayer3.EXPECT().Field().Return(mockPlayer3Field)
		strongerWarrior := mocks.NewMockWarrior(ctrl)
		mockPlayer3Field.EXPECT().Warriors().Return([]cards.Warrior{strongerWarrior})
		strongerWarrior.EXPECT().Health().Return(8)
		// HealBy must NOT be called
		mockGame.EXPECT().Status(mockPlayer).Return(expectedStatus)

		result, statusFn, err := action.Execute(mockGame)
		_ = statusFn()

		assert.NoError(t, err)
		assert.Equal(t, 0, result.Attack.ChampionsBountyHeal)
	})
}
