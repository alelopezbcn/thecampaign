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
