package cards

import (
	"testing"

	"github.com/alelopezbcn/thecampaign/internal/domain/types"
	"github.com/stretchr/testify/assert"
)

// --- local stubs (no test/mocks to avoid import cycle) ---

type fakeWeapon struct {
	id       string
	damage   int
	observer CardMovedToPileObserver
}

func (f *fakeWeapon) GetID() string                              { return f.id }
func (f *fakeWeapon) Name() string                               { return "fakeWeapon" }
func (f *fakeWeapon) AddCardMovedToPileObserver(o CardMovedToPileObserver) { f.observer = o }
func (f *fakeWeapon) GetCardMovedToPileObserver() CardMovedToPileObserver { return f.observer }
func (f *fakeWeapon) DamageAmount() int                          { return f.damage }
func (f *fakeWeapon) Type() types.WeaponType                     { return types.SwordWeaponType }
func (f *fakeWeapon) CanConstruct() bool                         { return false }
func (f *fakeWeapon) MultiplierFactor(_ Warrior) int             { return 1 }
func (f *fakeWeapon) CanBeUsedWith(_ FieldChecker) bool          { return true }
func (f *fakeWeapon) String() string                             { return "fakeWeapon" }

type fakeSP struct {
	id              string
	observer        CardMovedToPileObserver
	destroyedCalled bool
	receiveDamageFn func(Weapon, int) bool
}

func (f *fakeSP) GetID() string                              { return f.id }
func (f *fakeSP) Name() string                               { return "fakeSP" }
func (f *fakeSP) AddCardMovedToPileObserver(o CardMovedToPileObserver) { f.observer = o }
func (f *fakeSP) GetCardMovedToPileObserver() CardMovedToPileObserver { return f.observer }
func (f *fakeSP) DamageAmount() int                          { return specialPowerDamage }
func (f *fakeSP) Type() types.WeaponType                     { return types.SpecialPowerWeaponType }
func (f *fakeSP) CanConstruct() bool                         { return false }
func (f *fakeSP) MultiplierFactor(_ Warrior) int             { return 1 }
func (f *fakeSP) CanBeUsedWith(_ FieldChecker) bool          { return true }
func (f *fakeSP) String() string                             { return "fakeSP" }
func (f *fakeSP) BeAttacked(_ Weapon) error                  { return nil }
func (f *fakeSP) AttackedBy() []Weapon                       { return nil }
func (f *fakeSP) Health() int                                { return specialPowerMaxHealth }
func (f *fakeSP) ReceiveDamage(w Weapon, m int) bool {
	if f.receiveDamageFn != nil {
		return f.receiveDamageFn(w, m)
	}
	return false
}
func (f *fakeSP) Use(_ Warrior, _ Warrior) error { return nil }
func (f *fakeSP) Destroyed()                     { f.destroyedCalled = true }

type fakeCardObs struct{ called []Card }

func (f *fakeCardObs) OnCardMovedToPile(c Card) { f.called = append(f.called, c) }

type fakeWarriorDeadObs struct{ called []Warrior }

func (f *fakeWarriorDeadObs) OnWarriorDead(w Warrior) { f.called = append(f.called, w) }

// --- tests ---

func TestReceiveDamage_WithProtection_DoesNotGetDamage_ProtectionNotDestroyed(t *testing.T) {
	weapon := &fakeWeapon{}
	sp := &fakeSP{receiveDamageFn: func(_ Weapon, _ int) bool { return false }}
	w := warriorBase{
		protectedBy:    sp,
		attackableBase: &attackableBase{health: 5},
	}

	defeated := w.ReceiveDamage(weapon, 2)
	assert.False(t, defeated)
	assert.NotNil(t, w.protectedBy)
	assert.Equal(t, 5, w.health)
}

func TestReceiveDamage_WithProtection_DoesNotGetDamage_ProtectionDestroyed(t *testing.T) {
	weapon := &fakeWeapon{}
	sp := &fakeSP{receiveDamageFn: func(_ Weapon, _ int) bool { return true }}
	w := warriorBase{
		protectedBy:    sp,
		attackableBase: &attackableBase{health: 5},
	}

	defeated := w.ReceiveDamage(weapon, 2)
	assert.False(t, defeated)
	assert.Nil(t, w.protectedBy)
	assert.Equal(t, 5, w.health)
}

func TestReceiveDamage_KillsWarrior(t *testing.T) {
	obs := &fakeCardObs{}
	weapon := &fakeWeapon{id: "fw1", damage: 10, observer: obs}
	deadObs := &fakeWarriorDeadObs{}
	w := &warriorBase{
		attackableBase:      &attackableBase{health: 5},
		WarriorDeadObserver: deadObs,
	}

	defeated := w.ReceiveDamage(weapon, 1)
	assert.True(t, defeated)
	assert.Len(t, w.attackedBy, 0)
	assert.LessOrEqual(t, w.health, 0)
	assert.Len(t, obs.called, 1)
	assert.Len(t, deadObs.called, 1)
}

func TestReceiveDamage_ReducesHealth_NoDeath(t *testing.T) {
	weapon := &fakeWeapon{damage: 2}
	w := &warriorBase{
		attackableBase: &attackableBase{health: 5},
	}

	defeated := w.ReceiveDamage(weapon, 1)
	assert.False(t, defeated)
	assert.Equal(t, 3, w.health)
	assert.Len(t, w.attackedBy, 1)
	assert.Equal(t, weapon, w.attackedBy[0])
}

func TestWarriorBase_BeAttacked_ReturnsError(t *testing.T) {
	w := &warriorBase{}
	weapon := &fakeWeapon{}

	err := w.BeAttacked(weapon)
	assert.Error(t, err)
	assert.EqualError(t, err, "should be implemented by concrete warrior types")
}

func TestProtectedBy_SetsProtection(t *testing.T) {
	w := &warriorBase{}
	sp := &fakeSP{}

	w.Protect(sp)
	assert.Equal(t, sp, w.protectedBy)
}

func TestWarriorBase_Heal_RestoresHealthAndDiscardsWeapons(t *testing.T) {
	obs1 := &fakeCardObs{}
	obs2 := &fakeCardObs{}
	weapon1 := &fakeWeapon{id: "w1", damage: 4, observer: obs1}
	weapon2 := &fakeWeapon{id: "w2", damage: 5, observer: obs2}
	spObs := &fakeCardObs{}
	sp := &fakeSP{id: "sp1", observer: spObs}

	w := &warriorBase{
		attackableBase: &attackableBase{
			health:     2,
			attackedBy: []Weapon{weapon1, weapon2},
		},
	}

	w.Heal(sp)
	assert.Equal(t, warriorMaxHealth, w.health)
	assert.Empty(t, w.attackedBy)
	assert.Len(t, obs1.called, 1)
	assert.Len(t, obs2.called, 1)
	assert.Len(t, spObs.called, 1)
}

func TestWarriorBase_Heal_NoWeapons(t *testing.T) {
	spObs := &fakeCardObs{}
	sp := &fakeSP{id: "sp1", observer: spObs}
	w := &warriorBase{
		attackableBase: &attackableBase{
			health:     1,
			attackedBy: []Weapon{},
		},
	}

	w.Heal(sp)
	assert.Equal(t, warriorMaxHealth, w.health)
	assert.Empty(t, w.attackedBy)
	assert.Len(t, spObs.called, 1)
}

func TestWarriorBase_InstantKill_WithProtection(t *testing.T) {
	sp := &fakeSP{}
	w := &warriorBase{protectedBy: sp}

	w.InstantKill(sp)
	assert.True(t, sp.destroyedCalled)
	assert.Nil(t, w.protectedBy)
}

func TestWarriorBase_InstantKill_WithoutProtection(t *testing.T) {
	spObs := &fakeCardObs{}
	sp := &fakeSP{id: "sp1", observer: spObs}
	deadObs := &fakeWarriorDeadObs{}
	w := &warriorBase{
		attackableBase:      &attackableBase{attackedBy: []Weapon{}},
		WarriorDeadObserver: deadObs,
	}

	w.InstantKill(sp)
	assert.Empty(t, w.attackedBy)
	assert.Len(t, spObs.called, 1)
	assert.Len(t, deadObs.called, 1)
}

func TestWarriorBase_String_AliveWithWeapons(t *testing.T) {
	k := NewKnight("k1")
	str := k.String()
	assert.Contains(t, str, "Knight")
	assert.Contains(t, str, "20")
}

func TestWarriorBase_String_AliveNoWeapons(t *testing.T) {
	a := NewArcher("a1")
	str := a.String()
	assert.Contains(t, str, "Archer")
	assert.Contains(t, str, "20")
}

func TestWarriorBase_Resurrect_RestoresHealthAndClearsDamageState(t *testing.T) {
	obs1 := &fakeCardObs{}
	weapon1 := &fakeWeapon{id: "w1", damage: 5, observer: obs1}
	sp := &fakeSP{}

	w := &warriorBase{
		attackableBase: &attackableBase{
			health:     3,
			attackedBy: []Weapon{weapon1},
		},
		protectedBy: sp,
	}

	w.Resurrect()
	assert.Equal(t, warriorMaxHealth, w.health)
	assert.Empty(t, w.attackedBy)
	assert.Nil(t, w.protectedBy)
}

func TestWarriorBase_Resurrect_AlreadyFullHealth(t *testing.T) {
	w := &warriorBase{
		attackableBase: &attackableBase{
			health:     warriorMaxHealth,
			attackedBy: []Weapon{},
		},
	}

	w.Resurrect()
	assert.Equal(t, warriorMaxHealth, w.health)
	assert.Empty(t, w.attackedBy)
	assert.Nil(t, w.protectedBy)
}

func TestWarriorBase_HealToMax_RestoresHealth(t *testing.T) {
	w := &warriorBase{
		attackableBase: &attackableBase{health: 5},
	}
	w.HealToMax()
	assert.Equal(t, warriorMaxHealth, w.health)
}

func TestWarriorBase_KillByAmbush_WithProtection_DestroysProtectionOnly(t *testing.T) {
	sp := &fakeSP{}
	deadObs := &fakeWarriorDeadObs{}
	w := &warriorBase{
		attackableBase:      &attackableBase{health: 20},
		protectedBy:         sp,
		WarriorDeadObserver: deadObs,
	}
	w.KillByAmbush()
	assert.True(t, sp.destroyedCalled, "protection should be destroyed")
	assert.Nil(t, w.protectedBy)
	assert.Empty(t, deadObs.called, "warrior should not die when protected")
}

func TestWarriorBase_KillByAmbush_WithoutProtection_KillsWarrior(t *testing.T) {
	deadObs := &fakeWarriorDeadObs{}
	w := &warriorBase{
		attackableBase:      &attackableBase{health: 20, attackedBy: []Weapon{}},
		WarriorDeadObserver: deadObs,
	}
	w.KillByAmbush()
	assert.LessOrEqual(t, w.health, 0)
	assert.Len(t, deadObs.called, 1, "warrior dead observer should be called")
}

func TestWarriorBase_String_Dead(t *testing.T) {
	m := NewMage("m1")
	str := m.String()
	assert.Contains(t, str, "Mage")
	assert.Contains(t, str, "20")
}
