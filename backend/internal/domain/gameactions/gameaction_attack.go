package gameactions

import (
	"fmt"
	"math/rand"

	"github.com/alelopezbcn/thecampaign/internal/domain/board"
	"github.com/alelopezbcn/thecampaign/internal/domain/cards"
	"github.com/alelopezbcn/thecampaign/internal/domain/gamestatus"
	"github.com/alelopezbcn/thecampaign/internal/domain/types"
)

// attackGame declares the minimum Game surface needed by attackAction
type attackGame interface {
	GamePlayers
	GameTurn
	GameCards
	GameHistory
	GameStatusProvider
}

// curseModifiedWeapon wraps a Weapon and overrides DamageAmount with the Curse-adjusted value.
type curseModifiedWeapon struct {
	cards.Weapon
	effectiveDamage int
}

func (w *curseModifiedWeapon) DamageAmount() int { return w.effectiveDamage }

// applyWeaponModifier returns the weapon unchanged when mod is 0,
// otherwise returns a wrapper with the Curse-adjusted damage (minimum 0).
func applyWeaponModifier(weapon cards.Weapon, mod int) cards.Weapon {
	if mod == 0 {
		return weapon
	}
	effective := weapon.DamageAmount() + mod
	if effective < 0 {
		effective = 0
	}
	return &curseModifiedWeapon{Weapon: weapon, effectiveDamage: effective}
}

type attackAction struct {
	playerName       string
	warriorID        string
	targetPlayerName string
	targetID         string
	weaponID         string

	currentPlayer board.Player
	targetPlayer  board.Player
	target        cards.Attackable
	weapon        cards.Weapon
	attacker      cards.Warrior
}

func NewAttackAction(playerName, warriorID, targetPlayerName, targetID, weaponID string) *attackAction {
	return &attackAction{
		playerName:       playerName,
		warriorID:        warriorID,
		targetPlayerName: targetPlayerName,
		targetID:         targetID,
		weaponID:         weaponID,
	}
}

func (a *attackAction) PlayerName() string { return a.playerName }

func (a *attackAction) Validate(g Game) error {
	if g.CurrentAction() != types.PhaseTypeAttack {
		return fmt.Errorf("cannot attack in the %s phase", g.CurrentAction())
	}

	targetPlayer, err := g.GetTargetPlayer(a.playerName, a.targetPlayerName)
	if err != nil {
		return err
	}

	targetCard, ok := targetPlayer.GetCardFromField(a.targetID)
	if !ok {
		return fmt.Errorf("target card not in enemy field: %s", a.targetID)
	}

	p := g.CurrentPlayer()
	a.currentPlayer = p
	a.targetPlayer = targetPlayer
	weaponCard, ok := p.GetCardFromHand(a.weaponID)
	if !ok {
		return fmt.Errorf("weapon card not in hand: %s", a.weaponID)
	}

	a.target, ok = targetCard.(cards.Attackable)
	if !ok {
		return fmt.Errorf("the target card cannot be attacked")
	}

	a.weapon, ok = weaponCard.(cards.Weapon)
	if !ok {
		return fmt.Errorf("the card is not a weapon")
	}

	if !a.weapon.CanBeUsedWith(a.currentPlayer.Field()) {
		return fmt.Errorf("%s weapon cannot be used", a.weapon.Type())
	}

	attackerCard, ok := p.GetCardFromField(a.warriorID)
	if !ok {
		return fmt.Errorf("warrior %s not found in your field", a.warriorID)
	}
	a.attacker, ok = attackerCard.(cards.Warrior)
	if !ok {
		return fmt.Errorf("card %s is not a warrior", a.warriorID)
	}
	wepType := a.weapon.Type()
	if !a.attacker.CanUseWeapon(wepType) {
		return fmt.Errorf("%s cannot use %s weapon", a.attacker.Type(), wepType)
	}

	return nil
}

func (a *attackAction) Execute(g Game) (*Result, func() gamestatus.GameStatus, error) {
	return a.execute(g)
}

func (a *attackAction) execute(g attackGame) (*Result, func() gamestatus.GameStatus, error) {
	handler := g.EventHandler()

	// Apply Curse event modifier to weapon damage (cached to avoid repeated Type() calls).
	weaponType := a.weapon.Type()
	effectiveWeapon := applyWeaponModifier(a.weapon, handler.WeaponDamageModifier(weaponType))

	// Apply attacker kill bonus (+1 DMG per kill earned by this warrior).
	effectiveWeapon = applyWeaponModifier(effectiveWeapon, a.attacker.Kills())

	// Check if the defender has an ambush in their field.
	if ambush, ok := board.GetFieldSlotCard[cards.Ambush](a.targetPlayer.Field()); ok {
		a.targetPlayer.Field().RemoveSlotCard(ambush)

		// Discard the ambush card via its observer (set when it was in defender's hand).
		ambush.GetCardMovedToPileObserver().OnCardMovedToPile(ambush)

		result := a.applyAmbushEffect(ambush.Effect(), effectiveWeapon, weaponType, g)
		statusFn := func() gamestatus.GameStatus {
			return g.Status(a.currentPlayer)
		}
		return result, statusFn, nil
	}

	// Champion's Bounty: snapshot target player's total field HP before the attack,
	// so we can compare against other enemies after the kill.
	bountyCards := handler.OnKillBountyCards()
	var targetPlayerPreKillHP int
	if bountyCards > 0 {
		targetPlayerPreKillHP = totalFieldHP(a.targetPlayer)
	}

	// Snapshot pre-attack HP for damage tracking.
	preAttackHP := a.target.Health()

	// Normal attack.
	err := a.target.BeAttacked(effectiveWeapon)
	if err != nil {
		result := &Result{}
		return result, nil, fmt.Errorf("attack action failed: %w", err)
	}

	// Post-attack state (single Health() call used for kill detection and damage calculation).
	postAttackHP := a.target.Health()
	killed := postAttackHP == 0
	if killed {
		a.attacker.AddKill()
	}

	// Bloodlust: if the target was killed, restore HP to a random field warrior.
	if healAmount := handler.OnKillHealAmount(); healAmount > 0 && killed {
		a.applyBloodlust(healAmount)
	}

	// Remove the used weapon before drawing bounty cards so the freed slot counts
	// toward CanTakeCards when Champion's Bounty tries to draw.
	a.currentPlayer.RemoveFromHand(a.weaponID)

	g.AddHistory(fmt.Sprintf("%s attacked %s with %s",
		a.currentPlayer.Name(), a.target.String(), a.weapon.String()),
		types.CategoryAction)

	// Champion's Bounty: if the target was killed and their player had the highest
	// total field HP (ties included), draw bonus cards.
	var bountyEarner string
	var bountyDrawn int
	var bountyCards2 []cards.Card
	if bountyCards > 0 && killed {
		if isTopEnemy(targetPlayerPreKillHP, a.targetPlayerName, g.Enemies(g.PlayerIndex(a.playerName))) {
			drawn, drawErr := g.DrawCards(a.currentPlayer, bountyCards)
			if drawErr == nil {
				a.currentPlayer.TakeCards(drawn...)
				bountyEarner = a.currentPlayer.Name()
				bountyDrawn = len(drawn)
				bountyCards2 = drawn
				g.AddHistory(fmt.Sprintf("%s earned Champion's Bounty — drew %d card(s)",
					a.currentPlayer.Name(), len(drawn)), types.CategoryInfo)
			}
		}
	}

	killsGranted := 0
	if killed {
		killsGranted = 1
	}
	result := &Result{
		Action: types.LastActionAttack,
		Attack: &AttackDetails{
			WeaponID:              a.weaponID,
			TargetID:              a.targetID,
			TargetPlayer:          a.targetPlayerName,
			ChampionsBountyEarner: bountyEarner,
			ChampionsBountyCards:  bountyDrawn,
			KillsGranted:          killsGranted,
			DamageDealt:           preAttackHP - postAttackHP,
		},
	}
	statusFn := func() gamestatus.GameStatus {
		return g.Status(a.currentPlayer, bountyCards2...)
	}

	return result, statusFn, nil
}

func (a *attackAction) NextPhase() types.PhaseType {
	return types.PhaseTypeSpySteal
}

// applyAmbushEffect applies the pre-determined ambush effect and returns the result.
// effectiveWeapon carries the Curse-adjusted damage; weaponType is cached to avoid extra Type() calls.
// totalFieldHP returns the sum of current HP of all warriors in a player's field.
func totalFieldHP(p board.Player) int {
	total := 0
	for _, w := range p.Field().Warriors() {
		total += w.Health()
	}
	return total
}

// isTopEnemy returns true if targetPreKillHP is >= the total field HP of every other enemy.
func isTopEnemy(targetPreKillHP int, targetPlayerName string, enemies []board.Player) bool {
	for _, e := range enemies {
		if e.Name() == targetPlayerName {
			continue
		}
		if totalFieldHP(e) > targetPreKillHP {
			return false
		}
	}
	return true
}

// applyBloodlust heals a random field warrior when a kill is scored.
func (a *attackAction) applyBloodlust(healAmount int) {
	warriors := a.currentPlayer.Field().Warriors()
	if len(warriors) == 0 {
		return
	}
	warriors[rand.Intn(len(warriors))].HealBy(healAmount)
}

func (a *attackAction) applyAmbushEffect(effect types.AmbushEffect, effectiveWeapon cards.Weapon, weaponType types.WeaponType, g attackGame) *Result {
	result := &Result{
		Action: types.LastActionAmbush,
		Attack: &AttackDetails{
			WeaponID:           a.weaponID,
			TargetID:           a.targetID,
			TargetPlayer:       a.targetPlayerName,
			AmbushEffect:       effect,
			AmbushAttackerName: a.playerName,
		},
	}

	switch effect {
	case types.AmbushEffectReflectDamage:
		// Reflected damage equals the exact amount the original target would have received.
		// The damage is reflected back to the warrior who performed the attack.
		multiplier := 1
		if targetWarrior, ok := a.target.(cards.Warrior); ok {
			multiplier = effectiveWeapon.MultiplierFactor(targetWarrior)
		}
		a.attacker.ReceiveDamage(effectiveWeapon, multiplier)
		g.AddHistory(fmt.Sprintf("%s's attack was reflected — %s takes damage",
			a.currentPlayer.Name(), a.attacker.String()), types.CategoryAction)
		a.currentPlayer.RemoveFromHand(a.weaponID)

	case types.AmbushEffectCancelAttack:
		// Attack cancelled; weapon discarded.
		a.currentPlayer.RemoveFromHand(a.weaponID)
		a.weapon.GetCardMovedToPileObserver().OnCardMovedToPile(a.weapon)
		g.AddHistory(fmt.Sprintf("%s's attack was cancelled by an ambush",
			a.currentPlayer.Name()), types.CategoryAction)

	case types.AmbushEffectStealWeapon:
		// Weapon transferred to the defender's hand (bypasses hand limit — forced effect).
		a.currentPlayer.RemoveFromHand(a.weaponID)
		a.targetPlayer.ForceAddCard(a.weapon)
		g.AddHistory(fmt.Sprintf("%s's weapon was stolen by an ambush",
			a.currentPlayer.Name()), types.CategoryAction)

	case types.AmbushEffectDrainLife:
		// The attack is absorbed: warrior takes no damage and gains HP equal to the weapon damage.
		// BeAttacked is skipped deliberately — calling it would risk killing the warrior (via
		// dead()) before the heal can run, and the net behaviour is the same as absorbing the hit.
		if warrior, ok := a.target.(cards.Warrior); ok {
			multiplier := effectiveWeapon.MultiplierFactor(warrior)
			warrior.HealBy(effectiveWeapon.DamageAmount() * multiplier)
		}
		a.currentPlayer.RemoveFromHand(a.weaponID)
		a.weapon.GetCardMovedToPileObserver().OnCardMovedToPile(a.weapon)
		g.AddHistory(fmt.Sprintf("%s's attack was drained — target gained HP",
			a.currentPlayer.Name()), types.CategoryAction)

	case types.AmbushEffectInstantKill:
		// The warrior who performed the attack is instantly killed.
		a.attacker.KillByAmbush()
		g.AddHistory(fmt.Sprintf("%s triggered an ambush — %s was instantly killed!",
			a.currentPlayer.Name(), a.attacker.String()), types.CategoryAction)
		a.currentPlayer.RemoveFromHand(a.weaponID)
		a.weapon.GetCardMovedToPileObserver().OnCardMovedToPile(a.weapon)
	}

	return result
}
