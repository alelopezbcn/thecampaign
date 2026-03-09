package gameactions

import (
	"fmt"
	"math/rand"

	"github.com/alelopezbcn/thecampaign/internal/domain/board"
	"github.com/alelopezbcn/thecampaign/internal/domain/cards"
	"github.com/alelopezbcn/thecampaign/internal/domain/gamestatus"
	"github.com/alelopezbcn/thecampaign/internal/domain/types"
)

// harpoonGame declares the minimum Game surface needed by harpoonAction
type harpoonGame interface {
	GamePlayers
	GameTurn
	GameCards
	GameHistory
	GameStatusProvider
}

type harpoonAction struct {
	playerName       string
	targetPlayerName string
	targetID         string
	weaponID         string

	targetPlayer board.Player
	dragon       cards.Dragon
	harpoon      cards.Harpoon
}

func NewHarpoonAction(playerName, targetPlayerName, targetID, weaponID string) *harpoonAction {
	return &harpoonAction{
		playerName:       playerName,
		targetPlayerName: targetPlayerName,
		weaponID:         weaponID,
		targetID:         targetID,
	}
}

func (a *harpoonAction) PlayerName() string { return a.playerName }

func (a *harpoonAction) Validate(g Game) error {
	if g.CurrentAction() != types.PhaseTypeAttack {
		return fmt.Errorf("cannot use harpoon in the %s phase",
			g.CurrentAction())
	}

	targetPlayer, err := g.GetTargetPlayer(a.playerName, a.targetPlayerName)
	if err != nil {
		return err
	}
	a.targetPlayer = targetPlayer

	targetCard, ok := targetPlayer.GetCardFromField(a.targetID)
	if !ok {
		return fmt.Errorf("dragon card not in enemy field: %s", a.targetID)
	}

	a.dragon, ok = targetCard.(cards.Dragon)
	if !ok {
		return fmt.Errorf("the target card is not a dragon")
	}

	p := g.CurrentPlayer()
	// Look up the specific harpoon card by the ID provided by the client.
	// This avoids the mismatch possible in FFA5 where both HA1 and HA2 exist.
	harpoonCard, ok := p.GetCardFromHand(a.weaponID)
	if !ok {
		return fmt.Errorf("player does not have a harpoon to use")
	}
	harpoon, ok := harpoonCard.(cards.Harpoon)
	if !ok {
		return fmt.Errorf("card %s is not a harpoon", a.weaponID)
	}

	a.harpoon = harpoon

	return nil
}

func (a *harpoonAction) Execute(g Game) (*Result, func() gamestatus.GameStatus, error) {
	return a.execute(g)
}

func (a *harpoonAction) execute(g harpoonGame) (*Result, func() gamestatus.GameStatus, error) {
	p := g.CurrentPlayer()
	handler := g.EventHandler()
	bountyCards := handler.OnKillBountyCards()
	healAmount := handler.OnKillHealAmount()

	preKillHP := a.snapshotPreKillHP(bountyCards)

	if err := a.harpoon.Attack(a.dragon); err != nil {
		return &Result{}, nil, fmt.Errorf("harpoon action failed: %w", err)
	}

	a.applyBloodlust(healAmount, p)

	if _, err := p.RemoveFromHand(a.harpoon.GetID()); err != nil {
		return &Result{}, nil, fmt.Errorf("removing harpoon from hand failed: %w", err)
	}
	g.OnCardMovedToPile(a.harpoon)
	g.AddHistory(fmt.Sprintf("%s used harpoon on %s", a.playerName, a.dragon.String()), types.CategoryAction)

	newCards, earner, drawn := a.applyChampionsBounty(g, p, bountyCards, preKillHP)

	result := &Result{
		Action: types.LastActionHarpoon,
		Attack: &AttackDetails{
			WeaponID:              a.weaponID,
			TargetID:              a.targetID,
			TargetPlayer:          a.targetPlayerName,
			ChampionsBountyEarner: earner,
			ChampionsBountyCards:  drawn,
		},
	}
	return result, func() gamestatus.GameStatus { return g.Status(p, newCards...) }, nil
}

func (a *harpoonAction) NextPhase() types.PhaseType {
	return types.PhaseTypeSpySteal
}

// snapshotPreKillHP returns the target player's total field HP before the attack,
// used for Champion's Bounty eligibility. Returns 0 when the event is not active.
func (a *harpoonAction) snapshotPreKillHP(bountyCards int) int {
	if bountyCards == 0 {
		return 0
	}
	return totalFieldHP(a.targetPlayer)
}

// applyBloodlust heals a random field warrior if the dragon was killed.
func (a *harpoonAction) applyBloodlust(healAmount int, p board.Player) {
	if healAmount <= 0 || a.dragon.Health() != 0 {
		return
	}
	warriors := p.Field().Warriors()
	if len(warriors) == 0 {
		return
	}
	warriors[rand.Intn(len(warriors))].HealBy(healAmount)
}

// applyChampionsBounty draws bonus cards when the dragon was killed and its player
// had the highest total field HP. Returns the drawn cards, earner name, and count.
func (a *harpoonAction) applyChampionsBounty(g harpoonGame, p board.Player, bountyCards, preKillHP int) ([]cards.Card, string, int) {
	if bountyCards == 0 || a.dragon.Health() != 0 {
		return nil, "", 0
	}
	if !isTopEnemy(preKillHP, a.targetPlayerName, g.Enemies(g.PlayerIndex(a.playerName))) {
		return nil, "", 0
	}
	drawn, err := g.DrawCards(p, bountyCards)
	if err != nil {
		return nil, "", 0
	}
	p.TakeCards(drawn...)
	name := p.Name()
	g.AddHistory(fmt.Sprintf("%s earned Champion's Bounty — drew %d card(s)", name, len(drawn)), types.CategoryInfo)
	return drawn, name, len(drawn)
}
