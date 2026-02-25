package gamestatus

import (
	"time"

	"github.com/alelopezbcn/thecampaign/internal/domain/cards"
	"github.com/alelopezbcn/thecampaign/internal/domain/types"
)

type AmbushTrigger struct {
	Effect        types.AmbushEffect `json:"effect"`
	EffectDisplay string             `json:"effect_display"`
}

type GameStatus struct {
	CurrentPlayer             string               `json:"current_player"`
	TurnPlayer                string               `json:"turn_player"`
	CurrentAction             string               `json:"current_action"`
	LastAction                types.LastActionType `json:"last_action,omitempty"`
	NewCards                  []string             `json:"new_cards"`
	CanMoveWarrior            bool                 `json:"can_move_warrior"`
	CanTrade                  bool                 `json:"can_trade"`
	CurrentPlayerHand         []HandCard           `json:"current_player_hand"`
	CurrentPlayerField        []FieldCard          `json:"current_player_field"`
	CurrentPlayerCastle       Castle               `json:"current_player_castle"`
	CurrentPlayerAmbushInField bool                `json:"current_player_ambush_in_field"`
	IsEliminated              bool                 `json:"is_eliminated"`
	IsDisconnected            bool                 `json:"is_disconnected"`
	Opponents                 []OpponentStatus     `json:"opponents"`
	GameMode                  string               `json:"game_mode"`
	Cemetery                  Cemetery             `json:"cemetery"`
	DiscardPile               DiscardPile          `json:"discard_pile"`
	CardsInDeck               int                  `json:"deck"`
	ModalCards                []Card               `json:"modal_cards"`
	LastMovedWarriorID        string               `json:"last_moved_warrior_id,omitempty"`
	LastAttackWeaponID        string               `json:"last_attack_weapon_id,omitempty"`
	LastAttackTargetID        string               `json:"last_attack_target_id,omitempty"`
	LastAttackTargetPlayer    string               `json:"last_attack_target_player,omitempty"`
	StolenFromYouCard         []Card               `json:"stolen_from_you_card,omitempty"`
	SabotagedFromYouCard      []Card               `json:"sabotaged_from_you_card,omitempty"`
	SpyNotification           string               `json:"spy_notification,omitempty"`
	AmbushTriggered           *AmbushTrigger       `json:"ambush_triggered,omitempty"`
	History                   []HistoryLine        `json:"history"`
	PlayersOrder              []string             `json:"players_order"`
	NextTurnPlayer            string               `json:"next_turn_player,omitempty"`
	GameOverMgs               string               `json:"game_over_msg"`
	IsWinner                  bool                 `json:"is_winner"`
	GameStartedAt             time.Time            `json:"game_started_at"`
	TurnStartedAt             time.Time            `json:"turn_started_at"`
	TurnTimeLimitSecs         int                  `json:"turn_time_limit_secs"`
}

type OpponentStatus struct {
	PlayerName     string
	Field          []FieldCard
	Castle         Castle
	CardsInHand    int
	IsAlly         bool
	IsEliminated   bool
	IsDisconnected bool
	AmbushInField  bool
}

func NewGameStatus(dto GameStatusDTO) GameStatus {
	playersOrder := make([]string, len(dto.PlayersNames))
	copy(playersOrder, dto.PlayersNames)

	gs := GameStatus{
		CurrentPlayer:       dto.Viewer.Name,
		NextTurnPlayer:      dto.NextTurnPlayer,
		TurnPlayer:          dto.TurnPlayer,
		CurrentAction:       string(dto.CurrentAction),
		LastAction:          dto.LastAction,
		GameMode:            dto.GameMode,
		NewCards:            []string{},
		CurrentPlayerHand:   []HandCard{},
		CurrentPlayerField:  []FieldCard{},
		CurrentPlayerCastle: NewCastle(dto.Viewer.Castle),
		IsEliminated:        dto.IsEliminated,
		IsDisconnected:      dto.IsDisconnected,
		CanTrade:            dto.CanTrade,
		Cemetery:            NewCemetery(dto.CemeteryCount, dto.CemeteryLastDead),
		DiscardPile:         NewDiscardPile(dto.DiscardPileCount, dto.DiscardPileLastCard),
		CardsInDeck:         dto.DeckCount,
		History:             []HistoryLine{},
		PlayersOrder:        playersOrder,
		GameStartedAt:       dto.GameStartedAt,
		TurnStartedAt:       dto.TurnStartedAt,
		TurnTimeLimitSecs:   120,
	}

	for _, line := range dto.History {
		gs.History = append(gs.History, NewHistoryLine(
			line.Msg, line.Category))
	}

	if len(dto.NewCards) > 0 {
		for _, c := range dto.NewCards {
			gs.NewCards = append(gs.NewCards, c.GetID())
		}
	}
	if len(dto.ModalCards) > 0 {
		for _, c := range dto.ModalCards {
			gs.ModalCards = append(gs.ModalCards, fromDomainCard(c))
		}
	}

	// Include last moved warrior ID for animation (only on the move action itself)
	if dto.LastAction == types.LastActionMoveWarrior && dto.LastMovedWarriorID != "" {
		gs.LastMovedWarriorID = dto.LastMovedWarriorID
	}

	// Include attack animation info (only on the attack action itself)
	if (dto.LastAction == types.LastActionAttack || dto.LastAction == types.LastActionHarpoon) && dto.LastAttackWeaponID != "" {
		gs.LastAttackWeaponID = dto.LastAttackWeaponID
		gs.LastAttackTargetID = dto.LastAttackTargetID
		gs.LastAttackTargetPlayer = dto.LastAttackTargetPlayer
	}

	// Include blood rain animation info (target player only, AoE attack)
	if dto.LastAction == types.LastActionBloodRain && dto.LastAttackTargetPlayer != "" {
		gs.LastAttackTargetPlayer = dto.LastAttackTargetPlayer
	}

	// Include stolen card info for the victim (only on the steal action itself)
	if dto.LastAction == types.LastActionSteal && dto.StolenFrom != "" &&
		dto.StolenCard != nil && dto.Viewer.Name == dto.StolenFrom {
		gs.StolenFromYouCard = fromDomainCards([]cards.Card{dto.StolenCard})
	}

	// Include sabotaged card info for the victim (only on the sabotage action itself)
	if dto.LastAction == types.LastActionSabotage && dto.SabotagedFrom != "" &&
		dto.SabotagedCard != nil && dto.Viewer.Name == dto.SabotagedFrom {
		gs.SabotagedFromYouCard = fromDomainCards([]cards.Card{dto.SabotagedCard})
	}

	// Include spy notification for all players except the spy
	if dto.SpyTarget != "" && dto.LastAction == types.LastActionSpy &&
		dto.Viewer.Name != dto.CurrentPlayerName {
		spyPlayer := dto.CurrentPlayerName
		if dto.SpyTarget == types.SpyTargetDeck {
			gs.SpyNotification = spyPlayer + " spied on the deck"
		} else {
			gs.SpyNotification = spyPlayer + " spied on " + dto.SpyTargetPlayer + "'s hand"
		}
	}

	// Include ambush trigger notification for the attacker only
	if dto.LastAction == types.LastActionAmbush && dto.AmbushAttackerName != "" &&
		dto.Viewer.Name == dto.AmbushAttackerName {
		gs.AmbushTriggered = &AmbushTrigger{
			Effect:        dto.AmbushEffect,
			EffectDisplay: dto.AmbushEffect.DisplayName(),
		}
	}

	processHandCards(dto.Viewer, dto, &gs)

	for _, warrior := range dto.Viewer.Field.Warriors {
		gs.CurrentPlayerField = append(gs.CurrentPlayerField, NewFieldCard(warrior))
	}
	gs.CurrentPlayerAmbushInField = dto.Viewer.Field.HasAmbush

	processOpponents(dto, &gs)

	if dto.IsGameOver {
		gs.GameOverMgs = "Game over! The winner is " + dto.Winner
		gs.IsWinner = dto.IsPlayerWinner
	}

	return gs
}

func processHandCards(viewer ViewerInput, game GameStatusDTO, gs *GameStatus) {
	action := game.CurrentAction
	canMove := game.CanMoveWarrior

	for _, card := range viewer.Hand {
		switch ct := card.(type) {
		case cards.Warrior:
			gs.CanMoveWarrior = canMove
			gs.CurrentPlayerHand = append(gs.CurrentPlayerHand, NewWarriorHandCard(ct))

		case cards.Weapon:
			if builder, ok := specialWeaponHandCardBuilders[ct.Type()]; ok {
				gs.CurrentPlayerHand = append(gs.CurrentPlayerHand,
					builder(ct.GetID(), viewer, game, action))
			} else {
				gs.CurrentPlayerHand = append(gs.CurrentPlayerHand,
					NewWeaponHandCard(ct, viewer.Field,
						game.EnemyFields, viewer.Castle.IsConstructed, action))
			}

		case cards.Catapult:
			gs.CurrentPlayerHand = append(gs.CurrentPlayerHand,
				NewCatapultHandCard(ct.GetID(), game.AnyEnemyCastleAttackable,
					action))

		case cards.Spy:
			gs.CurrentPlayerHand = append(gs.CurrentPlayerHand,
				NewSpyHandCard(ct.GetID(), action))

		case cards.Thief:
			gs.CurrentPlayerHand = append(gs.CurrentPlayerHand,
				NewThiefHandCard(ct.GetID(), action))

		case cards.Sabotage:
			gs.CurrentPlayerHand = append(gs.CurrentPlayerHand,
				NewSabotageHandCard(ct.GetID(), game.AnyEnemyHasCards, action))

		case cards.Fortress:
			gs.CurrentPlayerHand = append(gs.CurrentPlayerHand,
				NewFortressHandCard(ct.GetID(), viewer.Castle.IsConstructed,
					game.AllyHasCastleConstructed, action))

		case cards.Resurrection:
			gs.CurrentPlayerHand = append(gs.CurrentPlayerHand,
				NewResurrectionHandCard(ct.GetID(), game.CemeteryCount, action))

		case cards.Ambush:
			gs.CurrentPlayerHand = append(gs.CurrentPlayerHand,
				NewAmbushHandCard(ct.GetID(), viewer.Field.HasAmbush, action))

		case cards.Resource:
			gs.CurrentPlayerHand = append(gs.CurrentPlayerHand,
				NewResourceHandCard(ct, viewer.Castle.IsConstructed,
					game.AllyHasCastleConstructed, viewer.CanBuyWith(ct), action))
		}
	}
}

func processOpponents(game GameStatusDTO, gs *GameStatus) {
	for _, opp := range game.Opponents {
		o := OpponentStatus{
			PlayerName:     opp.Name,
			CardsInHand:    opp.CardsInHand,
			Castle:         NewCastle(opp.Castle),
			IsAlly:         opp.IsAlly,
			IsEliminated:   opp.IsEliminated,
			IsDisconnected: opp.IsDisconnected,
			AmbushInField:  opp.Field.HasAmbush,
		}
		for _, warrior := range opp.Field.Warriors {
			o.Field = append(o.Field, NewFieldCard(warrior))
		}
		gs.Opponents = append(gs.Opponents, o)
	}
}
