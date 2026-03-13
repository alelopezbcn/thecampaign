package gamestatus

import (
	"github.com/alelopezbcn/thecampaign/internal/domain/cards"
	"github.com/alelopezbcn/thecampaign/internal/domain/gameevents"
	"github.com/alelopezbcn/thecampaign/internal/domain/types"
)

const turnTimeLimitSecs = 120

const timeFormat = "2006-01-02T15:04:05Z"

type AmbushTrigger struct {
	Effect              types.AmbushEffect `json:"effect"`
	EffectDisplay       string             `json:"effect_display"`
	AttackerName        string             `json:"attacker_name"`
	DefenderName        string             `json:"defender_name"`
	AttackerWarriorType string             `json:"attacker_warrior_type,omitempty"`
	AttackerHPBefore    int                `json:"attacker_hp_before,omitempty"`
	AttackerHPAfter     int                `json:"attacker_hp_after,omitempty"`
	AttackerDied        bool               `json:"attacker_died,omitempty"`
	TargetWarriorType   string             `json:"target_warrior_type,omitempty"`
	TargetHPBefore      int                `json:"target_hp_before,omitempty"`
	TargetHPAfter       int                `json:"target_hp_after,omitempty"`
	WeaponType          string             `json:"weapon_type,omitempty"`
	DamageAmount        int                `json:"damage_amount,omitempty"`
}

type TreasonNotification struct {
	WarriorCard Card   `json:"warrior_card"`
	StolenBy    string `json:"stolen_by"`
}

type ChampionsBountyNotification struct {
	EarnedBy string `json:"earned_by"`
	Cards    int    `json:"cards"`
}

type PlayerStat struct {
	Name        string `json:"name"`
	Kills       int    `json:"kills"`
	Damage      int    `json:"damage"`
	CastleValue int    `json:"castle_value"`
	IsWinner    bool   `json:"is_winner"`
	IsMVP       bool   `json:"is_mvp"`
}

type ResurrectionNotification struct {
	WarriorCard  Card   `json:"warrior_card"`
	TargetPlayer string `json:"target_player"`
	PlayerName   string `json:"player_name"`
}

type CatapultNotification struct {
	AttackerName string `json:"attacker_name"`
	TargetPlayer string `json:"target_player"`
	GoldStolen   int    `json:"gold_stolen"`
	Blocked      bool   `json:"blocked"`
}

type GameStatus struct {
	CurrentPlayer                string                       `json:"current_player"`
	TurnPlayer                   string                       `json:"turn_player"`
	CurrentAction                string                       `json:"current_action"`
	LastAction                   types.LastActionType         `json:"last_action,omitempty"`
	NewCards                     []string                     `json:"new_cards"`
	CanMoveWarrior               bool                         `json:"can_move_warrior"`
	CanTrade                     bool                         `json:"can_trade"`
	CanForge                     bool                         `json:"can_forge"`
	CurrentPlayerHand            []HandCard                   `json:"current_player_hand"`
	CurrentPlayerField           []FieldCard                  `json:"current_player_field"`
	CurrentPlayerFieldHP         int                          `json:"current_player_field_hp"`
	CurrentPlayerCastle          Castle                       `json:"current_player_castle"`
	CurrentPlayerAmbushInField   bool                         `json:"current_player_ambush_in_field"`
	IsEliminated                 bool                         `json:"is_eliminated"`
	IsDisconnected               bool                         `json:"is_disconnected"`
	Opponents                    []OpponentStatus             `json:"opponents"`
	GameMode                     string                       `json:"game_mode"`
	Cemetery                     Cemetery                     `json:"cemetery"`
	DiscardPile                  DiscardPile                  `json:"discard_pile"`
	CardsInDeck                  int                          `json:"cards_in_deck"`
	ModalCards                   []Card                       `json:"modal_cards,omitempty"`
	LastMovedWarriorID           string                       `json:"last_moved_warrior_id,omitempty"`
	LastAttackWeaponID           string                       `json:"last_attack_weapon_id,omitempty"`
	LastAttackTargetID           string                       `json:"last_attack_target_id,omitempty"`
	LastAttackTargetPlayer       string                       `json:"last_attack_target_player,omitempty"`
	StolenFromYouCard            []Card                       `json:"stolen_from_you_card,omitempty"`
	SabotagedFromYouCard         []Card                       `json:"sabotaged_from_you_card,omitempty"`
	SpyNotification              string                       `json:"spy_notification,omitempty"`
	AmbushTriggered              *AmbushTrigger               `json:"ambush_triggered,omitempty"`
	AmbushPlacedOn               string                       `json:"ambush_placed_on,omitempty"`
	TreasonNotification          *TreasonNotification         `json:"treason_notification,omitempty"`
	ChampionsBounty              *ChampionsBountyNotification `json:"champions_bounty,omitempty"`
	ResurrectionNotification     *ResurrectionNotification    `json:"resurrection_notification,omitempty"`
	CatapultNotification         *CatapultNotification        `json:"catapult_notification,omitempty"`
	History                      []HistoryLine                `json:"history"`
	PlayersOrder                 []string                     `json:"players_order"`
	NextTurnPlayer               string                       `json:"next_turn_player,omitempty"`
	GameOverMsg                  string                       `json:"game_over_msg,omitempty"`
	IsWinner                     bool                         `json:"is_winner"`
	PlayerStats                  []PlayerStat                 `json:"player_stats,omitempty"`
	GameStartedAt                string                       `json:"game_started_at"`
	TurnStartedAt                string                       `json:"turn_started_at"`
	TurnTimeLimitSecs            int                          `json:"turn_time_limit_secs"`
	CurrentEvent                 string                       `json:"current_event"`
	CurrentEventDisplay          string                       `json:"current_event_display"`
	CurrentEventDescription      string                       `json:"current_event_description"`
	CurrentEventWeaponModifier   int                          `json:"current_event_weapon_modifier,omitempty"`
	CurrentEventExcludedWeapon   string                       `json:"current_event_excluded_weapon,omitempty"`
	CurrentEventResourceModifier int                          `json:"current_event_resource_modifier,omitempty"`
}

type OpponentStatus struct {
	PlayerName     string      `json:"player_name"`
	Field          []FieldCard `json:"field"`
	FieldHP        int         `json:"field_hp"`
	Castle         Castle      `json:"castle"`
	CardsInHand    int         `json:"cards_in_hand"`
	IsAlly         bool        `json:"is_ally"`
	IsEliminated   bool        `json:"is_eliminated"`
	IsDisconnected bool        `json:"is_disconnected"`
	AmbushInField  bool        `json:"ambush_in_field"`
}

func NewGameStatus(in BuildInput) GameStatus {
	playersOrder := make([]string, len(in.PlayersNames))
	copy(playersOrder, in.PlayersNames)

	gs := GameStatus{
		CurrentPlayer:       in.Viewer.Name,
		NextTurnPlayer:      in.NextTurnPlayer,
		TurnPlayer:          in.TurnPlayer,
		CurrentAction:       string(in.CurrentAction),
		LastAction:          in.LastAction,
		GameMode:            in.GameMode,
		NewCards:            []string{},
		CurrentPlayerHand:   []HandCard{},
		CurrentPlayerField:  []FieldCard{},
		CurrentPlayerCastle: NewCastle(in.Viewer.Castle),
		IsEliminated:        in.IsEliminated,
		IsDisconnected:      in.IsDisconnected,
		CanTrade:            in.CanTrade,
		CanForge:            in.CanForge,
		Cemetery:            NewCemetery(in.CemeteryCount, in.CemeteryLastDead),
		DiscardPile:         NewDiscardPile(in.DiscardPileCount, in.DiscardPileLastCard),
		CardsInDeck:         in.DeckCount,
		History:             []HistoryLine{},
		PlayersOrder:        playersOrder,
		GameStartedAt:       in.GameStartedAt.UTC().Format(timeFormat),
		TurnStartedAt:       in.TurnStartedAt.UTC().Format(timeFormat),
		TurnTimeLimitSecs:   turnTimeLimitSecs,
	}

	for _, line := range in.History {
		gs.History = append(gs.History, NewHistoryLine(line.Msg, line.Category))
	}

	for _, c := range in.NewCards {
		gs.NewCards = append(gs.NewCards, c.GetID())
	}
	for _, c := range in.ModalCards {
		gs.ModalCards = append(gs.ModalCards, fromDomainCard(c))
	}

	applyEventInfo(in, &gs)
	applyMoveWarriorAnimation(in, &gs)
	applyAttackAnimation(in, &gs)
	applyBloodRainAnimation(in, &gs)
	applyStealNotification(in, &gs)
	applySabotageNotification(in, &gs)
	applySpyNotification(in, &gs)
	applyPlaceAmbushAnimation(in, &gs)
	applyAmbushNotification(in, &gs)
	applyTreasonNotification(in, &gs)
	applyChampionsBountyNotification(in, &gs)
	applyResurrectionNotification(in, &gs)
	applyCatapultNotification(in, &gs)

	processHandCards(in.Viewer, in, &gs)

	for _, warrior := range in.Viewer.Field.Warriors {
		gs.CurrentPlayerField = append(gs.CurrentPlayerField, NewFieldCard(warrior))
		gs.CurrentPlayerFieldHP += warrior.Health()
	}
	gs.CurrentPlayerAmbushInField = in.Viewer.Field.HasAmbush

	processOpponents(in, &gs)

	if in.IsGameOver {
		gs.GameOverMsg = "Game over! The winner is " + in.Winner
		gs.IsWinner = in.IsPlayerWinner
		for _, s := range in.PlayerStats {
			gs.PlayerStats = append(gs.PlayerStats, PlayerStat{
				Name:        s.Name,
				Kills:       s.Kills,
				Damage:      s.Damage,
				CastleValue: s.CastleValue,
				IsWinner:    s.IsWinner,
				IsMVP:       s.IsMVP,
			})
		}
	}

	return gs
}

func applyEventInfo(in BuildInput, gs *GameStatus) {
	handler := gameevents.NewHandler(in.CurrentEvent)
	name, desc := handler.Display()
	gs.CurrentEvent = string(in.CurrentEvent.Type)
	gs.CurrentEventDisplay = name
	gs.CurrentEventDescription = desc
	if in.CurrentEvent.Type == types.EventTypeCurse {
		gs.CurrentEventWeaponModifier = in.CurrentEvent.CurseModifier
		gs.CurrentEventExcludedWeapon = string(in.CurrentEvent.CurseExcludedWeapon)
	}
	if in.CurrentEvent.Type == types.EventTypeHarvest {
		gs.CurrentEventResourceModifier = in.CurrentEvent.HarvestModifier
	}
}

func applyMoveWarriorAnimation(in BuildInput, gs *GameStatus) {
	if in.LastAction == types.LastActionMoveWarrior && in.LastMovedWarriorID != "" {
		gs.LastMovedWarriorID = in.LastMovedWarriorID
	}
}

func applyAttackAnimation(in BuildInput, gs *GameStatus) {
	if (in.LastAction == types.LastActionAttack || in.LastAction == types.LastActionHarpoon) && in.LastAttackWeaponID != "" {
		gs.LastAttackWeaponID = in.LastAttackWeaponID
		gs.LastAttackTargetID = in.LastAttackTargetID
		gs.LastAttackTargetPlayer = in.LastAttackTargetPlayer
	}
}

func applyBloodRainAnimation(in BuildInput, gs *GameStatus) {
	if in.LastAction == types.LastActionBloodRain && in.LastAttackTargetPlayer != "" {
		gs.LastAttackTargetPlayer = in.LastAttackTargetPlayer
	}
}

func applyStealNotification(in BuildInput, gs *GameStatus) {
	if in.LastAction == types.LastActionSteal && in.StolenFrom != "" &&
		in.StolenCard != nil && in.Viewer.Name == in.StolenFrom {
		gs.StolenFromYouCard = fromDomainCards([]cards.Card{in.StolenCard})
	}
}

func applySabotageNotification(in BuildInput, gs *GameStatus) {
	if in.LastAction == types.LastActionSabotage && in.SabotagedFrom != "" &&
		in.SabotagedCard != nil && in.Viewer.Name == in.SabotagedFrom {
		gs.SabotagedFromYouCard = fromDomainCards([]cards.Card{in.SabotagedCard})
	}
}

func applySpyNotification(in BuildInput, gs *GameStatus) {
	if in.SpyTarget == "" || in.LastAction != types.LastActionSpy || in.Viewer.Name == in.CurrentPlayerName {
		return
	}
	if in.SpyTarget == types.SpyTargetDeck {
		gs.SpyNotification = in.CurrentPlayerName + " spied on the deck"
	} else {
		gs.SpyNotification = in.CurrentPlayerName + " spied on " + in.SpyTargetPlayer + "'s hand"
	}
}

func applyPlaceAmbushAnimation(in BuildInput, gs *GameStatus) {
	if in.LastAction == types.LastActionPlaceAmbush && in.AmbushPlacedOn != "" {
		gs.AmbushPlacedOn = in.AmbushPlacedOn
	}
}

func applyAmbushNotification(in BuildInput, gs *GameStatus) {
	if in.LastAction == types.LastActionAmbush && in.AmbushAttackerName != "" {
		gs.AmbushTriggered = &AmbushTrigger{
			Effect:              in.AmbushEffect,
			EffectDisplay:       in.AmbushEffect.DisplayName(),
			AttackerName:        in.AmbushAttackerName,
			DefenderName:        in.LastAttackTargetPlayer,
			AttackerWarriorType: in.AmbushAttackerWarriorType,
			AttackerHPBefore:    in.AmbushAttackerHPBefore,
			AttackerHPAfter:     in.AmbushAttackerHPAfter,
			AttackerDied:        in.AmbushAttackerDied,
			TargetWarriorType:   in.AmbushTargetWarriorType,
			TargetHPBefore:      in.AmbushTargetHPBefore,
			TargetHPAfter:       in.AmbushTargetHPAfter,
			WeaponType:          in.AmbushWeaponType,
			DamageAmount:        in.AmbushDamageAmount,
		}
	}
}

func applyChampionsBountyNotification(in BuildInput, gs *GameStatus) {
	if in.ChampionsBountyEarner != "" && in.ChampionsBountyCards > 0 {
		gs.ChampionsBounty = &ChampionsBountyNotification{
			EarnedBy: in.ChampionsBountyEarner,
			Cards:    in.ChampionsBountyCards,
		}
	}
}

func applyResurrectionNotification(in BuildInput, gs *GameStatus) {
	if in.LastAction == types.LastActionResurrection && in.ResurrectionWarrior != nil {
		gs.ResurrectionNotification = &ResurrectionNotification{
			WarriorCard:  fromDomainCard(in.ResurrectionWarrior),
			TargetPlayer: in.ResurrectionTargetPlayer,
			PlayerName:   in.ResurrectionPlayerName,
		}
	}
}

func applyTreasonNotification(in BuildInput, gs *GameStatus) {
	if in.LastAction == types.LastActionTreason && in.TraitorFromPlayer != "" &&
		in.Viewer.Name == in.TraitorFromPlayer && in.TraitorWarrior != nil {
		gs.TreasonNotification = &TreasonNotification{
			WarriorCard: fromDomainCard(in.TraitorWarrior),
			StolenBy:    in.CurrentPlayerName,
		}
	}
}

// applyCatapultNotification sends the catapult result to every player except the attacker.
// The target player receives it so they can show a detailed modal; others see a toast.
func applyCatapultNotification(in BuildInput, gs *GameStatus) {
	if in.CatapultAttacker == "" {
		return
	}
	// Attacker already knows the outcome — skip for them.
	if in.Viewer.Name == in.CatapultAttacker {
		return
	}
	gs.CatapultNotification = &CatapultNotification{
		AttackerName: in.CatapultAttacker,
		TargetPlayer: in.CatapultTarget,
		GoldStolen:   in.CatapultGoldStolen,
		Blocked:      in.CatapultBlocked,
	}
}

func processHandCards(viewer ViewerInput, game BuildInput, gs *GameStatus) {
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

		case cards.Treason:
			gs.CurrentPlayerHand = append(gs.CurrentPlayerHand,
				NewTreasonHandCard(ct.GetID(), game.AnyEnemyHasWeakWarriors, action))

		case cards.Fortress:
			gs.CurrentPlayerHand = append(gs.CurrentPlayerHand,
				NewFortressHandCard(ct.GetID(), viewer.Castle.IsConstructed,
					game.AllyHasCastleConstructed, action))

		case cards.Resurrection:
			gs.CurrentPlayerHand = append(gs.CurrentPlayerHand,
				NewResurrectionHandCard(ct.GetID(), game.CemeteryCount, action))

		case cards.Ambush:
			canBePlaced := !viewer.Field.HasAmbush
			for _, f := range game.AllyFields {
				if !f.HasAmbush {
					canBePlaced = true
				}
			}
			gs.CurrentPlayerHand = append(gs.CurrentPlayerHand,
				NewAmbushHandCard(ct.GetID(), canBePlaced, action))

		case cards.Resource:
			gs.CurrentPlayerHand = append(gs.CurrentPlayerHand,
				NewResourceHandCard(ct, viewer.Castle.IsConstructed,
					game.AllyHasCastleConstructed, viewer.CanBuyWith(ct), action))
		}
	}
}

func processOpponents(game BuildInput, gs *GameStatus) {
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
			o.FieldHP += warrior.Health()
		}
		gs.Opponents = append(gs.Opponents, o)
	}
}
