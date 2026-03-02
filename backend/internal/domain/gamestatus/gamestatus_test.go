package gamestatus_test

import (
	"testing"

	"github.com/alelopezbcn/thecampaign/internal/domain/cards"
	"github.com/alelopezbcn/thecampaign/internal/domain/gamestatus"
	"github.com/alelopezbcn/thecampaign/internal/domain/types"
	"github.com/stretchr/testify/assert"
)

// minimalDTO returns a GameStatusDTO with the bare minimum required to call NewGameStatus
// without panicking. Callers are expected to override fields as needed.
func minimalDTO(viewerName string) gamestatus.GameStatusDTO {
	return gamestatus.GameStatusDTO{
		Viewer:       gamestatus.ViewerInput{Name: viewerName},
		PlayersNames: []string{viewerName},
		TurnPlayer:   viewerName,
	}
}

// ──────────────────────────────────────────────────────────────────────────────
// DesertionNotification tests
// ──────────────────────────────────────────────────────────────────────────────

func TestNewGameStatus_DesertionNotification_SetForVictim(t *testing.T) {
	warrior := cards.NewKnight("K1")

	dto := minimalDTO("Player2") // Player2 is the viewer (victim)
	dto.LastAction = types.LastActionDesertion
	dto.DeserterFromPlayer = "Player2" // warrior was stolen FROM Player2
	dto.DeserterWarrior = warrior
	dto.CurrentPlayerName = "Player1" // Player1 made the move

	gs := gamestatus.NewGameStatus(dto)

	assert.NotNil(t, gs.DesertionNotification, "victim should receive a DesertionNotification")
	assert.Equal(t, "Player1", gs.DesertionNotification.StolenBy)
	assert.Equal(t, "K1", gs.DesertionNotification.WarriorCard.CardID)
	assert.Equal(t, gamestatus.CardTypeKnight, gs.DesertionNotification.WarriorCard.CardType)
}

func TestNewGameStatus_DesertionNotification_NilForAttacker(t *testing.T) {
	warrior := cards.NewKnight("K1")

	dto := minimalDTO("Player1") // Player1 is the viewer (attacker/deserted-to)
	dto.LastAction = types.LastActionDesertion
	dto.DeserterFromPlayer = "Player2"
	dto.DeserterWarrior = warrior
	dto.CurrentPlayerName = "Player1"

	gs := gamestatus.NewGameStatus(dto)

	assert.Nil(t, gs.DesertionNotification, "attacker should NOT receive a DesertionNotification")
}

func TestNewGameStatus_DesertionNotification_NilForThirdPlayer(t *testing.T) {
	warrior := cards.NewKnight("K1")

	dto := minimalDTO("Player3") // Player3 is a spectator
	dto.LastAction = types.LastActionDesertion
	dto.DeserterFromPlayer = "Player2"
	dto.DeserterWarrior = warrior
	dto.CurrentPlayerName = "Player1"

	gs := gamestatus.NewGameStatus(dto)

	assert.Nil(t, gs.DesertionNotification, "bystander should NOT receive a DesertionNotification")
}

func TestNewGameStatus_DesertionNotification_NilWhenLastActionDiffers(t *testing.T) {
	warrior := cards.NewKnight("K1")

	dto := minimalDTO("Player2")
	dto.LastAction = types.LastActionSpy // different last action
	dto.DeserterFromPlayer = "Player2"
	dto.DeserterWarrior = warrior
	dto.CurrentPlayerName = "Player1"

	gs := gamestatus.NewGameStatus(dto)

	assert.Nil(t, gs.DesertionNotification)
}

func TestNewGameStatus_DesertionNotification_NilWhenWarriorIsNil(t *testing.T) {
	dto := minimalDTO("Player2")
	dto.LastAction = types.LastActionDesertion
	dto.DeserterFromPlayer = "Player2"
	dto.DeserterWarrior = nil // no warrior set
	dto.CurrentPlayerName = "Player1"

	gs := gamestatus.NewGameStatus(dto)

	assert.Nil(t, gs.DesertionNotification)
}

// ──────────────────────────────────────────────────────────────────────────────
// AmbushTriggered tests
// ──────────────────────────────────────────────────────────────────────────────

func TestNewGameStatus_AmbushTriggered_SetForAttacker(t *testing.T) {
	dto := minimalDTO("Player1") // Player1 is the attacker
	dto.LastAction = types.LastActionAmbush
	dto.AmbushAttackerName = "Player1"
	dto.LastAttackTargetPlayer = "Player2"
	dto.AmbushEffect = types.AmbushEffectCancelAttack

	gs := gamestatus.NewGameStatus(dto)

	assert.NotNil(t, gs.AmbushTriggered, "attacker should receive AmbushTriggered notification")
	assert.Equal(t, types.AmbushEffectCancelAttack, gs.AmbushTriggered.Effect)
	assert.Equal(t, "Attack Cancelled", gs.AmbushTriggered.EffectDisplay)
}

func TestNewGameStatus_AmbushTriggered_SetForDefender(t *testing.T) {
	dto := minimalDTO("Player2") // Player2 is the defender
	dto.LastAction = types.LastActionAmbush
	dto.AmbushAttackerName = "Player1"
	dto.LastAttackTargetPlayer = "Player2"
	dto.AmbushEffect = types.AmbushEffectStealWeapon

	gs := gamestatus.NewGameStatus(dto)

	assert.NotNil(t, gs.AmbushTriggered, "defender should receive AmbushTriggered notification")
	assert.Equal(t, types.AmbushEffectStealWeapon, gs.AmbushTriggered.Effect)
	assert.Equal(t, "Weapon Stolen", gs.AmbushTriggered.EffectDisplay)
}

func TestNewGameStatus_AmbushTriggered_NilForThirdPlayer(t *testing.T) {
	dto := minimalDTO("Player3") // Player3 is a bystander
	dto.LastAction = types.LastActionAmbush
	dto.AmbushAttackerName = "Player1"
	dto.LastAttackTargetPlayer = "Player2"
	dto.AmbushEffect = types.AmbushEffectReflectDamage

	gs := gamestatus.NewGameStatus(dto)

	assert.Nil(t, gs.AmbushTriggered, "bystander should NOT receive AmbushTriggered notification")
}

func TestNewGameStatus_AmbushTriggered_NilWhenLastActionDiffers(t *testing.T) {
	dto := minimalDTO("Player1")
	dto.LastAction = types.LastActionAttack // different last action
	dto.AmbushAttackerName = "Player1"
	dto.LastAttackTargetPlayer = "Player2"
	dto.AmbushEffect = types.AmbushEffectCancelAttack

	gs := gamestatus.NewGameStatus(dto)

	assert.Nil(t, gs.AmbushTriggered)
}

func TestNewGameStatus_AmbushTriggered_NilWhenAttackerNameEmpty(t *testing.T) {
	dto := minimalDTO("Player1")
	dto.LastAction = types.LastActionAmbush
	dto.AmbushAttackerName = "" // no attacker name
	dto.LastAttackTargetPlayer = "Player2"
	dto.AmbushEffect = types.AmbushEffectCancelAttack

	gs := gamestatus.NewGameStatus(dto)

	assert.Nil(t, gs.AmbushTriggered)
}

func TestNewGameStatus_AmbushTriggered_AllEffectsSetCorrectly(t *testing.T) {
	tests := []struct {
		effect      types.AmbushEffect
		wantDisplay string
	}{
		{types.AmbushEffectReflectDamage, "Reflect Damage"},
		{types.AmbushEffectCancelAttack, "Attack Cancelled"},
		{types.AmbushEffectStealWeapon, "Weapon Stolen"},
		{types.AmbushEffectDrainLife, "Drain Life"},
		{types.AmbushEffectInstantKill, "Instant Kill"},
	}

	for _, tt := range tests {
		t.Run(tt.wantDisplay, func(t *testing.T) {
			dto := minimalDTO("Player1")
			dto.LastAction = types.LastActionAmbush
			dto.AmbushAttackerName = "Player1"
			dto.LastAttackTargetPlayer = "Player2"
			dto.AmbushEffect = tt.effect

			gs := gamestatus.NewGameStatus(dto)

			assert.NotNil(t, gs.AmbushTriggered)
			assert.Equal(t, tt.effect, gs.AmbushTriggered.Effect)
			assert.Equal(t, tt.wantDisplay, gs.AmbushTriggered.EffectDisplay)
		})
	}
}

// ──────────────────────────────────────────────────────────────────────────────
// CurrentPlayerAmbushInField and OpponentStatus.AmbushInField tests
// ──────────────────────────────────────────────────────────────────────────────

func TestNewGameStatus_CurrentPlayerAmbushInField(t *testing.T) {
	t.Run("True when viewer field has ambush", func(t *testing.T) {
		dto := minimalDTO("Player1")
		dto.Viewer.Field.HasAmbush = true

		gs := gamestatus.NewGameStatus(dto)

		assert.True(t, gs.CurrentPlayerAmbushInField)
	})

	t.Run("False when viewer field has no ambush", func(t *testing.T) {
		dto := minimalDTO("Player1")
		dto.Viewer.Field.HasAmbush = false

		gs := gamestatus.NewGameStatus(dto)

		assert.False(t, gs.CurrentPlayerAmbushInField)
	})
}

func TestNewGameStatus_OpponentStatus_AmbushInField(t *testing.T) {
	t.Run("True when opponent field has ambush", func(t *testing.T) {
		dto := minimalDTO("Player1")
		dto.Opponents = []gamestatus.OpponentInput{
			{Name: "Player2", Field: gamestatus.FieldInput{HasAmbush: true}},
		}

		gs := gamestatus.NewGameStatus(dto)

		assert.Len(t, gs.Opponents, 1)
		assert.True(t, gs.Opponents[0].AmbushInField)
	})

	t.Run("False when opponent field has no ambush", func(t *testing.T) {
		dto := minimalDTO("Player1")
		dto.Opponents = []gamestatus.OpponentInput{
			{Name: "Player2", Field: gamestatus.FieldInput{HasAmbush: false}},
		}

		gs := gamestatus.NewGameStatus(dto)

		assert.Len(t, gs.Opponents, 1)
		assert.False(t, gs.Opponents[0].AmbushInField)
	})
}

// ──────────────────────────────────────────────────────────────────────────────
// StolenFromYouCard tests
// ──────────────────────────────────────────────────────────────────────────────

func TestNewGameStatus_StolenFromYouCard_SetForVictim(t *testing.T) {
	stolen := cards.NewKnight("K1")

	dto := minimalDTO("Player2") // Player2 is the viewer (victim)
	dto.LastAction = types.LastActionSteal
	dto.StolenFrom = "Player2"
	dto.StolenCard = stolen

	gs := gamestatus.NewGameStatus(dto)

	assert.Len(t, gs.StolenFromYouCard, 1)
	assert.Equal(t, "K1", gs.StolenFromYouCard[0].CardID)
}

func TestNewGameStatus_StolenFromYouCard_NilForAttacker(t *testing.T) {
	dto := minimalDTO("Player1") // Player1 is the viewer (thief)
	dto.LastAction = types.LastActionSteal
	dto.StolenFrom = "Player2"
	dto.StolenCard = cards.NewKnight("K1")

	gs := gamestatus.NewGameStatus(dto)

	assert.Empty(t, gs.StolenFromYouCard)
}

func TestNewGameStatus_StolenFromYouCard_NilWhenLastActionDiffers(t *testing.T) {
	dto := minimalDTO("Player2")
	dto.LastAction = types.LastActionSpy
	dto.StolenFrom = "Player2"
	dto.StolenCard = cards.NewKnight("K1")

	gs := gamestatus.NewGameStatus(dto)

	assert.Empty(t, gs.StolenFromYouCard)
}

func TestNewGameStatus_StolenFromYouCard_NilWhenStolenFromEmpty(t *testing.T) {
	dto := minimalDTO("Player2")
	dto.LastAction = types.LastActionSteal
	dto.StolenFrom = ""
	dto.StolenCard = cards.NewKnight("K1")

	gs := gamestatus.NewGameStatus(dto)

	assert.Empty(t, gs.StolenFromYouCard)
}

func TestNewGameStatus_StolenFromYouCard_NilWhenCardIsNil(t *testing.T) {
	dto := minimalDTO("Player2")
	dto.LastAction = types.LastActionSteal
	dto.StolenFrom = "Player2"
	dto.StolenCard = nil

	gs := gamestatus.NewGameStatus(dto)

	assert.Empty(t, gs.StolenFromYouCard)
}

// ──────────────────────────────────────────────────────────────────────────────
// SabotagedFromYouCard tests
// ──────────────────────────────────────────────────────────────────────────────

func TestNewGameStatus_SabotagedFromYouCard_SetForVictim(t *testing.T) {
	sabotaged := cards.NewSword("S1", 7)

	dto := minimalDTO("Player2") // Player2 is the viewer (victim)
	dto.LastAction = types.LastActionSabotage
	dto.SabotagedFrom = "Player2"
	dto.SabotagedCard = sabotaged

	gs := gamestatus.NewGameStatus(dto)

	assert.Len(t, gs.SabotagedFromYouCard, 1)
	assert.Equal(t, "S1", gs.SabotagedFromYouCard[0].CardID)
}

func TestNewGameStatus_SabotagedFromYouCard_NilForAttacker(t *testing.T) {
	dto := minimalDTO("Player1") // Player1 is the viewer (saboteur)
	dto.LastAction = types.LastActionSabotage
	dto.SabotagedFrom = "Player2"
	dto.SabotagedCard = cards.NewSword("S1", 7)

	gs := gamestatus.NewGameStatus(dto)

	assert.Empty(t, gs.SabotagedFromYouCard)
}

func TestNewGameStatus_SabotagedFromYouCard_NilWhenLastActionDiffers(t *testing.T) {
	dto := minimalDTO("Player2")
	dto.LastAction = types.LastActionSteal
	dto.SabotagedFrom = "Player2"
	dto.SabotagedCard = cards.NewSword("S1", 7)

	gs := gamestatus.NewGameStatus(dto)

	assert.Empty(t, gs.SabotagedFromYouCard)
}

// ──────────────────────────────────────────────────────────────────────────────
// SpyNotification tests
// ──────────────────────────────────────────────────────────────────────────────

func TestNewGameStatus_SpyNotification_DeckSpySeenByBystander(t *testing.T) {
	dto := minimalDTO("Player2") // Player2 is a bystander
	dto.LastAction = types.LastActionSpy
	dto.SpyTarget = types.SpyTargetDeck
	dto.CurrentPlayerName = "Player1"

	gs := gamestatus.NewGameStatus(dto)

	assert.Equal(t, "Player1 spied on the deck", gs.SpyNotification)
}

func TestNewGameStatus_SpyNotification_PlayerSpySeenByBystander(t *testing.T) {
	dto := minimalDTO("Player3") // Player3 is a bystander
	dto.LastAction = types.LastActionSpy
	dto.SpyTarget = types.SpyTargetPlayer
	dto.SpyTargetPlayer = "Player2"
	dto.CurrentPlayerName = "Player1"

	gs := gamestatus.NewGameStatus(dto)

	assert.Equal(t, "Player1 spied on Player2's hand", gs.SpyNotification)
}

func TestNewGameStatus_SpyNotification_EmptyForSpy(t *testing.T) {
	dto := minimalDTO("Player1") // Player1 is the viewer (spy)
	dto.LastAction = types.LastActionSpy
	dto.SpyTarget = types.SpyTargetDeck
	dto.CurrentPlayerName = "Player1"

	gs := gamestatus.NewGameStatus(dto)

	assert.Empty(t, gs.SpyNotification)
}

func TestNewGameStatus_SpyNotification_EmptyWhenLastActionDiffers(t *testing.T) {
	dto := minimalDTO("Player2")
	dto.LastAction = types.LastActionAttack
	dto.SpyTarget = types.SpyTargetDeck
	dto.CurrentPlayerName = "Player1"

	gs := gamestatus.NewGameStatus(dto)

	assert.Empty(t, gs.SpyNotification)
}

func TestNewGameStatus_SpyNotification_EmptyWhenSpyTargetEmpty(t *testing.T) {
	dto := minimalDTO("Player2")
	dto.LastAction = types.LastActionSpy
	dto.SpyTarget = "" // no target set
	dto.CurrentPlayerName = "Player1"

	gs := gamestatus.NewGameStatus(dto)

	assert.Empty(t, gs.SpyNotification)
}

// ──────────────────────────────────────────────────────────────────────────────
// Attack animation fields tests
// ──────────────────────────────────────────────────────────────────────────────

func TestNewGameStatus_LastMovedWarriorID_SetOnMoveWarrior(t *testing.T) {
	dto := minimalDTO("Player1")
	dto.LastAction = types.LastActionMoveWarrior
	dto.LastMovedWarriorID = "K1"

	gs := gamestatus.NewGameStatus(dto)

	assert.Equal(t, "K1", gs.LastMovedWarriorID)
}

func TestNewGameStatus_LastMovedWarriorID_EmptyOnOtherAction(t *testing.T) {
	dto := minimalDTO("Player1")
	dto.LastAction = types.LastActionAttack
	dto.LastMovedWarriorID = "K1"

	gs := gamestatus.NewGameStatus(dto)

	assert.Empty(t, gs.LastMovedWarriorID)
}

func TestNewGameStatus_AttackAnimationFields_SetOnAttack(t *testing.T) {
	dto := minimalDTO("Player1")
	dto.LastAction = types.LastActionAttack
	dto.LastAttackWeaponID = "S1"
	dto.LastAttackTargetID = "EK1"
	dto.LastAttackTargetPlayer = "Player2"

	gs := gamestatus.NewGameStatus(dto)

	assert.Equal(t, "S1", gs.LastAttackWeaponID)
	assert.Equal(t, "EK1", gs.LastAttackTargetID)
	assert.Equal(t, "Player2", gs.LastAttackTargetPlayer)
}

func TestNewGameStatus_AttackAnimationFields_SetOnHarpoon(t *testing.T) {
	dto := minimalDTO("Player1")
	dto.LastAction = types.LastActionHarpoon
	dto.LastAttackWeaponID = "H1"
	dto.LastAttackTargetID = "ED1"
	dto.LastAttackTargetPlayer = "Player2"

	gs := gamestatus.NewGameStatus(dto)

	assert.Equal(t, "H1", gs.LastAttackWeaponID)
	assert.Equal(t, "ED1", gs.LastAttackTargetID)
	assert.Equal(t, "Player2", gs.LastAttackTargetPlayer)
}

func TestNewGameStatus_AttackAnimationFields_EmptyOnOtherAction(t *testing.T) {
	dto := minimalDTO("Player1")
	dto.LastAction = types.LastActionBuy
	dto.LastAttackWeaponID = "S1"
	dto.LastAttackTargetID = "EK1"
	dto.LastAttackTargetPlayer = "Player2"

	gs := gamestatus.NewGameStatus(dto)

	assert.Empty(t, gs.LastAttackWeaponID)
	assert.Empty(t, gs.LastAttackTargetID)
	assert.Empty(t, gs.LastAttackTargetPlayer)
}

func TestNewGameStatus_BloodRainTargetPlayer_SetOnBloodRain(t *testing.T) {
	dto := minimalDTO("Player1")
	dto.LastAction = types.LastActionBloodRain
	dto.LastAttackTargetPlayer = "Player2"

	gs := gamestatus.NewGameStatus(dto)

	assert.Equal(t, "Player2", gs.LastAttackTargetPlayer)
}

func TestNewGameStatus_BloodRainTargetPlayer_EmptyOnOtherAction(t *testing.T) {
	dto := minimalDTO("Player1")
	dto.LastAction = types.LastActionAttack
	// LastAttackWeaponID is empty, so attack branch won't set it either
	dto.LastAttackTargetPlayer = "Player2"

	gs := gamestatus.NewGameStatus(dto)

	assert.Empty(t, gs.LastAttackTargetPlayer)
}

// ──────────────────────────────────────────────────────────────────────────────
// Game over tests
// ──────────────────────────────────────────────────────────────────────────────

func TestNewGameStatus_GameOver_WinnerSeesWinMessage(t *testing.T) {
	dto := minimalDTO("Player1")
	dto.IsGameOver = true
	dto.Winner = "Player1"
	dto.IsPlayerWinner = true

	gs := gamestatus.NewGameStatus(dto)

	assert.Equal(t, "Game over! The winner is Player1", gs.GameOverMgs)
	assert.True(t, gs.IsWinner)
}

func TestNewGameStatus_GameOver_LoserSeesWinMessage(t *testing.T) {
	dto := minimalDTO("Player2")
	dto.IsGameOver = true
	dto.Winner = "Player1"
	dto.IsPlayerWinner = false

	gs := gamestatus.NewGameStatus(dto)

	assert.Equal(t, "Game over! The winner is Player1", gs.GameOverMgs)
	assert.False(t, gs.IsWinner)
}

func TestNewGameStatus_GameOver_EmptyWhenNotOver(t *testing.T) {
	dto := minimalDTO("Player1")
	dto.IsGameOver = false

	gs := gamestatus.NewGameStatus(dto)

	assert.Empty(t, gs.GameOverMgs)
	assert.False(t, gs.IsWinner)
}

// ──────────────────────────────────────────────────────────────────────────────
// NewCards and ModalCards tests
// ──────────────────────────────────────────────────────────────────────────────

func TestNewGameStatus_NewCards_IDsCollected(t *testing.T) {
	dto := minimalDTO("Player1")
	dto.NewCards = []cards.Card{cards.NewKnight("K1"), cards.NewSword("S1", 7)}

	gs := gamestatus.NewGameStatus(dto)

	assert.Equal(t, []string{"K1", "S1"}, gs.NewCards)
}

func TestNewGameStatus_NewCards_EmptyWhenNone(t *testing.T) {
	dto := minimalDTO("Player1")

	gs := gamestatus.NewGameStatus(dto)

	assert.Empty(t, gs.NewCards)
}

func TestNewGameStatus_ModalCards_Converted(t *testing.T) {
	dto := minimalDTO("Player1")
	dto.ModalCards = []cards.Card{cards.NewKnight("K1")}

	gs := gamestatus.NewGameStatus(dto)

	assert.Len(t, gs.ModalCards, 1)
	assert.Equal(t, "K1", gs.ModalCards[0].CardID)
	assert.Equal(t, gamestatus.CardTypeKnight, gs.ModalCards[0].CardType)
}

func TestNewGameStatus_ModalCards_NilWhenNone(t *testing.T) {
	dto := minimalDTO("Player1")

	gs := gamestatus.NewGameStatus(dto)

	assert.Nil(t, gs.ModalCards)
}

// ──────────────────────────────────────────────────────────────────────────────
// History tests
// ──────────────────────────────────────────────────────────────────────────────

func TestNewGameStatus_History_LinesConverted(t *testing.T) {
	dto := minimalDTO("Player1")
	dto.History = []types.HistoryLine{
		{Msg: "Player1 attacked Player2", Category: types.CategoryAction},
		{Msg: "Turn ended", Category: types.CategoryEndTurn},
	}

	gs := gamestatus.NewGameStatus(dto)

	assert.Len(t, gs.History, 2)
	assert.Equal(t, "Player1 attacked Player2", gs.History[0].Msg)
	assert.Equal(t, "#33C1FF", gs.History[0].Color)
	assert.Equal(t, "Turn ended", gs.History[1].Msg)
	assert.Equal(t, "#F39C12", gs.History[1].Color)
}

func TestNewGameStatus_History_EmptyWhenNone(t *testing.T) {
	dto := minimalDTO("Player1")

	gs := gamestatus.NewGameStatus(dto)

	assert.Empty(t, gs.History)
}

// ──────────────────────────────────────────────────────────────────────────────
// CurrentPlayerField tests
// ──────────────────────────────────────────────────────────────────────────────

func TestNewGameStatus_CurrentPlayerField_PopulatedFromViewerWarriors(t *testing.T) {
	dto := minimalDTO("Player1")
	dto.Viewer.Field.Warriors = []cards.Warrior{cards.NewKnight("K1")}

	gs := gamestatus.NewGameStatus(dto)

	assert.Len(t, gs.CurrentPlayerField, 1)
	assert.Equal(t, "K1", gs.CurrentPlayerField[0].CardID)
	assert.Equal(t, gamestatus.CardTypeKnight, gs.CurrentPlayerField[0].CardType)
}

func TestNewGameStatus_CurrentPlayerField_EmptyWhenNoWarriors(t *testing.T) {
	dto := minimalDTO("Player1")

	gs := gamestatus.NewGameStatus(dto)

	assert.Empty(t, gs.CurrentPlayerField)
}
