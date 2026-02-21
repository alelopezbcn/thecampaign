package game

import (
	"fmt"

	"github.com/alelopezbcn/thecampaign/internal/domain/gamestatus"
	"github.com/alelopezbcn/thecampaign/internal/domain/ports"
	"github.com/alelopezbcn/thecampaign/internal/domain/types"
)

type SpecialPowerAction struct {
	playerName string
	userID     string
	targetID   string
	weaponID   string

	usedBy       ports.Warrior
	usedOn       ports.Warrior
	specialPower ports.SpecialPower
}

func NewSpecialPowerAction(playerName, userID, targetID, weaponID string) *SpecialPowerAction {
	return &SpecialPowerAction{
		playerName: playerName,
		userID:     userID,
		targetID:   targetID,
		weaponID:   weaponID,
	}
}

func (a *SpecialPowerAction) PlayerName() string { return a.playerName }

func (a *SpecialPowerAction) Validate(g *Game) error {
	if g.currentAction != types.PhaseTypeAttack {
		return fmt.Errorf("cannot use special power in the %s phase",
			g.currentAction)
	}

	p := g.CurrentPlayer()

	userCard, ok := p.GetCardFromField(a.userID)
	if !ok {
		return fmt.Errorf("warrior card not in field: %s", a.userID)
	}

	// Determine user warrior type for validation
	a.usedBy, ok = userCard.(ports.Warrior)
	if !ok {
		return fmt.Errorf("the attacking card is not a warrior")
	}
	userType := a.usedBy.Type()

	var targetCard ports.Card
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
				break
			}
		}
	}
	if !ok {
		return fmt.Errorf("target card not valid: %s", a.targetID)
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

	a.specialPower, ok = weaponCard.(ports.SpecialPower)
	if !ok {
		return fmt.Errorf("the card is not a special power")
	}

	a.usedOn, ok = targetCard.(ports.Warrior)
	if !ok {
		return fmt.Errorf("the target card is not a warrior")
	}

	return nil
}

func (a *SpecialPowerAction) Execute(g *Game) (*GameActionResult, func() gamestatus.GameStatus, error) {
	p := g.CurrentPlayer()

	if err := p.UseSpecialPower(a.usedBy, a.usedOn, a.specialPower); err != nil {
		result := &GameActionResult{}
		return result, nil, fmt.Errorf("special power action failed: %w", err)
	}

	g.AddHistory(fmt.Sprintf("%s used special power on %s",
		a.playerName, a.usedOn.String()), types.CategoryAction)

	result := &GameActionResult{
		Action: types.LastActionSpecialPower,
	}
	statusFn := func() gamestatus.GameStatus {
		return g.gameStatusProvider.Get(p, g)
	}

	return result, statusFn, nil
}

func (a *SpecialPowerAction) NextPhase() types.PhaseType {
	return types.PhaseTypeSpySteal
}
