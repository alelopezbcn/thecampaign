package cards_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewSword(t *testing.T) {
	s := NewSword("s1", 7)
	assert.Equal(t, "S1", s.GetID())
	assert.Equal(t, 7, s.DamageAmount())
	// String() returns format: "WeaponType (DamageAmount)"
	assert.Contains(t, s.String(), "Sword")
	assert.Contains(t, s.String(), "7")
}

func TestNewArrow(t *testing.T) {
	a := NewArrow("a1", 3)
	assert.Equal(t, "A1", a.GetID())
	assert.Equal(t, 3, a.DamageAmount())
	// String() returns format: "WeaponType (DamageAmount)"
	assert.Contains(t, a.String(), "Arrow")
	assert.Contains(t, a.String(), "3")
}

func TestNewPoison(t *testing.T) {
	p := NewPoison("p1", 5)
	assert.Equal(t, "P1", p.GetID())
	assert.Equal(t, 5, p.DamageAmount())
	// String() returns format: "WeaponType (DamageAmount)"
	assert.Contains(t, p.String(), "Poison")
	assert.Contains(t, p.String(), "5")
}

func TestSword_String(t *testing.T) {
	s := NewSword("s2", 10)
	str := s.String()
	// String() returns format: "WeaponType (DamageAmount)"
	assert.Contains(t, str, "Sword")
	assert.Contains(t, str, "10")
}

func TestArrow_String(t *testing.T) {
	a := NewArrow("a2", 4)
	str := a.String()
	// String() returns format: "WeaponType (DamageAmount)"
	assert.Contains(t, str, "Arrow")
	assert.Contains(t, str, "4")
}

func TestPoison_String(t *testing.T) {
	p := NewPoison("p2", 8)
	str := p.String()
	// String() returns format: "WeaponType (DamageAmount)"
	assert.Contains(t, str, "Poison")
	assert.Contains(t, str, "8")
}
