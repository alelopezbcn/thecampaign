package gamestatus_test

import (
	"testing"

	"github.com/alelopezbcn/thecampaign/internal/domain/gamestatus"
	"github.com/alelopezbcn/thecampaign/internal/domain/types"
	"github.com/stretchr/testify/assert"
)

func TestNewHistoryLine_MsgIsPreserved(t *testing.T) {
	hl := gamestatus.NewHistoryLine("Player1 attacked Player2", types.CategoryAction)

	assert.Equal(t, "Player1 attacked Player2", hl.Msg)
}

func TestNewHistoryLine_ColorMappings(t *testing.T) {
	tests := []struct {
		category  types.Category
		wantColor string
	}{
		{types.CategoryAction, "#33C1FF"},
		{types.CategoryInfo, "#22a25a"},
		{types.CategoryError, "#FF3333"},
		{types.CategorySkip, "#959896"},
		{types.CategoryEndTurn, "#F39C12"},
		{types.CategoryTurnExpired, "#f65b07"},
		{types.CategoryElimination, "#8E44AD"},
	}

	for _, tt := range tests {
		t.Run(string(tt.category), func(t *testing.T) {
			hl := gamestatus.NewHistoryLine("msg", tt.category)

			assert.Equal(t, tt.wantColor, hl.Color)
		})
	}
}

func TestNewHistoryLine_UnknownCategoryDefaultsToWhite(t *testing.T) {
	hl := gamestatus.NewHistoryLine("msg", types.Category("unknown"))

	assert.Equal(t, "#FFFFFF", hl.Color)
}
