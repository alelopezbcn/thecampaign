package gamestatus_test

import (
	"testing"

	. "github.com/alelopezbcn/thecampaign/internal/domain/gamestatus"
	"github.com/stretchr/testify/assert"
)

func TestNewCastle_NotConstructed(t *testing.T) {
	c := NewCastle(CastleInput{IsConstructed: false, ResourceCardsCount: 2, Value: 5})

	assert.False(t, c.IsConstructed)
	assert.Equal(t, 2, c.ResourceCards)
	assert.Equal(t, 5, c.Value)
}

func TestNewCastle_Constructed(t *testing.T) {
	c := NewCastle(CastleInput{IsConstructed: true, ResourceCardsCount: 5, Value: 15})

	assert.True(t, c.IsConstructed)
	assert.Equal(t, 5, c.ResourceCards)
	assert.Equal(t, 15, c.Value)
}

func TestNewCastle_Empty(t *testing.T) {
	c := NewCastle(CastleInput{IsConstructed: false, ResourceCardsCount: 0, Value: 0})

	assert.False(t, c.IsConstructed)
	assert.Equal(t, 0, c.ResourceCards)
	assert.Equal(t, 0, c.Value)
}

func TestNewCastle_IsProtected(t *testing.T) {
	t.Run("True when input is protected", func(t *testing.T) {
		c := NewCastle(CastleInput{IsProtected: true})

		assert.True(t, c.IsProtected)
	})

	t.Run("False when input is not protected", func(t *testing.T) {
		c := NewCastle(CastleInput{IsProtected: false})

		assert.False(t, c.IsProtected)
	})
}
