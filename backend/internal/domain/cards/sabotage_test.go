package cards

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewSabotage(t *testing.T) {
	s := NewSabotage("sab1")
	assert.Equal(t, "SAB1", s.GetID())
	assert.Equal(t, "Sabotage", s.Name())
}

func TestSabotage_ImplementsInterface(t *testing.T) {
	s := NewSabotage("sab1")
	var _ Sabotage = s
}
