package types

import "math/rand"

// AmbushEffect represents the random outcome of an Ambush card trigger.
type AmbushEffect int

const (
	AmbushEffectReflectDamage AmbushEffect = iota // 23% — full weapon damage reflected to attacker warrior
	AmbushEffectCancelAttack                      // 23% — attack cancelled, weapon discarded
	AmbushEffectStealWeapon                       // 23% — weapon added to defender's hand
	AmbushEffectDrainLife                         // 23% — attack absorbed; target heals HP equal to weapon damage
	AmbushEffectInstantKill                       //  8% — random attacker warrior is killed
)

// RandomAmbushEffect returns a randomly selected AmbushEffect using weighted probabilities.
// Assigned at card creation time; the effect stays hidden until the ambush triggers.
func RandomAmbushEffect() AmbushEffect {
	r := rand.Intn(100)
	switch {
	case r < 23:
		return AmbushEffectReflectDamage
	case r < 46:
		return AmbushEffectCancelAttack
	case r < 69:
		return AmbushEffectStealWeapon
	case r < 92:
		return AmbushEffectDrainLife
	default:
		return AmbushEffectInstantKill
	}
}

// DisplayName returns a human-readable name for the effect shown in the UI.
func (e AmbushEffect) DisplayName() string {
	switch e {
	case AmbushEffectReflectDamage:
		return "Reflect Damage"
	case AmbushEffectCancelAttack:
		return "Attack Cancelled"
	case AmbushEffectStealWeapon:
		return "Weapon Stolen"
	case AmbushEffectDrainLife:
		return "Drain Life"
	case AmbushEffectInstantKill:
		return "Instant Kill"
	default:
		return "Unknown"
	}
}
