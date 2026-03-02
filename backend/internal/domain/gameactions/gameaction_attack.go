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
	targetPlayerName string
	targetID         string
	weaponID         string

	currentPlayer board.Player
	targetPlayer  board.Player
	target        cards.Attackable
	weapon        cards.Weapon
}

func NewAttackAction(playerName, targetPlayerName, targetID, weaponID string) *attackAction {
	return &attackAction{
		playerName:       playerName,
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

	return nil
}

func (a *attackAction) Execute(g Game) (*Result, func() gamestatus.GameStatus, error) {
	return a.execute(g)
}

func (a *attackAction) execute(g attackGame) (*Result, func() gamestatus.GameStatus, error) {
	// Apply Curse event modifier to weapon damage (cached to avoid repeated Type() calls).
	weaponType := a.weapon.Type()
	weaponMod := g.EventHandler().WeaponDamageModifier(weaponType)
	effectiveWeapon := applyWeaponModifier(a.weapon, weaponMod)

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

	// Normal attack.
	err := a.target.BeAttacked(effectiveWeapon)
	if err != nil {
		result := &Result{}
		return result, nil, fmt.Errorf("attack action failed: %w", err)
	}

	a.currentPlayer.RemoveFromHand(a.weaponID)

	g.AddHistory(fmt.Sprintf("%s attacked %s with %s",
		a.currentPlayer.Name(), a.target.String(), a.weapon.String()),
		types.CategoryAction)

	result := &Result{
		Action:             types.LastActionAttack,
		AttackWeaponID:     a.weaponID,
		AttackTargetID:     a.targetID,
		AttackTargetPlayer: a.targetPlayerName,
	}
	statusFn := func() gamestatus.GameStatus {
		return g.Status(a.currentPlayer)
	}

	return result, statusFn, nil
}

func (a *attackAction) NextPhase() types.PhaseType {
	return types.PhaseTypeSpySteal
}

// applyAmbushEffect applies the pre-determined ambush effect and returns the result.
// effectiveWeapon carries the Curse-adjusted damage; weaponType is cached to avoid extra Type() calls.
func (a *attackAction) applyAmbushEffect(effect types.AmbushEffect, effectiveWeapon cards.Weapon, weaponType types.WeaponType, g attackGame) *Result {
	result := &Result{
		Action:             types.LastActionAmbush,
		AttackWeaponID:     a.weaponID,
		AttackTargetID:     a.targetID,
		AttackTargetPlayer: a.targetPlayerName,
		AmbushEffect:       effect,
		AmbushAttackerName: a.playerName,
	}

	switch effect {
	case types.AmbushEffectReflectDamage:
		// Reflected damage equals the exact amount the original target would have received.
		w := findWarriorForWeapon(a.currentPlayer.Field(), weaponType)
		if w != nil {
			multiplier := 1
			if targetWarrior, ok := a.target.(cards.Warrior); ok {
				multiplier = effectiveWeapon.MultiplierFactor(targetWarrior)
			}
			w.ReceiveDamage(effectiveWeapon, multiplier)
			g.AddHistory(fmt.Sprintf("%s's attack was reflected — %s takes damage",
				a.currentPlayer.Name(), w.String()), types.CategoryAction)
		} else {
			g.AddHistory(fmt.Sprintf("%s's attack was reflected (no target found)",
				a.currentPlayer.Name()), types.CategoryAction)
		}
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
		// A random warrior in the attacker's field is instantly killed.
		w := findAnyWarrior(a.currentPlayer.Field())
		if w != nil {
			w.KillByAmbush()
			g.AddHistory(fmt.Sprintf("%s triggered an ambush — %s was instantly killed!",
				a.currentPlayer.Name(), w.String()), types.CategoryAction)
		} else {
			g.AddHistory(fmt.Sprintf("%s triggered an ambush — instant kill (no warriors found)",
				a.currentPlayer.Name()), types.CategoryAction)
		}
		a.currentPlayer.RemoveFromHand(a.weaponID)
		a.weapon.GetCardMovedToPileObserver().OnCardMovedToPile(a.weapon)
	}

	return result
}

// findWarriorForWeapon finds a warrior in the field that matches the weapon type.
// Falls back to any random warrior if no type match exists.
func findWarriorForWeapon(f board.FieldReader, weaponType types.WeaponType) cards.Warrior {
	warriors := f.Warriors()
	if len(warriors) == 0 {
		return nil
	}

	var compatible []cards.Warrior
	for _, w := range warriors {
		switch weaponType {
		case types.SwordWeaponType:
			if w.Type() == types.KnightWarriorType || w.Type() == types.DragonWarriorType || w.Type() == types.MercenaryWarriorType {
				compatible = append(compatible, w)
			}
		case types.ArrowWeaponType:
			if w.Type() == types.ArcherWarriorType || w.Type() == types.DragonWarriorType || w.Type() == types.MercenaryWarriorType {
				compatible = append(compatible, w)
			}
		case types.PoisonWeaponType:
			if w.Type() == types.MageWarriorType || w.Type() == types.DragonWarriorType || w.Type() == types.MercenaryWarriorType {
				compatible = append(compatible, w)
			}
		default:
			compatible = append(compatible, w)
		}
	}

	if len(compatible) > 0 {
		return compatible[rand.Intn(len(compatible))]
	}
	return warriors[rand.Intn(len(warriors))]
}

// findAnyWarrior returns a random warrior from the field, or nil if empty.
func findAnyWarrior(f board.FieldReader) cards.Warrior {
	warriors := f.Warriors()
	if len(warriors) == 0 {
		return nil
	}
	return warriors[rand.Intn(len(warriors))]
}
