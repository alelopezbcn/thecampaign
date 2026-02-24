package cards

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewFortress(t *testing.T) {
	f := NewFortress("fw1")
	assert.Equal(t, "FW1", f.GetID())
	assert.Equal(t, "Fortress", f.Name())
}

func TestFortress_ImplementsInterface(t *testing.T) {
	f := NewFortress("fw1")
	var _ Fortress = f
}
