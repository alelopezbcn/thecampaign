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
	GameCards
	GameHistory
	GameStatusProvider
}

type bloodRainAction struct {
	playerName       string
	targetPlayerName string
	weaponID         string

	currentPlayer board.Player
	targetPlayer  board.Player
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
	a.targetPlayer = targetPlayer

	targets := targetPlayer.Field().Warriors()
	a.targets = targets

	p := g.CurrentPlayer()
	a.currentPlayer = p

	// Look up the specific blood rain card by the ID provided by the client.
	// This avoids the mismatch that occurs when HasCardTypeInHand returns a
	// different card than the one the client intended (possible in FFA5 where
	// both BR1 and BR2 are in the deck and a player can hold either).
	c, ok := p.GetCardFromHand(a.weaponID)
	if !ok {
		return fmt.Errorf("player does not have a blood rain card to use")
	}
	bloodRain, ok := c.(cards.BloodRain)
	if !ok {
		return fmt.Errorf("card %s is not a blood rain card", a.weaponID)
	}

	a.bloodRain = bloodRain

	return nil
}

func (a *bloodRainAction) Execute(g Game) (*Result, func() gamestatus.GameStatus, error) {
	return a.execute(g)
}

func (a *bloodRainAction) execute(g bloodRainGame) (*Result, func() gamestatus.GameStatus, error) {
	handler := g.EventHandler()
	bountyCards := handler.OnKillBountyCards()
	healAmount := handler.OnKillHealAmount()

	preKillHP := a.snapshotPreKillHP(bountyCards)

	if err := a.bloodRain.Attack(a.targets); err != nil {
		return &Result{}, nil, fmt.Errorf("blood rain action failed: %w", err)
	}

	kills := a.countKills(healAmount, bountyCards)
	a.applyBloodlust(healAmount, kills)

	if _, err := a.currentPlayer.RemoveFromHand(a.bloodRain.GetID()); err != nil {
		return &Result{}, nil, fmt.Errorf("removing blood rain from hand failed: %w", err)
	}
	g.OnCardMovedToPile(a.bloodRain)
	g.AddHistory(fmt.Sprintf("%s used blood rain on %s", a.playerName, a.targetPlayerName), types.CategoryAction)

	newCards, earner, drawn := a.applyChampionsBounty(g, bountyCards, preKillHP, kills)

	result := &Result{
		Action: types.LastActionBloodRain,
		Attack: &AttackDetails{
			WeaponID:              a.weaponID,
			TargetPlayer:          a.targetPlayerName,
			ChampionsBountyEarner: earner,
			ChampionsBountyCards:  drawn,
		},
	}
	return result, func() gamestatus.GameStatus { return g.Status(a.currentPlayer, newCards...) }, nil
}

func (a *bloodRainAction) NextPhase() types.PhaseType {
	return types.PhaseTypeSpySteal
}

// snapshotPreKillHP returns the target player's total field HP before the attack,
// used for Champion's Bounty eligibility. Returns 0 when the event is not active.
func (a *bloodRainAction) snapshotPreKillHP(bountyCards int) int {
	if bountyCards == 0 {
		return 0
	}
	return totalFieldHP(a.targetPlayer)
}

// countKills counts target warriors that were killed by the attack.
// Skips the count entirely when neither Bloodlust nor Champion's Bounty is active.
func (a *bloodRainAction) countKills(healAmount, bountyCards int) int {
	if healAmount == 0 && bountyCards == 0 {
		return 0
	}
	kills := 0
	for _, t := range a.targets {
		if t.Health() == 0 {
			kills++
		}
	}
	return kills
}

// applyBloodlust heals all of the player's field warriors for each target killed.
func (a *bloodRainAction) applyBloodlust(healAmount, kills int) {
	if healAmount > 0 && kills > 0 {
		for _, w := range a.currentPlayer.Field().Warriors() {
			w.HealBy(healAmount * kills)
		}
	}
}

// applyChampionsBounty draws bonus cards when any target was killed and the target
// player had the highest total field HP. Returns the drawn cards, earner name, and count.
func (a *bloodRainAction) applyChampionsBounty(g bloodRainGame, bountyCards, preKillHP, kills int) ([]cards.Card, string, int) {
	if bountyCards == 0 || kills == 0 {
		return nil, "", 0
	}
	if !isTopEnemy(preKillHP, a.targetPlayerName, g.Enemies(g.PlayerIndex(a.playerName))) {
		return nil, "", 0
	}
	drawn, err := g.DrawCards(a.currentPlayer, bountyCards)
	if err != nil {
		return nil, "", 0
	}
	a.currentPlayer.TakeCards(drawn...)
	name := a.currentPlayer.Name()
	g.AddHistory(fmt.Sprintf("%s earned Champion's Bounty — drew %d card(s)", name, len(drawn)), types.CategoryInfo)
	return drawn, name, len(drawn)
}
