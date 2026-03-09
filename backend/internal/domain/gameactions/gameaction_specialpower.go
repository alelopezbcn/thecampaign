package gameactions

import (
	"fmt"

	"github.com/alelopezbcn/thecampaign/internal/domain/board"
	"github.com/alelopezbcn/thecampaign/internal/domain/cards"
	"github.com/alelopezbcn/thecampaign/internal/domain/gamestatus"
	"github.com/alelopezbcn/thecampaign/internal/domain/types"
)

// specialPowerGame declares the minimum Game surface needed by specialPowerAction
type specialPowerGame interface {
	GamePlayers
	GameTurn
	GameCards
	GameHistory
	GameStatusProvider
}

type specialPowerAction struct {
	playerName string
	userID     string
	targetID   string
	weaponID   string

	usedBy       cards.Warrior
	usedOn       cards.Warrior
	specialPower cards.SpecialPower
	targetPlayer board.Player
}

func NewSpecialPowerAction(playerName, userID, targetID, weaponID string) *specialPowerAction {
	return &specialPowerAction{
		playerName: playerName,
		userID:     userID,
		targetID:   targetID,
		weaponID:   weaponID,
	}
}

func (a *specialPowerAction) PlayerName() string { return a.playerName }

func (a *specialPowerAction) Validate(g Game) error {
	if g.CurrentAction() != types.PhaseTypeAttack {
		return fmt.Errorf("cannot use special power in the %s phase",
			g.CurrentAction())
	}

	p := g.CurrentPlayer()

	userCard, ok := p.GetCardFromField(a.userID)
	if !ok {
		return fmt.Errorf("warrior card not in field: %s", a.userID)
	}

	// Determine user warrior type for validation
	a.usedBy, ok = userCard.(cards.Warrior)
	if !ok {
		return fmt.Errorf("the attacking card is not a warrior")
	}
	userType := a.usedBy.Type()

	var targetCard cards.Card
	targetIsAllyOrSelf := false

	// Search own field
	targetCard, ok = p.GetCardFromField(a.targetID)
	if ok {
		targetIsAllyOrSelf = true
	}
	if !ok {
		// Search ally fields (2v2)
		for _, ally := range g.Allies(g.PlayerIndex(a.playerName)) {
			targetCard, ok = ally.GetCardFromField(a.targetID)
			if ok {
				targetIsAllyOrSelf = true
				break
			}
		}
	}
	if !ok {
		// Search enemy fields
		for _, enemy := range g.Enemies(g.PlayerIndex(a.playerName)) {
			targetCard, ok = enemy.GetCardFromField(a.targetID)
			if ok {
				a.targetPlayer = enemy
				break
			}
		}
	}
	if !ok {
		return fmt.Errorf("target card not valid: %s", a.targetID)
	}

	// Only Archer, Knight and Mage can use special powers
	if userType != types.ArcherWarriorType &&
		userType != types.KnightWarriorType &&
		userType != types.MageWarriorType {
		return fmt.Errorf("warrior type %s cannot use special powers", userType)
	}

	// Validate target side based on warrior type
	if userType == types.ArcherWarriorType && targetIsAllyOrSelf {
		return fmt.Errorf("archer instant kill can only target enemies")
	}
	if (userType == types.KnightWarriorType || userType == types.MageWarriorType) && !targetIsAllyOrSelf {
		return fmt.Errorf("knight/mage special power can only target allies")
	}

	weaponCard, ok := p.GetCardFromHand(a.weaponID)
	if !ok {
		return fmt.Errorf("weapon card not in hand: %s", a.weaponID)
	}

	a.specialPower, ok = weaponCard.(cards.SpecialPower)
	if !ok {
		return fmt.Errorf("the card is not a special power")
	}

	a.usedOn, ok = targetCard.(cards.Warrior)
	if !ok {
		return fmt.Errorf("the target card is not a warrior")
	}

	return nil
}

func (a *specialPowerAction) Execute(g Game) (*Result, func() gamestatus.GameStatus, error) {
	return a.execute(g)
}

func (a *specialPowerAction) execute(g specialPowerGame) (*Result, func() gamestatus.GameStatus, error) {
	p := g.CurrentPlayer()
	handler := g.EventHandler()
	bountyCards := handler.OnKillBountyCards()
	healAmount := handler.OnKillHealAmount()

	preKillHP := a.snapshotPreKillHP(bountyCards)

	// Snapshot pre-attack HP for damage tracking.
	targetPreHP := a.usedOn.Health()

	if err := a.specialPower.Use(a.usedBy, a.usedOn); err != nil {
		return &Result{}, nil, fmt.Errorf("special power action failed: %w", err)
	}

	// Post-attack state (single Health() call used for kill detection and damage calculation).
	postHP := a.usedOn.Health()
	killed := postHP == 0
	if killed {
		a.usedBy.AddKill()
	}
	a.applyBloodlust(healAmount, killed)

	if _, err := p.RemoveFromHand(a.specialPower.GetID()); err != nil {
		return &Result{}, nil, fmt.Errorf("removing special power from hand failed: %w", err)
	}

	g.AddHistory(fmt.Sprintf("%s used special power on %s",
		a.playerName, a.usedOn.String()), types.CategoryAction)

	newCards, earner, drawn := a.applyChampionsBounty(g, p, bountyCards, preKillHP, killed)

	killsGranted := 0
	if killed {
		killsGranted = 1
	}
	result := &Result{
		Action: types.LastActionSpecialPower,
		Attack: &AttackDetails{
			TargetID:              a.targetID,
			ChampionsBountyEarner: earner,
			ChampionsBountyCards:  drawn,
			KillsGranted:          killsGranted,
			DamageDealt:           targetPreHP - postHP,
		},
	}
	return result, func() gamestatus.GameStatus { return g.Status(p, newCards...) }, nil
}

func (a *specialPowerAction) NextPhase() types.PhaseType {
	return types.PhaseTypeSpySteal
}

// snapshotPreKillHP returns the target player's total field HP before the attack,
// used for Champion's Bounty eligibility. Returns 0 when the event is not active
// or when the target is not an enemy (ally/self targets cannot earn bounty).
func (a *specialPowerAction) snapshotPreKillHP(bountyCards int) int {
	if bountyCards == 0 || a.targetPlayer == nil {
		return 0
	}
	return totalFieldHP(a.targetPlayer)
}

// applyBloodlust heals the attacking warrior when the target was instantly killed.
func (a *specialPowerAction) applyBloodlust(healAmount int, killed bool) {
	if healAmount > 0 && killed {
		a.usedBy.HealBy(healAmount)
	}
}

// applyChampionsBounty draws bonus cards when the target was killed and the target
// player had the highest total field HP. Returns the drawn cards, earner name, and count.
func (a *specialPowerAction) applyChampionsBounty(g specialPowerGame, p board.Player, bountyCards, preKillHP int, killed bool) ([]cards.Card, string, int) {
	if bountyCards == 0 || !killed || a.targetPlayer == nil {
		return nil, "", 0
	}
	targetPlayerName := a.targetPlayer.Name()
	if !isTopEnemy(preKillHP, targetPlayerName, g.Enemies(g.PlayerIndex(a.playerName))) {
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
