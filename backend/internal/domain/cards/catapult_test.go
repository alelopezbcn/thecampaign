package cards

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
)

type fakeCastleTarget struct {
	gold Resource
	err  error
}

func (f *fakeCastleTarget) RemoveGold(_ int) (Resource, error) {
	return f.gold, f.err
}

func TestNewCatapultCard(t *testing.T) {
	c := NewCatapultCard("c1")
	assert.Equal(t, "C1", c.GetID())
	assert.Equal(t, "Catapult", c.Name())
}

func TestCatapult_Attack_Success(t *testing.T) {
	c := NewCatapultCard("c1")
	gold := NewGold("g1", 1)
	castle := &fakeCastleTarget{gold: gold}

	result, err := c.Attack(castle, 0)

	assert.NoError(t, err)
	assert.Equal(t, gold, result)
}

func TestCatapult_Attack_Error(t *testing.T) {
	c := NewCatapultCard("c1")
	castle := &fakeCastleTarget{err: errors.New("no gold at position")}

	result, err := c.Attack(castle, 3)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "no gold at position")
	assert.Nil(t, result)
}
