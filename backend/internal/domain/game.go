package domain

import (
	"errors"
	"fmt"
	"time"

	"github.com/alelopezbcn/thecampaign/internal/domain/board"
	"github.com/alelopezbcn/thecampaign/internal/domain/cards"
	"github.com/alelopezbcn/thecampaign/internal/domain/gameactions"
	"github.com/alelopezbcn/thecampaign/internal/domain/gamestatus"
	"github.com/alelopezbcn/thecampaign/internal/domain/types"
	"github.com/google/uuid"
)

const (
	maxCastleResourcesFFA = 25
	maxCastleResources2v2 = 30
)

type Games []game

type game struct {
	id                  string
	board               board.Board
	createdAt           time.Time
	mode                types.GameMode
	teams               map[int][]int // teamID -> player indices (2v2 only)
	eliminatedPlayers   map[int]bool  // player index -> eliminated (FFA only)
	disconnectedPlayers map[int]bool  // player index -> disconnected
	currentTurn         int
	currentAction       types.PhaseType
	turnState           types.TurnState
	history             []types.HistoryLine
	historyTracker      int
	lastResult          gameactions.Result
	winState            winState
	gameStartedAt       time.Time
}

func NewGame(playerNames []string, mode types.GameMode, dealer cards.Dealer,
) (*game, error) {
	if err := validatePlayers(len(playerNames), mode); err != nil {
		return nil, err
	}

	now := time.Now()
	players := make([]board.Player, len(playerNames))
	g := &game{
		id:                  uuid.NewString(),
		currentTurn:         0,
		history:             []types.HistoryLine{},
		mode:                mode,
		eliminatedPlayers:   make(map[int]bool),
		disconnectedPlayers: make(map[int]bool),
		gameStartedAt:       now,
		turnState:           types.TurnState{StartedAt: now},
	}

	castleResourcesToWin := maxCastleResourcesFFA
	if mode == types.GameMode2v2 {
		castleResourcesToWin = maxCastleResources2v2
		g.teams = map[int][]int{
			1: {0, 2}, // Team 1: Player 1 and Player 3
			2: {1, 3}, // Team 2: Player 2 and Player 4
		}
	}

	for i, name := range playerNames {
		p := board.NewPlayer(name, i, g, g, g, g, castleResourcesToWin)
		players[i] = p
	}

	g.board = board.New(dealer, players)

	g.board.Deck().Deal(g.board.Players())

	return g, nil
}

func (g *game) CurrentAction() types.PhaseType {
	return g.currentAction
}

func (g *game) Board() board.Board {
	return g.board
}

func (g *game) TurnState() types.TurnState {
	return g.turnState
}

func (g *game) SetHasMovedWarrior(value bool) {
	g.turnState.HasMovedWarrior = value
}

func (g *game) SetHasTraded(value bool) {
	g.turnState.HasTraded = value
}

func (g *game) SetCanMoveWarrior(value bool) {
	g.turnState.CanMoveWarrior = value
}

func (g *game) SetCanTrade(value bool) {
	g.turnState.CanTrade = value
}

// AutoMoveWarriorToField moves a warrior to the field during game setup (no turn validation)
func (g *game) AutoMoveWarriorsToField(playerName string) error {
	p := g.GetPlayer(playerName)
	if p == nil {
		return fmt.Errorf("player %s not found", playerName)
	}

	i := 0
	snapshot := make([]cards.Card, len(p.Hand().ShowCards()))
	copy(snapshot, p.Hand().ShowCards())
	for _, c := range snapshot {
		if w, ok := c.(cards.Warrior); ok {
			if w.Type() == types.DragonWarriorType {
				continue // Skip dragon warriors during auto-move
			}

			_ = p.MoveCardToField(w.GetID())

			i++
			if i == 3 {
				return nil
			}
		}
	}

	return nil
}

func (g *game) AddHistory(msg string, cat types.Category) {
	if len(msg) == 0 {
		return
	}

	hl := types.HistoryLine{
		Msg:      msg,
		Category: cat,
	}
	g.history = append(g.history, hl)
}

func (g *game) getHistory() []types.HistoryLine {
	if g.historyTracker == 0 {
		g.historyTracker = len(g.history)
		return g.history
	}
	newMessages := g.history[g.historyTracker:]
	g.historyTracker = len(g.history)
	return newMessages
}

func (g *game) OnCardMovedToPile(card cards.Card) {
	g.board.DiscardPile().Discard(card)
}

func (g *game) OnWarriorMovedToCemetery(warrior cards.Warrior) {
	g.board.Cemetery().AddCorp(warrior)

	g.AddHistory("warrior buried in cemetery", types.CategoryInfo)
}

func (g *game) OnCastleCompletion(playerName string) {
	g.winState.GameOver = true
	g.winState.WinnerIdx = g.PlayerIndex(playerName)
	if g.mode == types.GameMode2v2 {
		g.winState.Winner = playerName + "'s team"
	} else {
		g.winState.Winner = playerName
	}
}

func (g *game) OnFieldWithoutWarriors(playerName string) {
	eliminatedIdx := g.PlayerIndex(playerName)

	switch g.mode {
	case types.GameMode1v1:
		g.winState.GameOver = true
		g.winState.Winner = g.CurrentPlayer().Name()
		g.winState.WinnerIdx = g.currentTurn
		return

	case types.GameModeFFA3, types.GameModeFFA5:
		g.eliminatedPlayers[eliminatedIdx] = true
		active := 0
		var lastActive string
		for i, p := range g.board.Players() {
			if !g.eliminatedPlayers[i] {
				active++
				lastActive = p.Name()
			}
		}
		if active == 1 {
			g.winState.GameOver = true
			g.winState.Winner = lastActive
			g.winState.WinnerIdx = g.PlayerIndex(lastActive)
		}

	case types.GameMode2v2:
		g.eliminatedPlayers[eliminatedIdx] = true
		// Check if all enemies of the eliminated player's team are also eliminated
		// (i.e., the opposing team is fully eliminated)
		attackerIdx := g.currentTurn
		allEnemiesEliminated := true
		for _, enemy := range g.Enemies(attackerIdx) {
			enemyIdx := g.PlayerIndex(enemy.Name())
			if !g.eliminatedPlayers[enemyIdx] {
				allEnemiesEliminated = false
				break
			}
		}
		if allEnemiesEliminated {
			g.winState.GameOver = true
			g.winState.Winner = g.CurrentPlayer().Name() + "'s team"
			g.winState.WinnerIdx = g.currentTurn
		}
	}

	g.AddHistory(playerName+" has been eliminated!", types.CategoryElimination)

	eliminatedPlayer := g.board.Players()[eliminatedIdx]
	// Move all cards from the eliminated player's hand to the discard pile
	for _, c := range eliminatedPlayer.Hand().ShowCards() {
		g.board.DiscardPile().Discard(c)
	}
	// Move all castled cards to the discard pile
	for _, c := range eliminatedPlayer.Castle().ResourceCards() {
		g.board.DiscardPile().Discard(c)
	}
}

// DisconnectPlayer marks a player as disconnected, removing them from turn rotation.
// If it's their turn, switches to the next player. State is preserved for reconnection.
func (g *game) DisconnectPlayer(playerName string) error {
	playerIdx := g.PlayerIndex(playerName)
	if playerIdx == -1 {
		return errors.New("player not found")
	}

	if g.winState.GameOver || g.eliminatedPlayers[playerIdx] || g.disconnectedPlayers[playerIdx] {
		return nil
	}

	wasTheirTurn := g.currentTurn == playerIdx
	g.disconnectedPlayers[playerIdx] = true
	g.AddHistory(playerName+" disconnected", types.CategoryElimination)

	// Check win conditions
	isOut := func(i int) bool {
		return g.eliminatedPlayers[i] || g.disconnectedPlayers[i]
	}

	switch g.mode {
	case types.GameMode2v2:
		// Check if all members of any team are out
		for _, members := range g.teams {
			allOut := true
			for _, idx := range members {
				if !isOut(idx) {
					allOut = false
					break
				}
			}
			if allOut {
				// Opposing team wins
				for _, idx := range members {
					for j, p := range g.board.Players() {
						if j != idx && !isOut(j) && !g.SameTeam(idx, j) {
							g.winState.GameOver = true
							g.winState.Winner = p.Name() + "'s team"
							g.winState.WinnerIdx = j
							break
						}
					}
					if g.winState.GameOver {
						break
					}
				}
			}
		}
	default:
		active := 0
		var lastActive string
		for i, p := range g.board.Players() {
			if !isOut(i) {
				active++
				lastActive = p.Name()
			}
		}
		if active == 1 {
			g.winState.GameOver = true
			g.winState.Winner = lastActive
			g.winState.WinnerIdx = g.PlayerIndex(lastActive)
		} else if active == 0 {
			g.winState.GameOver = true
			g.winState.Winner = "nobody"
			g.winState.WinnerIdx = -1
		}
	}

	if wasTheirTurn && !g.winState.GameOver {
		g.SwitchTurn()
	}

	return nil
}

// ReconnectPlayer restores a disconnected player back into turn rotation.
func (g *game) ReconnectPlayer(playerName string) {
	playerIdx := g.PlayerIndex(playerName)
	if playerIdx == -1 {
		return
	}

	if !g.disconnectedPlayers[playerIdx] {
		return
	}

	g.disconnectedPlayers[playerIdx] = false
	g.AddHistory(playerName+" reconnected", types.CategoryElimination)
}

func (g *game) ExecuteAction(action gameactions.GameAction) (
	status gamestatus.GameStatus, err error,
) {
	if g.CurrentPlayer().Name() != action.PlayerName() {
		return status, fmt.Errorf("%s not your turn", action.PlayerName())
	}
	if err := action.Validate(g); err != nil {
		return status, err
	}

	result, gameStatusFn, err := action.Execute(g)
	if err != nil {
		return status, err
	}

	g.lastResult = *result

	nextPhase := action.NextPhase()

	return g.nextAction(nextPhase, gameStatusFn), nil
}

// CurrentPlayer returns the player whose turn it is
func (g *game) CurrentPlayer() board.Player {
	return g.board.Players()[g.currentTurn]
}

// GetPlayer returns a player by name, or nil if not found
func (g *game) GetPlayer(name string) board.Player {
	for _, p := range g.board.Players() {
		if p.Name() == name {
			return p
		}
	}
	return nil
}

// PlayerIndex returns the index of a player by name, or -1
func (g *game) PlayerIndex(name string) int {
	for i, p := range g.board.Players() {
		if p.Name() == name {
			return i
		}
	}
	return -1
}

// Enemies returns all opponents (non-eliminated, non-ally) of a given player
func (g *game) Enemies(playerIdx int) []board.Player {
	var enemies []board.Player
	for i, p := range g.board.Players() {
		if i == playerIdx {
			continue
		}
		if g.eliminatedPlayers[i] {
			continue
		}
		if g.mode == types.GameMode2v2 && g.SameTeam(playerIdx, i) {
			continue
		}
		enemies = append(enemies, p)
	}
	return enemies
}

// Allies returns teammates (for 2v2 only, excluding self)
func (g *game) Allies(playerIdx int) []board.Player {
	if g.mode != types.GameMode2v2 {
		return nil
	}
	var allies []board.Player
	for i, p := range g.board.Players() {
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
func (g *game) SameTeam(i, j int) bool {
	if g.mode != types.GameMode2v2 {
		return false
	}
	for _, team := range g.teams {
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

func (g *game) GetTargetPlayer(playerName string, targetPlayerName string) (
	board.Player, error,
) {
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

	if g.eliminatedPlayers[tIdx] {
		return nil, errors.New("cannot attack eliminated player")
	}

	return targetPlayer, nil
}

func (g *game) isPlayerWinner(playerIdx int) bool {
	if !g.winState.GameOver {
		return false
	}
	if playerIdx == g.winState.WinnerIdx {
		return true
	}
	return g.SameTeam(playerIdx, g.winState.WinnerIdx)
}

func (g *game) DrawCards(p board.Player, count int) (cards []cards.Card, err error) {
	if !p.CanTakeCards(count) {
		return nil, board.ErrHandLimitExceeded
	}

	return g.board.Deck().DrawCards(count,
		g.board.DiscardPile())
}

func (g *game) SwitchTurn() {
	g.turnState = types.TurnState{StartedAt: time.Now()}
	g.lastResult = gameactions.Result{}
	g.currentAction = types.PhaseTypeDrawCard

	for {
		g.currentTurn = (g.currentTurn + 1) % len(g.board.Players())
		if !g.eliminatedPlayers[g.currentTurn] && !g.disconnectedPlayers[g.currentTurn] {
			break
		}
	}
}

func (g *game) nextAction(expectedAction types.PhaseType,
	gameStatusFn func() gamestatus.GameStatus,
) gamestatus.GameStatus {
	p := g.CurrentPlayer()
	g.turnState.CanMoveWarrior = !g.turnState.HasMovedWarrior && p.HasWarriorsInHand()
	g.turnState.CanTrade = !g.turnState.HasTraded && p.CanTradeCards()

	if expectedAction == types.PhaseTypeAttack {
		// Check if player can attack with weapons OR catapult
		canAttackWithCatapult := false

		if ok := board.HasCardTypeInHand[cards.Catapult](p); ok {
			for _, e := range g.Enemies(g.currentTurn) {
				if e.Castle().CanBeAttacked() {
					canAttackWithCatapult = true
					break
				}
			}
		}

		canAttackWithBloodRain := false
		if board.HasCardTypeInHand[cards.BloodRain](p) {
			for _, e := range g.Enemies(g.currentTurn) {
				if len(e.Field().Warriors()) > 0 {
					canAttackWithBloodRain = true
					break
				}
			}
		}

		canAttackWithHarpoon := false
		if board.HasCardTypeInHand[cards.Harpoon](p) {
			for _, e := range g.Enemies(g.currentTurn) {
				for _, w := range e.Field().Warriors() {
					if w.Type() == types.DragonWarriorType {
						canAttackWithHarpoon = true
						break
					}
				}
				if canAttackWithHarpoon {
					break
				}
			}
		}

		if p.CanAttack() || canAttackWithCatapult || canAttackWithBloodRain || canAttackWithHarpoon || g.turnState.CanMoveWarrior {
			g.currentAction = types.PhaseTypeAttack

			return gameStatusFn()
		}

		expectedAction = types.PhaseTypeSpySteal
	}

	if expectedAction == types.PhaseTypeSpySteal {
		hasSpy := board.HasCardTypeInHand[cards.Spy](p)
		hasThief := board.HasCardTypeInHand[cards.Thief](p)
		hasSabotage := board.HasCardTypeInHand[cards.Sabotage](p)

		canUseDesertion := false
		if board.HasCardTypeInHand[cards.Desertion](p) {
			for _, e := range g.Enemies(g.currentTurn) {
				for _, w := range e.Field().Warriors() {
					if w.Health() <= cards.DesertionMaxHP {
						canUseDesertion = true
						break
					}
				}
				if canUseDesertion {
					break
				}
			}
		}

		if hasSpy || hasThief || hasSabotage || canUseDesertion {
			g.currentAction = types.PhaseTypeSpySteal

			return gameStatusFn()
		}

		expectedAction = types.PhaseTypeBuy
	}

	if expectedAction == types.PhaseTypeBuy {
		canPlaceAmbush := false
		if board.HasCardTypeInHand[cards.Ambush](p) {
			canPlaceAmbush = !board.HasFieldSlotCard[cards.Ambush](p.Field())
		}
		if p.CanBuy() || g.turnState.CanTrade || canPlaceAmbush {
			g.currentAction = types.PhaseTypeBuy

			return gameStatusFn()
		}

		expectedAction = types.PhaseTypeConstruct
	}

	if expectedAction == types.PhaseTypeConstruct {
		canConstruct := p.CanConstruct()
		if !canConstruct {
			// In 2v2, check if player has resources and any ally has a constructed castle
			for _, ally := range g.Allies(g.PlayerIndex(p.Name())) {
				if ally.Castle().IsConstructed() {
					for _, c := range p.Hand().ShowCards() {
						if _, ok := c.(cards.Resource); ok {
							canConstruct = true
							break
						}
					}
					break
				}
			}
		}
		if !canConstruct {
			// Check if player has a Fortress card and a valid target castle
			if board.HasCardTypeInHand[cards.Fortress](p) {
				if p.Castle().IsConstructed() && !p.Castle().IsProtected() {
					canConstruct = true
				}
				if !canConstruct {
					for _, ally := range g.Allies(g.PlayerIndex(p.Name())) {
						if ally.Castle().IsConstructed() && !ally.Castle().IsProtected() {
							canConstruct = true
							break
						}
					}
				}
			}
		}
		if canConstruct {
			g.currentAction = types.PhaseTypeConstruct

			return gameStatusFn()
		}

		expectedAction = types.PhaseTypeEndTurn
	}

	if expectedAction == types.PhaseTypeEndTurn {
		g.currentAction = types.PhaseTypeEndTurn

		return gameStatusFn()
	}

	g.currentAction = types.PhaseTypeDrawCard

	return gameStatusFn()
}

// NextActiveTurnPlayer returns the name of the player who will go next,
// without mutating any state.
func (g *game) nextActiveTurnPlayer() string {
	next := g.currentTurn
	for {
		next = (next + 1) % len(g.board.Players())
		if !g.eliminatedPlayers[next] && !g.disconnectedPlayers[next] {
			return g.board.Players()[next].Name()
		}
		if next == g.currentTurn {
			return ""
		}
	}
}

func validatePlayers(playersCount int, mode types.GameMode) error {
	switch mode {
	case types.GameMode1v1:
		if playersCount != 2 {
			return errors.New("1v1 mode requires 2 players")
		}
	case types.GameMode2v2:
		if playersCount != 4 {
			return errors.New("2v2 mode requires 4 players")
		}
	case types.GameModeFFA3:
		if playersCount != 3 {
			return errors.New("FFA3 mode requires 3 players")
		}
	case types.GameModeFFA5:
		if playersCount != 5 {
			return errors.New("FFA mode requires 5 players")
		}
	default:
		return errors.New("invalid game mode")
	}

	return nil
}

func (g *game) Status(viewer board.Player, newCards ...cards.Card,
) gamestatus.GameStatus {
	return g.getStatus(viewer, newCards, nil)
}

func (g *game) StatusWithModal(viewer board.Player,
	modalCards []cards.Card,
) gamestatus.GameStatus {
	return g.getStatus(viewer, nil, modalCards)
}

func (g *game) getStatus(viewer board.Player,
	newCards []cards.Card, modalCards []cards.Card,
) gamestatus.GameStatus {
	viewerIdx := g.PlayerIndex(viewer.Name())

	viewerInput := gamestatus.ViewerInput{
		Name:       viewer.Name(),
		Idx:        viewerIdx,
		Hand:       viewer.Hand().ShowCards(),
		Field:      extractField(viewer.Field()),
		Castle:     extractCastle(viewer.Castle()),
		CanBuyWith: viewer.CanBuyWith,
	}

	allPlayers := g.board.Players()
	playersNames := make([]string, len(allPlayers))
	for i, p := range allPlayers {
		playersNames[i] = p.Name()
	}

	opponents := []gamestatus.OpponentInput{}
	for i, p := range allPlayers {
		if i == viewerIdx {
			continue
		}
		opponents = append(opponents, gamestatus.OpponentInput{
			Name:           p.Name(),
			CardsInHand:    p.CardsInHand(),
			Field:          extractField(p.Field()),
			Castle:         extractCastle(p.Castle()),
			IsAlly:         g.SameTeam(viewerIdx, i),
			IsEliminated:   g.eliminatedPlayers[i],
			IsDisconnected: g.disconnectedPlayers[i],
		})
	}

	enemyFields := []gamestatus.FieldInput{}
	anyEnemyCastleAttackable := false
	anyEnemyHasCards := false
	anyEnemyHasWeakWarriors := false
	for _, enemy := range g.Enemies(viewerIdx) {
		enemyFields = append(enemyFields, extractField(enemy.Field()))
		if enemy.Castle().CanBeAttacked() {
			anyEnemyCastleAttackable = true
		}
		if enemy.CardsInHand() > 0 {
			anyEnemyHasCards = true
		}
		for _, w := range enemy.Field().Warriors() {
			if w.Health() <= cards.DesertionMaxHP {
				anyEnemyHasWeakWarriors = true
			}
		}
	}

	allyFields := []gamestatus.FieldInput{}
	allyHasCastleConstructed := false
	for _, ally := range g.Allies(viewerIdx) {
		allyFields = append(allyFields, extractField(ally.Field()))
		if ally.Castle().IsConstructed() {
			allyHasCastleConstructed = true
		}
	}

	gameStatusDTO := gamestatus.GameStatusDTO{
		Viewer:                   viewerInput,
		PlayersNames:             playersNames,
		Opponents:                opponents,
		EnemyFields:              enemyFields,
		AllyFields:               allyFields,
		AnyEnemyCastleAttackable: anyEnemyCastleAttackable,
		AnyEnemyHasCards:         anyEnemyHasCards,
		AnyEnemyHasWeakWarriors:  anyEnemyHasWeakWarriors,
		AllyHasCastleConstructed: allyHasCastleConstructed,
		NewCards:                 newCards,
		ModalCards:               modalCards,
		NextTurnPlayer:           g.nextActiveTurnPlayer(),
		TurnPlayer:               g.CurrentPlayer().Name(),
		CurrentAction:            g.CurrentAction(),
		LastAction:               g.lastResult.Action,
		GameMode:                 string(g.mode),
		IsEliminated:             g.eliminatedPlayers[viewerIdx],
		IsDisconnected:           g.disconnectedPlayers[viewerIdx],
		CanTrade:                 g.turnState.CanTrade,
		CemeteryCount:            g.board.Cemetery().Count(),
		CemeteryLastDead:         g.board.Cemetery().GetLast(),
		DiscardPileCount:         g.board.DiscardPile().Count(),
		DiscardPileLastCard:      g.board.DiscardPile().GetLast(),
		DeckCount:                g.board.Deck().Count(),
		GameStartedAt:            g.gameStartedAt,
		TurnStartedAt:            g.turnState.StartedAt,
		History:                  g.getHistory(),
		LastMovedWarriorID:       g.lastResult.MovedWarriorID,
		LastAttackWeaponID:       g.lastResult.AttackWeaponID,
		LastAttackTargetID:       g.lastResult.AttackTargetID,
		LastAttackTargetPlayer:   g.lastResult.AttackTargetPlayer,
		StolenFrom:               g.lastResult.StolenFrom,
		StolenCard:               g.lastResult.StolenCard,
		SabotagedFrom:            g.lastResult.SabotagedFrom,
		SabotagedCard:            g.lastResult.SabotagedCard,
		SpyTarget:                g.lastResult.Spy.Target,
		SpyTargetPlayer:          g.lastResult.Spy.TargetPlayer,
		CurrentPlayerName:        g.CurrentPlayer().Name(),
		IsPlayerWinner:           g.isPlayerWinner(viewerIdx),
		CanMoveWarrior:           g.turnState.CanMoveWarrior,
		AmbushEffect:             g.lastResult.AmbushEffect,
		AmbushAttackerName:       g.lastResult.AmbushAttackerName,
		DeserterFromPlayer:       g.lastResult.DeserterFromPlayer,
		DeserterWarrior:          g.lastResult.DeserterWarrior,
	}

	gameStatusDTO.IsGameOver, gameStatusDTO.Winner = g.IsGameOver()

	return gamestatus.NewGameStatus(gameStatusDTO)
}

func extractField(f board.Field) gamestatus.FieldInput {
	return gamestatus.FieldInput{
		Warriors:     f.Warriors(),
		HasArcher:    f.HasWarriorType(types.ArcherWarriorType),
		HasKnight:    f.HasWarriorType(types.KnightWarriorType),
		HasMage:      f.HasWarriorType(types.MageWarriorType),
		HasDragon:    f.HasWarriorType(types.DragonWarriorType),
		HasMercenary: f.HasWarriorType(types.MercenaryWarriorType),
		HasAmbush:    board.HasFieldSlotCard[cards.Ambush](f),
	}
}

func extractCastle(c board.Castle) gamestatus.CastleInput {
	return gamestatus.CastleInput{
		IsConstructed:      c.IsConstructed(),
		IsProtected:        c.IsProtected(),
		ResourceCardsCount: c.ResourceCardsCount(),
		Value:              c.Value(),
	}
}

func (g *game) IsGameOver() (bool, string) {
	return g.winState.GameOver, g.winState.Winner
}
