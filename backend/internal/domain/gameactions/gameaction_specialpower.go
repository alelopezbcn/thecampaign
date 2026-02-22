package gameactions

import (
	"fmt"

	"github.com/alelopezbcn/thecampaign/internal/domain/cards"
	"github.com/alelopezbcn/thecampaign/internal/domain/gamestatus"
	"github.com/alelopezbcn/thecampaign/internal/domain/types"
)

type specialPowerAction struct {
	playerName string
	userID     string
	targetID   string
	weaponID   string

	usedBy       cards.Warrior
	usedOn       cards.Warrior
	specialPower cards.SpecialPower
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
	p := g.CurrentPlayer()

	if err := a.specialPower.Use(a.usedBy, a.usedOn); err != nil {
		result := &Result{}
		return result, nil, fmt.Errorf("special power action failed: %w", err)
	}

	p.RemoveFromHand(a.specialPower.GetID())

	g.AddHistory(fmt.Sprintf("%s used special power on %s",
		a.playerName, a.usedOn.String()), types.CategoryAction)

	result := &Result{
		Action: types.LastActionSpecialPower,
	}
	statusFn := func() gamestatus.GameStatus {
		return g.Status(p)
	}

	return result, statusFn, nil
}

func (a *specialPowerAction) NextPhase() types.PhaseType {
	return types.PhaseTypeSpySteal
}
