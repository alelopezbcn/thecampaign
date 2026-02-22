package gamestatus

import (
	"time"

	"github.com/alelopezbcn/thecampaign/internal/domain/board"
	"github.com/alelopezbcn/thecampaign/internal/domain/cards"
	"github.com/alelopezbcn/thecampaign/internal/domain/types"
)

type GameStatus struct {
	CurrentPlayer          string               `json:"current_player"`
	TurnPlayer             string               `json:"turn_player"`
	CurrentAction          string               `json:"current_action"`
	LastAction             types.LastActionType `json:"last_action,omitempty"`
	NewCards               []string             `json:"new_cards"`
	CanMoveWarrior         bool                 `json:"can_move_warrior"`
	CanTrade               bool                 `json:"can_trade"`
	CurrentPlayerHand      []HandCard           `json:"current_player_hand"`
	CurrentPlayerField     []FieldCard          `json:"current_player_field"`
	CurrentPlayerCastle    Castle               `json:"current_player_castle"`
	IsEliminated           bool                 `json:"is_eliminated"`
	IsDisconnected         bool                 `json:"is_disconnected"`
	Opponents              []OpponentStatus     `json:"opponents"`
	GameMode               string               `json:"game_mode"`
	Cemetery               Cemetery             `json:"cemetery"`
	DiscardPile            DiscardPile          `json:"discard_pile"`
	CardsInDeck            int                  `json:"deck"`
	ModalCards             []Card               `json:"modal_cards"`
	LastMovedWarriorID     string               `json:"last_moved_warrior_id,omitempty"`
	LastAttackWeaponID     string               `json:"last_attack_weapon_id,omitempty"`
	LastAttackTargetID     string               `json:"last_attack_target_id,omitempty"`
	LastAttackTargetPlayer string               `json:"last_attack_target_player,omitempty"`
	StolenFromYouCard      []Card               `json:"stolen_from_you_card,omitempty"`
	SpyNotification        string               `json:"spy_notification,omitempty"`
	History                []HistoryLine        `json:"history"`
	PlayersOrder           []string             `json:"players_order"`
	NextTurnPlayer         string               `json:"next_turn_player,omitempty"`
	GameOverMgs            string               `json:"game_over_msg"`
	IsWinner               bool                 `json:"is_winner"`
	GameStartedAt          time.Time            `json:"game_started_at"`
	TurnStartedAt          time.Time            `json:"turn_started_at"`
	TurnTimeLimitSecs      int                  `json:"turn_time_limit_secs"`
}

type OpponentStatus struct {
	PlayerName     string
	Field          []FieldCard
	Castle         Castle
	CardsInHand    int
	IsAlly         bool
	IsEliminated   bool
	IsDisconnected bool
}

func NewGameStatus(dto GameStatusDTO) GameStatus {
	playersOrder := make([]string, len(dto.Players))
	for i, p := range dto.Players {
		playersOrder[i] = p.Name()
	}

	gs := GameStatus{
		CurrentPlayer:       dto.Viewer.Name(),
		NextTurnPlayer:      dto.NextTurnPlayer,
		TurnPlayer:          dto.TurnPlayer,
		CurrentAction:       string(dto.CurrentAction),
		LastAction:          dto.LastAction,
		GameMode:            dto.GameMode,
		NewCards:            []string{},
		CurrentPlayerHand:   []HandCard{},
		CurrentPlayerField:  []FieldCard{},
		CurrentPlayerCastle: NewCastle(dto.Viewer.Castle()),
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
	if dto.LastAction == types.LastActionAttack && dto.LastAttackWeaponID != "" {
		gs.LastAttackWeaponID = dto.LastAttackWeaponID
		gs.LastAttackTargetID = dto.LastAttackTargetID
		gs.LastAttackTargetPlayer = dto.LastAttackTargetPlayer
	}

	// Include stolen card info for the victim (only on the steal action itself)
	if dto.LastAction == types.LastActionSteal && dto.StolenFrom != "" &&
		dto.StolenCard != nil && dto.Viewer.Name() == dto.StolenFrom {
		gs.StolenFromYouCard = fromDomainCards([]cards.Card{dto.StolenCard})
	}

	// Include spy notification for all players except the spy
	if dto.SpyTarget != "" && dto.LastAction == types.LastActionSpy &&
		dto.Viewer.Name() != dto.CurrentPlayerName {
		spyPlayer := dto.CurrentPlayerName
		if dto.SpyTarget == types.SpyTargetDeck {
			gs.SpyNotification = spyPlayer + " spied on the deck"
		} else {
			gs.SpyNotification = spyPlayer + " spied on " + dto.SpyTargetPlayer + "'s hand"
		}
	}

	processHandCards(dto.Viewer, dto, &gs)

	for _, warrior := range dto.Viewer.Field().Warriors() {
		gs.CurrentPlayerField = append(gs.CurrentPlayerField, NewFieldCard(warrior))
	}

	processOpponents(dto.Viewer, dto, &gs)

	if dto.IsGameOver {
		gs.GameOverMgs = "Game over! The winner is " + dto.Winner
		gs.IsWinner = dto.IsPlayerWinner
	}

	return gs
}

func processHandCards(viewer board.Player, game GameStatusDTO, gs *GameStatus) {
	action := game.CurrentAction
	canMove := game.CanMoveWarrior

	for _, card := range viewer.Hand().ShowCards() {
		switch ct := card.(type) {
		case cards.Warrior:
			gs.CanMoveWarrior = canMove
			gs.CurrentPlayerHand = append(gs.CurrentPlayerHand, NewWarriorHandCard(ct))

		case cards.Weapon:
			var enemyFields []board.Field
			for _, enemy := range game.EnemiesFn(viewer.Idx()) {
				enemyFields = append(enemyFields, enemy.Field())
			}

			switch ct.Type() {
			case types.SpecialPowerWeaponType:
				var allyFields []board.Field
				for _, ally := range game.AlliesFn(viewer.Idx()) {
					allyFields = append(allyFields, ally.Field())
				}

				gs.CurrentPlayerHand = append(gs.CurrentPlayerHand,
					NewSpecialPowerHandCard(ct.GetID(), viewer.Field(),
						allyFields, enemyFields, action))

				continue
			case types.HarpoonWeaponType:
				gs.CurrentPlayerHand = append(gs.CurrentPlayerHand,
					NewHarpoonHandCard(ct.GetID(), enemyFields, action))

				continue
			case types.BloodRainWeaponType:
				gs.CurrentPlayerHand = append(gs.CurrentPlayerHand,
					NewBloodRainHandCard(ct.GetID(), enemyFields,
						action))
				continue
			default:
				gs.CurrentPlayerHand = append(gs.CurrentPlayerHand,
					NewWeaponHandCard(ct, viewer.Field(),
						enemyFields, viewer.Castle().IsConstructed(), action))
			}

		case cards.Catapult:
			canBeAttacked := false
			for _, enemy := range game.EnemiesFn(viewer.Idx()) {
				if enemy.Castle().CanBeAttacked() {
					canBeAttacked = true
					break
				}
			}
			gs.CurrentPlayerHand = append(gs.CurrentPlayerHand,
				NewCatapultHandCard(ct.GetID(), canBeAttacked,
					action))

		case cards.Spy:
			gs.CurrentPlayerHand = append(gs.CurrentPlayerHand,
				NewSpyHandCard(ct.GetID(), action))

		case cards.Thief:
			gs.CurrentPlayerHand = append(gs.CurrentPlayerHand,
				NewThiefHandCard(ct.GetID(), action))

		case cards.Resource:
			allyCastleConstructed := false
			for _, ally := range game.AlliesFn(game.PlayerIndex) {
				if ally.Castle().IsConstructed() {
					allyCastleConstructed = true
					break
				}
			}
			gs.CurrentPlayerHand = append(gs.CurrentPlayerHand,
				NewResourceHandCard(ct, viewer.Castle().IsConstructed(),
					allyCastleConstructed, viewer.CanBuyWith(ct), action))
		}
	}
}

func processOpponents(viewer board.Player, game GameStatusDTO, gs *GameStatus) {
	viewerIdx := game.PlayerIndex

	for i, p := range game.Players {
		if i == viewerIdx {
			continue
		}
		opp := OpponentStatus{
			PlayerName:     p.Name(),
			CardsInHand:    p.CardsInHand(),
			Castle:         NewCastle(p.Castle()),
			IsAlly:         game.SameTeamFn(viewerIdx, i),
			IsEliminated:   game.EliminatedPlayers[i],
			IsDisconnected: game.DisconnectedPlayers[i],
		}
		for _, warrior := range p.Field().Warriors() {
			opp.Field = append(opp.Field, NewFieldCard(warrior))
		}
		gs.Opponents = append(gs.Opponents, opp)
	}
}
