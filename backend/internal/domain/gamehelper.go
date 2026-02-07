package domain

import (
	"errors"
	"fmt"

	"github.com/alelopezbcn/thecampaign/internal/domain/ports"
	"github.com/alelopezbcn/thecampaign/internal/domain/types"
)

// CurrentPlayer returns the player whose turn it is
func (g *Game) CurrentPlayer() ports.Player {
	return g.Players[g.CurrentTurn]
}

// GetPlayer returns a player by name, or nil if not found
func (g *Game) GetPlayer(name string) ports.Player {
	for _, p := range g.Players {
		if p.Name() == name {
			return p
		}
	}
	return nil
}

// PlayerIndex returns the index of a player by name, or -1
func (g *Game) PlayerIndex(name string) int {
	for i, p := range g.Players {
		if p.Name() == name {
			return i
		}
	}
	return -1
}

// Enemies returns all opponents (non-eliminated, non-ally) of a given player
func (g *Game) Enemies(playerIdx int) []ports.Player {
	var enemies []ports.Player
	for i, p := range g.Players {
		if i == playerIdx {
			continue
		}
		if g.EliminatedPlayers[i] {
			continue
		}
		if g.Mode == types.GameMode2v2 && g.SameTeam(playerIdx, i) {
			continue
		}
		enemies = append(enemies, p)
	}
	return enemies
}

// Allies returns teammates (for 2v2 only, excluding self)
func (g *Game) Allies(playerIdx int) []ports.Player {
	if g.Mode != types.GameMode2v2 {
		return nil
	}
	var allies []ports.Player
	for i, p := range g.Players {
		if i == playerIdx {
			continue
		}
		if g.SameTeam(playerIdx, i) {
			allies = append(allies, p)
		}
	}
	return allies
}

// SameTeam checks if two player indices are on the same team
func (g *Game) SameTeam(i, j int) bool {
	if g.Mode != types.GameMode2v2 {
		return false
	}
	for _, team := range g.Teams {
		hasI, hasJ := false, false
		for _, idx := range team {
			if idx == i {
				hasI = true
			}
			if idx == j {
				hasJ = true
			}
		}
		if hasI && hasJ {
			return true
		}
	}
	return false
}

func (g *Game) getTargetPlayer(playerName string, targetPlayerName string) (
	ports.Player, error) {

	// Validate target player is an enemy
	targetPlayer := g.GetPlayer(targetPlayerName)
	if targetPlayer == nil {
		return nil, fmt.Errorf("target player %s not found", targetPlayerName)
	}

	pIdx := g.PlayerIndex(playerName)
	tIdx := g.PlayerIndex(targetPlayerName)

	if pIdx == tIdx {
		return nil, errors.New("cannot attack yourself")
	}

	if g.SameTeam(pIdx, tIdx) {
		return nil, errors.New("cannot attack your ally")
	}

	if g.EliminatedPlayers[tIdx] {
		return nil, errors.New("cannot attack eliminated player")
	}

	return targetPlayer, nil
}
