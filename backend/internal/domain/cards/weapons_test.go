package cards

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewSword(t *testing.T) {
	s := NewSword("s1", 7)
	assert.Equal(t, "S1", s.GetID())
	assert.Equal(t, 7, s.DamageAmount())
	assert.Contains(t, s.String(), "7 Sword (S1)")
}

func TestNewArrow(t *testing.T) {
	a := NewArrow("a1", 3)
	assert.Equal(t, "A1", a.GetID())
	assert.Equal(t, 3, a.DamageAmount())
	assert.Contains(t, a.String(), "3 Arrow (A1)")
}

func TestNewPoison(t *testing.T) {
	p := NewPoison("p1", 5)
	assert.Equal(t, "P1", p.GetID())
	assert.Equal(t, 5, p.DamageAmount())
	assert.Contains(t, p.String(), "5 Poison (P1)")
}

func TestSword_String(t *testing.T) {
	s := NewSword("s2", 10)
	str := s.String()
	assert.Contains(t, str, "10 Sword (S2)")
}

func TestArrow_String(t *testing.T) {
	a := NewArrow("a2", 4)
	str := a.String()
	assert.Contains(t, str, "4 Arrow (A2)")
}

func TestPoison_String(t *testing.T) {
	p := NewPoison("p2", 8)
	str := p.String()
	assert.Contains(t, str, "8 Poison (P2)")
}
