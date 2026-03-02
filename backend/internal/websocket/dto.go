package websocket

import (
	"github.com/alelopezbcn/thecampaign/internal/domain/gamestatus"
)

// CardDTO represents a card for JSON serialization
type CardDTO struct {
	ID      string `json:"id"`
	Type    string `json:"type"`
	SubType string `json:"sub_type,omitempty"`
	Color   string `json:"color"`
	Value   int    `json:"value,omitempty"`
}

// CastleDTO represents a castle for JSON serialization
type CastleDTO struct {
	Constructed   bool `json:"constructed"`
	IsProtected   bool `json:"is_protected"`
	ResourceCards int  `json:"resource_cards"`
	Value         int  `json:"value"`
}

// CemeteryDTO represents the cemetery for JSON serialization
type CemeteryDTO struct {
	Corps    int      `json:"corps"`
	LastCorp *CardDTO `json:"last_corp,omitempty"`
}

// DiscardPileDTO represents the discard pile for JSON serialization
type DiscardPileDTO struct {
	Cards    int      `json:"cards"`
	LastCard *CardDTO `json:"last_card,omitempty"`
}

// ConvertGameStatus converts domain.GameStatus to GameStatusDTO
func ConvertGameStatus(status gamestatus.GameStatus) GameStatusDTO {
	opponents := make([]OpponentStatusDTO, len(status.Opponents))
	for i, opp := range status.Opponents {
		opponents[i] = OpponentStatusDTO{
			PlayerName:     opp.PlayerName,
			Field:          convertFieldCards(opp.Field),
			Castle:         convertCastle(opp.Castle),
			CardsInHand:    opp.CardsInHand,
			IsAlly:         opp.IsAlly,
			IsEliminated:   opp.IsEliminated,
			IsDisconnected: opp.IsDisconnected,
			AmbushInField:  opp.AmbushInField,
		}
	}

	return GameStatusDTO{
		CurrentPlayer:  status.CurrentPlayer,
		TurnPlayer:     status.TurnPlayer,
		CurrentAction:  status.CurrentAction,
		LastAction:     string(status.LastAction),
		NewCards:       status.NewCards,
		CanMoveWarrior: status.CanMoveWarrior,
		CanTrade:       status.CanTrade,

		CurrentPlayerHand:          convertHandCards(status.CurrentPlayerHand),
		CurrentPlayerField:         convertFieldCards(status.CurrentPlayerField),
		CurrentPlayerCastle:        convertCastle(status.CurrentPlayerCastle),
		IsEliminated:               status.IsEliminated,
		IsDisconnected:             status.IsDisconnected,
		Opponents:                  opponents,
		GameMode:                   status.GameMode,
		Cemetery:                   convertCemetery(status.Cemetery),
		DiscardPile:                convertDiscardPile(status.DiscardPile),
		CardsInDeck:                status.CardsInDeck,
		ModalCards:                 convertModalCards(status.ModalCards),
		LastMovedWarriorID:         status.LastMovedWarriorID,
		LastAttackWeaponID:         status.LastAttackWeaponID,
		LastAttackTargetID:         status.LastAttackTargetID,
		LastAttackTargetPlayer:     status.LastAttackTargetPlayer,
		StolenFromYouCard:          convertModalCards(status.StolenFromYouCard),
		SabotagedFromYouCard:       convertModalCards(status.SabotagedFromYouCard),
		SpyNotification:            status.SpyNotification,
		AmbushTriggered:            convertAmbushTriggered(status.AmbushTriggered),
		DesertionNotification:      convertDesertionNotification(status.DesertionNotification),
		CurrentPlayerAmbushInField: status.CurrentPlayerAmbushInField,
		History:                    convertHistory(status.History),
		PlayersOrder:               status.PlayersOrder,
		NextTurnPlayer:             status.NextTurnPlayer,
		GameOverMsg:                status.GameOverMsg,
		IsWinner:                   status.IsWinner,
		GameStartedAt:              status.GameStartedAt.UTC().Format("2006-01-02T15:04:05Z"),
		TurnStartedAt:              status.TurnStartedAt.UTC().Format("2006-01-02T15:04:05Z"),
		TurnTimeLimitSecs:          status.TurnTimeLimitSecs,
	}
}

func convertAmbushTriggered(at *gamestatus.AmbushTrigger) *AmbushTriggeredDTO {
	if at == nil {
		return nil
	}
	return &AmbushTriggeredDTO{
		Effect:        int(at.Effect),
		EffectDisplay: at.EffectDisplay,
	}
}

func convertDesertionNotification(dn *gamestatus.DesertionNotification) *DesertionNotificationDTO {
	if dn == nil {
		return nil
	}
	return &DesertionNotificationDTO{
		WarriorCard: CardDTO{
			ID:      dn.WarriorCard.CardID,
			Type:    dn.WarriorCard.CardType.Name,
			SubType: dn.WarriorCard.CardType.SubName,
			Color:   dn.WarriorCard.CardType.Color,
			Value:   dn.WarriorCard.Value,
		},
		StolenBy: dn.StolenBy,
	}
}

func convertModalCards(cards []gamestatus.Card) []CardDTO {
	if cards == nil {
		return nil
	}
	dtos := make([]CardDTO, len(cards))
	for i, card := range cards {
		dtos[i] = CardDTO{
			ID:      card.CardID,
			Type:    card.CardType.Name,
			SubType: card.CardType.SubName,
			Color:   card.CardType.Color,
			Value:   card.Value,
		}
	}
	return dtos
}

func convertHandCards(cards []gamestatus.HandCard) []HandCardDTO {
	dtos := make([]HandCardDTO, len(cards))
	for i, card := range cards {
		dtos[i] = HandCardDTO{
			CardDTO: CardDTO{
				ID:      card.Card.CardID,
				Type:    card.Card.CardType.Name,
				SubType: card.Card.CardType.SubName,
				Color:   card.Card.CardType.Color,
				Value:   card.Card.Value,
			},
			CanBeUsedOnIDs: card.CanBeUsedOnIDs,
			CanBeUsed:      card.CanBeUsed,
			DmgMultiplier:  card.DmgMultiplier,
			CanBeTraded:    card.CanBeTraded,
		}
	}
	return dtos
}

func convertFieldCards(cards []gamestatus.FieldCard) []FieldCardDTO {
	dtos := make([]FieldCardDTO, len(cards))
	for i, card := range cards {
		dto := FieldCardDTO{
			CardDTO: CardDTO{
				ID:      card.Card.CardID,
				Type:    card.Card.CardType.Name,
				SubType: card.Card.CardType.SubName,
				Color:   card.Card.CardType.Color,
				Value:   card.Card.Value,
			},
		}

		if len(card.AttackedBy) > 0 {
			dto.AttackedBy = make([]CardDTO, len(card.AttackedBy))
			for j, attacker := range card.AttackedBy {
				dto.AttackedBy[j] = CardDTO{
					ID:      attacker.CardID,
					Type:    attacker.CardType.Name,
					SubType: attacker.CardType.SubName,
					Color:   attacker.CardType.Color,
					Value:   attacker.Value,
				}
			}
		}

		if card.ProtectedBy.CardID != "" {
			dto.ProtectedBy = &CardDTO{
				ID:      card.ProtectedBy.CardID,
				Type:    card.ProtectedBy.CardType.Name,
				SubType: card.ProtectedBy.CardType.SubName,
				Color:   card.ProtectedBy.CardType.Color,
				Value:   card.ProtectedBy.Value,
			}
		}

		dtos[i] = dto
	}
	return dtos
}

func convertCastle(castle gamestatus.Castle) CastleDTO {
	return CastleDTO{
		Constructed:   castle.IsConstructed,
		IsProtected:   castle.IsProtected,
		ResourceCards: castle.ResourceCards,
		Value:         castle.Value,
	}
}

func convertCemetery(cemetery gamestatus.Cemetery) CemeteryDTO {
	dto := CemeteryDTO{
		Corps: cemetery.Corps,
	}

	if cemetery.LastCorp.CardID != "" {
		dto.LastCorp = &CardDTO{
			ID:      cemetery.LastCorp.CardID,
			Type:    cemetery.LastCorp.CardType.Name,
			SubType: cemetery.LastCorp.CardType.SubName,
			Color:   cemetery.LastCorp.CardType.Color,
			Value:   cemetery.LastCorp.Value,
		}
	}

	return dto
}

func convertDiscardPile(discardPile gamestatus.DiscardPile) DiscardPileDTO {
	dto := DiscardPileDTO{
		Cards: discardPile.Cards,
	}

	if discardPile.LastCard.CardID != "" {
		dto.LastCard = &CardDTO{
			ID:      discardPile.LastCard.CardID,
			Type:    discardPile.LastCard.CardType.Name,
			SubType: discardPile.LastCard.CardType.SubName,
			Color:   discardPile.LastCard.CardType.Color,
			Value:   discardPile.LastCard.Value,
		}
	}

	return dto
}

func convertHistory(historyLine []gamestatus.HistoryLine) []HistoryLineDTO {
	if historyLine == nil {
		return nil
	}
	dtos := make([]HistoryLineDTO, len(historyLine))
	for i, hl := range historyLine {
		dtos[i] = HistoryLineDTO{
			Msg:   hl.Msg,
			Color: hl.Color,
		}
	}
	return dtos
}
