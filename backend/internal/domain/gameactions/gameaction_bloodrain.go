package gameactions

import (
	"fmt"

	"github.com/alelopezbcn/thecampaign/internal/domain/board"
	"github.com/alelopezbcn/thecampaign/internal/domain/cards"
	"github.com/alelopezbcn/thecampaign/internal/domain/gamestatus"
	"github.com/alelopezbcn/thecampaign/internal/domain/types"
)

// bloodRainGame declares the minimum Game surface needed by bloodRainAction
type bloodRainGame interface {
	GamePlayers
	GameTurn
	GameHistory
	GameStatusProvider
}

type bloodRainAction struct {
	playerName       string
	targetPlayerName string
	weaponID         string

	currentPlayer board.Player
	targets       []cards.Warrior
	bloodRain     cards.BloodRain
}

func NewBloodRainAction(playerName, targetPlayerName, weaponID string) *bloodRainAction {
	return &bloodRainAction{
		playerName:       playerName,
		targetPlayerName: targetPlayerName,
		weaponID:         weaponID,
	}
}

func (a *bloodRainAction) PlayerName() string { return a.playerName }

func (a *bloodRainAction) Validate(g Game) error {
	if g.CurrentAction() != types.PhaseTypeAttack {
		return fmt.Errorf("cannot use blood rain in the %s phase",
			g.CurrentAction())
	}

	targetPlayer, err := g.GetTargetPlayer(a.playerName, a.targetPlayerName)
	if err != nil {
		return err
	}

	targets := targetPlayer.Field().Warriors()
	a.targets = targets

	p := g.CurrentPlayer()
	a.currentPlayer = p

	bloodRain, ok := board.HasCardTypeInHand[cards.BloodRain](p)
	if !ok {
		return fmt.Errorf("player does not have a blood rain card to use")
	}

	a.bloodRain = bloodRain

	return nil
}

func (a *bloodRainAction) Execute(g Game) (*Result, func() gamestatus.GameStatus, error) {
	return a.execute(g)
}

func (a *bloodRainAction) execute(g bloodRainGame) (*Result, func() gamestatus.GameStatus, error) {
	if err := a.bloodRain.Attack(a.targets); err != nil {
		result := &Result{}
		return result, nil, fmt.Errorf("blood rain action failed: %w", err)
	}

	if _, err := a.currentPlayer.RemoveFromHand(a.weaponID); err != nil {
		result := &Result{}
		return result, nil, fmt.Errorf("removing blood rain from hand failed: %w", err)
	}

	g.AddHistory(fmt.Sprintf("%s used blood rain on %s",
		a.playerName, a.targetPlayerName), types.CategoryAction)

	result := &Result{
		Action:             types.LastActionBloodRain,
		AttackWeaponID:     a.weaponID,
		AttackTargetPlayer: a.targetPlayerName,
	}
	statusFn := func() gamestatus.GameStatus {
		return g.Status(a.currentPlayer)
	}

	return result, statusFn, nil
}

func (a *bloodRainAction) NextPhase() types.PhaseType {
	return types.PhaseTypeSpySteal
}
