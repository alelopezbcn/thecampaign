package domain

import (
	"time"

	"github.com/alelopezbcn/thecampaign/internal/domain/gamestatus"
	"github.com/alelopezbcn/thecampaign/internal/domain/ports"
	"github.com/alelopezbcn/thecampaign/internal/domain/types"
)

type GameStatus struct {
	CurrentPlayer  string               `json:"current_player"`
	TurnPlayer     string               `json:"turn_player"`
	CurrentAction  string               `json:"current_action"`
	LastAction     types.LastActionType `json:"last_action,omitempty"`
	NewCards       []string             `json:"new_cards"`
	CanMoveWarrior bool                 `json:"can_move_warrior"`
	CanTrade       bool                 `json:"can_trade"`

	CurrentPlayerHand      []gamestatus.HandCard    `json:"current_player_hand"`
	CurrentPlayerField     []gamestatus.FieldCard   `json:"current_player_field"`
	CurrentPlayerCastle    gamestatus.Castle        `json:"current_player_castle"`
	IsEliminated           bool                     `json:"is_eliminated"`
	IsDisconnected         bool                     `json:"is_disconnected"`
	Opponents              []OpponentStatus         `json:"opponents"`
	GameMode               string                   `json:"game_mode"`
	Cemetery               gamestatus.Cemetery      `json:"cemetery"`
	DiscardPile            gamestatus.DiscardPile   `json:"discard_pile"`
	CardsInDeck            int                      `json:"deck"`
	ModalCards             []gamestatus.Card        `json:"modal_cards"`
	LastMovedWarriorID     string                   `json:"last_moved_warrior_id,omitempty"`
	LastAttackWeaponID     string                   `json:"last_attack_weapon_id,omitempty"`
	LastAttackTargetID     string                   `json:"last_attack_target_id,omitempty"`
	LastAttackTargetPlayer string                   `json:"last_attack_target_player,omitempty"`
	StolenFromYouCard      []gamestatus.Card        `json:"stolen_from_you_card,omitempty"`
	SpyNotification        string                   `json:"spy_notification,omitempty"`
	History                []gamestatus.HistoryLine `json:"history"`
	PlayersOrder           []string                 `json:"players_order"`
	NextTurnPlayer         string                   `json:"next_turn_player,omitempty"`
	GameOverMgs            string                   `json:"game_over_msg"`
	IsWinner               bool                     `json:"is_winner"`
	GameStartedAt          time.Time                `json:"game_started_at"`
	TurnStartedAt          time.Time                `json:"turn_started_at"`
	TurnTimeLimitSecs      int                      `json:"turn_time_limit_secs"`
}

type OpponentStatus struct {
	PlayerName     string
	Field          []gamestatus.FieldCard
	Castle         gamestatus.Castle
	CardsInHand    int
	IsAlly         bool
	IsEliminated   bool
	IsDisconnected bool
}

func newGameStatusWithModalCards(viewer ports.Player, game *Game,
	modalCards []ports.Card,
) GameStatus {
	gs := newGameStatus(viewer, game)

	gs.ModalCards = gamestatus.FromDomainCards(modalCards)

	return gs
}

type GameStatusDTO struct {
	Viewer                 ports.Player
	NewCards               []ports.Card
	ModalCards             []ports.Card
	PlayerIndex            int
	PlayersNames           []string
	Players                []ports.Player
	ViewerName             string
	NextTurnPlayer         string
	TurnPlayer             string
	CurrentAction          string
	LastAction             types.LastActionType
	GameMode               string
	CastleIsConstructed    bool
	CastleResourceCards    int
	CastleValue            int
	IsEliminated           bool
	IsDisconnected         bool
	CanTrade               bool
	CemeteryCount          int
	CemeteryGetLast        ports.Warrior
	DiscardPileCount       int
	DiscardPileLastCard    ports.Card
	DeckCount              int
	GameStartedAt          time.Time
	TurnStartedAt          time.Time
	History                types.HistoryLine
	LastMovedWarriorID     string
	LastAttackWeaponID     string
	LastAttackTargetID     string
	LastAttackTargetPlayer string
	StolenFrom             string
	StolenCard             ports.Card
	SpyTarget              string
	SpyTargetPlayer        string
	CurrentPlayerName      string
	IsGameOver             bool
	Winner                 string
	IsPlayerWinner         bool
	SameTeamFn             func(i, j int) bool
	EliminatedPlayers      map[int]bool
	DisconnectedPlayers    map[int]bool
	CanMoveWarrior         bool
	EnemiesFn              func(playerIdx int) []ports.Player
	AlliesFn               func(playerIdx int) []ports.Player
}

func newGameStatus(viewer ports.Player, game *Game, newCards ...ports.Card,
) GameStatus {
	viewerIdx := game.PlayerIndex(viewer.Name())

	playersOrder := make([]string, len(game.Players))
	for i, p := range game.Players {
		playersOrder[i] = p.Name()
	}

	gs := GameStatus{
		CurrentPlayer:       viewer.Name(),
		NextTurnPlayer:      game.nextActiveTurnPlayer(),
		TurnPlayer:          game.CurrentPlayer().Name(),
		CurrentAction:       string(game.currentAction),
		LastAction:          game.lastResult.Action,
		GameMode:            string(game.Mode),
		NewCards:            []string{},
		CurrentPlayerHand:   []gamestatus.HandCard{},
		CurrentPlayerField:  []gamestatus.FieldCard{},
		CurrentPlayerCastle: gamestatus.NewCastle(viewer.Castle()),
		IsEliminated:        game.EliminatedPlayers[viewerIdx],
		IsDisconnected:      game.DisconnectedPlayers[viewerIdx],
		CanTrade:            game.turnState.CanTrade,
		Cemetery:            gamestatus.NewCemetery(game.cemetery),
		DiscardPile:         gamestatus.NewDiscardPile(game.discardPile),
		CardsInDeck:         game.deck.Count(),
		History:             []gamestatus.HistoryLine{},
		PlayersOrder:        playersOrder,
		GameStartedAt:       game.GameStartedAt,
		TurnStartedAt:       game.turnState.StartedAt,
		TurnTimeLimitSecs:   120,
	}

	for _, line := range game.GetHistory() {
		gs.History = append(gs.History, gamestatus.NewHistoryLine(
			line.Msg, line.Category))
	}

	if len(newCards) > 0 {
		for _, c := range newCards {
			gs.NewCards = append(gs.NewCards, c.GetID())
		}
	}

	// Include last moved warrior ID for animation (only on the move action itself)
	if game.lastResult.Action == types.LastActionMoveWarrior && game.lastResult.MovedWarriorID != "" {
		gs.LastMovedWarriorID = game.lastResult.MovedWarriorID
	}

	// Include attack animation info (only on the attack action itself)
	if game.lastResult.Action == types.LastActionAttack && game.lastResult.AttackWeaponID != "" {
		gs.LastAttackWeaponID = game.lastResult.AttackWeaponID
		gs.LastAttackTargetID = game.lastResult.AttackTargetID
		gs.LastAttackTargetPlayer = game.lastResult.AttackTargetPlayer
	}

	// Include stolen card info for the victim (only on the steal action itself)
	if game.lastResult.Action == types.LastActionSteal && game.lastResult.StolenFrom != "" &&
		game.lastResult.StolenCard != nil && viewer.Name() == game.lastResult.StolenFrom {
		gs.StolenFromYouCard = gamestatus.FromDomainCards([]ports.Card{game.lastResult.StolenCard})
	}

	// Include spy notification for all players except the spy
	if game.lastResult.Spy.Target != "" && game.lastResult.Action == types.LastActionSpy &&
		viewer.Name() != game.CurrentPlayer().Name() {
		spyPlayer := game.CurrentPlayer().Name()
		if game.lastResult.Spy.Target == types.SpyTargetDeck {
			gs.SpyNotification = spyPlayer + " spied on the deck"
		} else {
			gs.SpyNotification = spyPlayer + " spied on " + game.lastResult.Spy.TargetPlayer + "'s hand"
		}
	}

	processHandCards(viewer, game, &gs)

	for _, warrior := range viewer.Field().Warriors() {
		gs.CurrentPlayerField = append(gs.CurrentPlayerField, gamestatus.NewFieldCard(warrior))
	}

	processOpponents(viewer, game, &gs)

	if over, winner := game.IsGameOver(); over {
		gs.GameOverMgs = "Game over! The winner is " + winner
		gs.IsWinner = game.isPlayerWinner(viewerIdx)
	}

	return gs
}

func processHandCards(viewer ports.Player, game *Game, gs *GameStatus) {
	action := game.currentAction
	canMove := game.turnState.CanMoveWarrior

	for _, card := range viewer.Hand().ShowCards() {
		switch ct := card.(type) {
		case ports.Warrior:
			gs.CanMoveWarrior = canMove
			gs.CurrentPlayerHand = append(gs.CurrentPlayerHand, gamestatus.NewWarriorHandCard(ct))

		case ports.Weapon:
			var enemyFields []ports.Field
			for _, enemy := range game.Enemies(viewer.Idx()) {
				enemyFields = append(enemyFields, enemy.Field())
			}

			if ct.Type() == types.SpecialPowerWeaponType {
				var allyFields []ports.Field
				for _, ally := range game.Allies(viewer.Idx()) {
					allyFields = append(allyFields, ally.Field())
				}

				gs.CurrentPlayerHand = append(gs.CurrentPlayerHand,
					gamestatus.NewSpecialPowerHandCard(ct.(ports.SpecialPower), viewer.Field(),
						allyFields, enemyFields, action))

				continue
			}

			gs.CurrentPlayerHand = append(gs.CurrentPlayerHand,
				gamestatus.NewWeaponHandCard(ct, viewer.Field(),
					enemyFields, viewer.Castle().IsConstructed(), action))

		case ports.Catapult:
			canBeAttacked := false
			for _, enemy := range game.Enemies(viewer.Idx()) {
				if enemy.Castle().CanBeAttacked() {
					canBeAttacked = true
					break
				}
			}
			gs.CurrentPlayerHand = append(gs.CurrentPlayerHand,
				gamestatus.NewCatapultHandCard(ct.GetID(), canBeAttacked,
					action))

		case ports.Spy:
			gs.CurrentPlayerHand = append(gs.CurrentPlayerHand,
				gamestatus.NewSpyHandCard(ct.GetID(), action))

		case ports.Thief:
			gs.CurrentPlayerHand = append(gs.CurrentPlayerHand,
				gamestatus.NewThiefHandCard(ct.GetID(), action))

		case ports.Resource:
			allyCastleConstructed := false
			for _, ally := range game.Allies(game.PlayerIndex(viewer.Name())) {
				if ally.Castle().IsConstructed() {
					allyCastleConstructed = true
					break
				}
			}
			gs.CurrentPlayerHand = append(gs.CurrentPlayerHand,
				gamestatus.NewResourceHandCard(ct, viewer.Castle().IsConstructed(),
					allyCastleConstructed, viewer.CanBuyWith(ct), action))
		}
	}
}

func processOpponents(viewer ports.Player, game *Game, gs *GameStatus) {
	viewerIdx := game.PlayerIndex(viewer.Name())

	for i, p := range game.Players {
		if i == viewerIdx {
			continue
		}
		opp := OpponentStatus{
			PlayerName:     p.Name(),
			CardsInHand:    p.CardsInHand(),
			Castle:         gamestatus.NewCastle(p.Castle()),
			IsAlly:         game.SameTeam(viewerIdx, i),
			IsEliminated:   game.EliminatedPlayers[i],
			IsDisconnected: game.DisconnectedPlayers[i],
		}
		for _, warrior := range p.Field().Warriors() {
			opp.Field = append(opp.Field, gamestatus.NewFieldCard(warrior))
		}
		gs.Opponents = append(gs.Opponents, opp)
	}
}
