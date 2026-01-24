package websocket

import (
	"github.com/alelopezbcn/thecampaign/internal/domain/gamestatus"
)

// CardDTO represents a card for JSON serialization
type CardDTO struct {
	ID    string `json:"id"`
	Type  string `json:"type"`
	Color string `json:"color"`
	Value int    `json:"value,omitempty"`
}

// CastleDTO represents a castle for JSON serialization
type CastleDTO struct {
	Constructed   bool `json:"constructed"`
	ResourceCards int  `json:"resource_cards"`
	Value         int  `json:"value"`
}

// ConvertGameStatus converts gamestatus.GameStatus to GameStatusDTO
func ConvertGameStatus(status gamestatus.GameStatus) GameStatusDTO {
	return GameStatusDTO{
		CurrentPlayer:     status.CurrentPlayer,
		CanMoveWarrior:    status.CanMoveWarrior,
		CanAttack:         status.CanAttack,
		CanCatapult:       status.CanCatapult,
		CanSpy:            status.CanSpy,
		CanSteal:          status.CanSteal,
		CanBuy:            status.CanBuy,
		CanInitiateCastle: status.CanInitiateCastle,
		CanGrowCastle:     status.CanGrowCastle,

		CurrentPlayerHand:          convertHandCards(status.CurrentPlayerHand),
		CurrentPlayerField:         convertFieldCards(status.CurrentPlayerField),
		CurrentPlayerCastle:        convertCastle(status.CurrentPlayerCastle),
		EnemyField:                 convertFieldCards(status.EnemyField),
		EnemyCastle:                convertCastle(status.EnemyCastle),
		CardsInEnemyHand:           status.CardsInEnemyHand,
		ResourceCardsInEnemyCastle: status.ResourceCardsInEnemyCastle,
	}
}

func convertHandCards(cards []gamestatus.HandCard) []HandCardDTO {
	dtos := make([]HandCardDTO, len(cards))
	for i, card := range cards {
		dtos[i] = HandCardDTO{
			CardDTO: CardDTO{
				ID:    card.Card.CardID,
				Type:  card.Card.CardType.Name,
				Color: card.Card.CardType.Color,
				Value: card.Card.Value,
			},
			CanBeUsedOnIDs: card.CanBeUsedOnIDs,
			CanConstruct:   card.CanConstruct,
		}
	}
	return dtos
}

func convertFieldCards(cards []gamestatus.FieldCard) []FieldCardDTO {
	dtos := make([]FieldCardDTO, len(cards))
	for i, card := range cards {
		dto := FieldCardDTO{
			CardDTO: CardDTO{
				ID:    card.Card.CardID,
				Type:  card.Card.CardType.Name,
				Color: card.Card.CardType.Color,
				Value: card.Card.Value,
			},
		}

		if len(card.AttackedBy) > 0 {
			dto.AttackedBy = make([]CardDTO, len(card.AttackedBy))
			for j, attacker := range card.AttackedBy {
				dto.AttackedBy[j] = CardDTO{
					ID:    attacker.CardID,
					Type:  attacker.CardType.Name,
					Color: attacker.CardType.Color,
					Value: attacker.Value,
				}
			}
		}

		if card.ProtectedBy.CardID != "" {
			dto.ProtectedBy = &CardDTO{
				ID:    card.ProtectedBy.CardID,
				Type:  card.ProtectedBy.CardType.Name,
				Color: card.ProtectedBy.CardType.Color,
				Value: card.ProtectedBy.Value,
			}
		}

		dtos[i] = dto
	}
	return dtos
}

func convertCastle(castle gamestatus.Castle) CastleDTO {
	return CastleDTO{
		Constructed:   castle.IsConstructed,
		ResourceCards: castle.ResourceCards,
		Value:         castle.Value,
	}
}
