package websocket

import (
	"github.com/alelopezbcn/thecampaign/internal/domain"
	"github.com/alelopezbcn/thecampaign/internal/domain/ports"
)

// CardDTO represents a card for JSON serialization
type CardDTO struct {
	ID          string `json:"id"`
	Type        string `json:"type"`
	Name        string `json:"name"`
	HitPoints   int    `json:"hit_points,omitempty"`
	Damage      int    `json:"damage_amount,omitempty"`
	Value       int    `json:"value,omitempty"`
	IsProtected bool   `json:"is_protected,omitempty"`
}

// CastleDTO represents a castle for JSON serialization
type CastleDTO struct {
	Constructed   bool `json:"constructed"`
	ResourceCards int  `json:"resource_cards"`
}

// ConvertCard converts a ports.Card to CardDTO
func ConvertCard(card ports.Card) CardDTO {
	dto := CardDTO{
		ID:   card.GetID(),
		Name: card.String(),
	}

	// Determine type and extract specific fields
	switch c := card.(type) {
	case ports.Warrior:
		dto.Type = string(c.Type())
		dto.HitPoints = c.Health()
		isProtected, _ := c.IsProtected()
		dto.IsProtected = isProtected
	case ports.Weapon:
		dto.Type = string(c.Type())
		dto.Damage = c.DamageAmount()
	case ports.Resource:
		dto.Type = "Gold"
		dto.Value = c.Value()
	case ports.Spy:
		dto.Type = "Spy"
	case ports.Thief:
		dto.Type = "Thief"
	case ports.Catapult:
		dto.Type = "Catapult"
	case ports.SpecialPower:
		dto.Type = "SpecialPower"
	default:
		dto.Type = "Unknown"
	}

	return dto
}

// ConvertCards converts a slice of cards to DTOs
func ConvertCards(cards []ports.Card) []CardDTO {
	dtos := make([]CardDTO, len(cards))
	for i, card := range cards {
		dtos[i] = ConvertCard(card)
	}
	return dtos
}

// ConvertWarriors converts a slice of warriors to DTOs
func ConvertWarriors(warriors []ports.Warrior) []CardDTO {
	dtos := make([]CardDTO, len(warriors))
	for i, warrior := range warriors {
		dtos[i] = ConvertCard(warrior)
	}
	return dtos
}

// ConvertCastle converts a castle to DTO
func ConvertCastle(castle ports.Castle) CastleDTO {
	if castle == nil {
		return CastleDTO{Constructed: false, ResourceCards: 0}
	}

	return CastleDTO{
		Constructed:   castle.IsConstructed(),
		ResourceCards: castle.ResourceCards(),
	}
}

// ConvertGameStatus converts domain.GameStatus to GameStatusDTO
func ConvertGameStatus(status domain.GameStatus) GameStatusDTO {
	return GameStatusDTO{
		CurrentPlayer:              status.CurrentPlayer,
		WarriorsInHandIDs:          status.WarriorsInHandIDs,
		UsableWeaponIDs:            status.UsableWeaponIDs,
		SpyID:                      status.SpyID,
		ThiefID:                    status.ThiefID,
		ResourceIDs:                status.ResourceIDs,
		SpecialPowerStatus:         ConvertSpecialPowerStatus(status.SpecialPowerStatus),
		ConstructionIDs:            status.ConstructionIDs,
		CatapultID:                 status.CatapultID,
		CurrentPlayerHand:          ConvertCards(status.CurrentPlayerHand),
		CurrentPlayerField:         ConvertWarriors(status.CurrentPlayerField),
		CurrentPlayerCastle:        ConvertCastle(status.CurrentPlayerCastle),
		EnemyField:                 ConvertWarriors(status.EnemyField),
		EnemyCastle:                ConvertCastle(status.EnemyCastle),
		CardsInEnemyHand:           status.CardsInEnemyHand,
		ResourceCardsInEnemyCastle: status.ResourceCardsInEnemyCastle,
	}
}

// ConvertSpecialPowerStatus converts domain special power status to DTO
func ConvertSpecialPowerStatus(status domain.SpecialPowerStatus) SpecialPowerStatusDTO {
	return SpecialPowerStatusDTO{
		SpecialPowerIDs:   status.SpecialPowerIDs,
		CanHealIDs:        status.CanHealIDs,
		CanInstantKillIDs: status.CanInstantKillIDs,
		CanProtectIDs:     status.CanProtectIDs,
	}
}
