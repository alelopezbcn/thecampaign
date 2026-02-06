# TheCampaign: Multi-Player Support Plan

## Overview

Extend TheCampaign from 2-player to support three game modes:
- **1v1**: 2 players (unchanged behavior)
- **2v2**: 4 players in 2 teams (can buff allies with special powers)
- **FFA**: 3-5 players free-for-all (last standing or first castle wins)

**UI**: Focus + mini panels (your field large at bottom, opponents as smaller panels at top).

**Design choices made**:
- 2v2: Separate fields per player, special powers (Knight protect / Mage heal) can target teammate's warriors
- FFA: Any attack/spy/steal/catapult requires selecting which opponent to target
- Eliminated players (FFA) are skipped in turn rotation
- 1v1 remains fully backward compatible

---

## Phase 1: Backend Domain - Game Mode & Player Helpers

### 1.1 Add GameMode type

**New file:** `internal/domain/types/gamemode.go`

```go
package types

type GameMode string

const (
    GameMode1v1 GameMode = "1v1"
    GameMode2v2 GameMode = "2v2"
    GameModeFFA GameMode = "ffa"
)
```

### 1.2 Extend Game struct

**File:** `internal/domain/game.go`

Add these fields to the `Game` struct:

```go
type Game struct {
    // ... existing fields ...
    Mode              types.GameMode
    Teams             map[int][]int    // teamID -> player indices (2v2 only)
    EliminatedPlayers map[int]bool     // player index -> eliminated (FFA only)
}
```

### 1.3 Change NewGame signature

**File:** `internal/domain/game.go:37`

**From:**
```go
func NewGame(player1, player2 string, dealer ports.Dealer,
    gameStatusProvider GameStatusProvider) *Game
```

**To:**
```go
func NewGame(playerNames []string, mode types.GameMode, dealer ports.Dealer,
    gameStatusProvider GameStatusProvider) *Game
```

Changes:
- Accept `[]string` instead of two separate player names
- Validate player count: 1v1=2, 2v2=4, FFA=3-5
- Create N players in a loop
- For 2v2: arrange seating as `[T1P1, T2P1, T1P2, T2P2]` so the modular turn rotation `(CurrentTurn+1)%len(Players)` naturally alternates teams
- Initialize `Teams` map for 2v2
- Initialize `EliminatedPlayers` map for FFA

Example 2v2 seating:
```
Players[0] = Team1-Player1  (turn 0)
Players[1] = Team2-Player1  (turn 1)
Players[2] = Team1-Player2  (turn 2)
Players[3] = Team2-Player2  (turn 3)
Teams = {0: [0,2], 1: [1,3]}
```

### 1.4 Add new helper methods

**File:** `internal/domain/game.go`

Replace the 2-player assumption in `WhoIsCurrent()` with flexible helpers:

```go
// CurrentPlayer returns the player whose turn it is
func (g *Game) CurrentPlayer() ports.Player {
    return g.Players[g.CurrentTurn]
}

// GetPlayer returns a player by name, or nil if not found
func (g *Game) GetPlayer(name string) ports.Player {
    for _, p := range g.Players {
        if p.Name() == name { return p }
    }
    return nil
}

// PlayerIndex returns the index of a player by name, or -1
func (g *Game) PlayerIndex(name string) int {
    for i, p := range g.Players {
        if p.Name() == name { return i }
    }
    return -1
}

// Enemies returns all opponents (non-eliminated, non-ally) of a given player
func (g *Game) Enemies(playerIdx int) []ports.Player {
    var enemies []ports.Player
    for i, p := range g.Players {
        if i == playerIdx { continue }
        if g.EliminatedPlayers[i] { continue }
        if g.Mode == types.GameMode2v2 && g.SameTeam(playerIdx, i) { continue }
        enemies = append(enemies, p)
    }
    return enemies
}

// Allies returns teammates (for 2v2 only, excluding self)
func (g *Game) Allies(playerIdx int) []ports.Player {
    if g.Mode != types.GameMode2v2 { return nil }
    var allies []ports.Player
    for i, p := range g.Players {
        if i == playerIdx { continue }
        if g.SameTeam(playerIdx, i) {
            allies = append(allies, p)
        }
    }
    return allies
}

// SameTeam checks if two player indices are on the same team
func (g *Game) SameTeam(i, j int) bool {
    if g.Mode != types.GameMode2v2 { return false }
    for _, team := range g.Teams {
        hasI, hasJ := false, false
        for _, idx := range team {
            if idx == i { hasI = true }
            if idx == j { hasJ = true }
        }
        if hasI && hasJ { return true }
    }
    return false
}
```

Then refactor all action methods to use `g.CurrentPlayer()` and `g.Enemies()` instead of `p, e := g.WhoIsCurrent()`.

---

## Phase 2: Observer & Win Condition Changes

### 2.1 Change FieldWithoutWarriorsObserver

**File:** `internal/domain/ports/observers.go:19`

**From:**
```go
type FieldWithoutWarriorsObserver interface {
    OnFieldWithoutWarriors()
}
```

**To:**
```go
type FieldWithoutWarriorsObserver interface {
    OnFieldWithoutWarriors(playerName string)
}
```

### 2.2 Update field to pass player name

**File:** `internal/domain/field.go`

Add `playerName` field:
```go
type field struct {
    playerName        string    // NEW
    cards             []ports.Warrior
    gameEndedObserver ports.FieldWithoutWarriorsObserver
}
```

Change constructor:
```go
func NewField(playerName string, o ports.FieldWithoutWarriorsObserver) ports.Field {
    return &field{
        playerName:        playerName,
        cards:             []ports.Warrior{},
        gameEndedObserver: o,
    }
}
```

Update `RemoveWarrior` (line 88):
```go
if len(h.cards) == 0 {
    h.gameEndedObserver.OnFieldWithoutWarriors(h.playerName)
}
```

### 2.3 Update NewPlayer

**File:** `internal/domain/player.go:21`

Pass player name when creating field:
```go
field: NewField(name, fieldWithoutWarriorsObserver),
```

### 2.4 Multi-mode win conditions

**File:** `internal/domain/game.go:641-650`

**OnCastleCompletion:**
```go
func (g *Game) OnCastleCompletion(p ports.Player) {
    g.gameOver = true
    if g.Mode == types.GameMode2v2 {
        g.winner = p.Name() + "'s team"
    } else {
        g.winner = p.Name()
    }
}
```

**OnFieldWithoutWarriors:**
```go
func (g *Game) OnFieldWithoutWarriors(playerName string) {
    eliminatedIdx := g.PlayerIndex(playerName)

    switch g.Mode {
    case types.GameMode1v1:
        g.gameOver = true
        g.winner = g.CurrentPlayer().Name()

    case types.GameModeFFA:
        g.EliminatedPlayers[eliminatedIdx] = true
        g.addToHistory(playerName + " has been eliminated!")
        active := 0
        var lastActive string
        for i, p := range g.Players {
            if !g.EliminatedPlayers[i] {
                active++
                lastActive = p.Name()
            }
        }
        if active == 1 {
            g.gameOver = true
            g.winner = lastActive
        }

    case types.GameMode2v2:
        g.EliminatedPlayers[eliminatedIdx] = true
        g.addToHistory(playerName + " has been eliminated!")
        // Check if all enemies of the eliminated player's team are also eliminated
        // (i.e., the opposing team is fully eliminated)
        attackerIdx := g.CurrentTurn
        allEnemiesEliminated := true
        for _, enemy := range g.Enemies(attackerIdx) {
            enemyIdx := g.PlayerIndex(enemy.Name())
            if !g.EliminatedPlayers[enemyIdx] {
                allEnemiesEliminated = false
                break
            }
        }
        if allEnemiesEliminated {
            g.gameOver = true
            g.winner = g.CurrentPlayer().Name() + "'s team"
        }
    }
}
```

### 2.5 switchTurn skips eliminated players

**File:** `internal/domain/game.go:652`

```go
func (g *Game) switchTurn() {
    g.hasMovedWarrior = false
    g.hasTraded = false
    g.currentAction = types.ActionTypeDrawCard
    for {
        g.CurrentTurn = (g.CurrentTurn + 1) % len(g.Players)
        if !g.EliminatedPlayers[g.CurrentTurn] {
            break
        }
    }
}
```

---

## Phase 3: Target Player on Actions

### 3.1 Attack

**File:** `internal/domain/game.go:240`

**From:** `Attack(playerName, targetID, weaponID string)`
**To:** `Attack(playerName, targetPlayerName, targetID, weaponID string)`

```go
func (g *Game) Attack(playerName, targetPlayerName, targetID, weaponID string) (
    status GameStatus, err error) {

    p := g.CurrentPlayer()
    if p.Name() != playerName {
        return status, fmt.Errorf("%s not your turn", playerName)
    }

    // Validate target player is an enemy
    targetPlayer := g.GetPlayer(targetPlayerName)
    if targetPlayer == nil {
        return status, fmt.Errorf("target player %s not found", targetPlayerName)
    }
    pIdx := g.PlayerIndex(playerName)
    tIdx := g.PlayerIndex(targetPlayerName)
    if pIdx == tIdx {
        return status, errors.New("cannot attack yourself")
    }
    if g.SameTeam(pIdx, tIdx) {
        return status, errors.New("cannot attack your ally")
    }
    if g.EliminatedPlayers[tIdx] {
        return status, errors.New("cannot attack eliminated player")
    }

    // Get target from target player's field
    targetCard, ok := targetPlayer.GetCardFromField(targetID)
    if !ok {
        return status, errors.New("target card not in enemy field: " + targetID)
    }
    // ... rest of attack logic unchanged ...
}
```

### 3.2 Catapult

**From:** `Catapult(playerName string, cardPosition int)`
**To:** `Catapult(playerName, targetPlayerName string, cardPosition int)`

- Same target player validation as Attack
- `t.Attack(targetPlayer.Castle(), cardPosition)` instead of `t.Attack(e.Castle(), cardPosition)`

### 3.3 Spy

**From:** `Spy(playerName string, option int)`
**To:** `Spy(playerName, targetPlayerName string, option int)`

- Option 1 (deck): `targetPlayerName` can be empty, deck peek doesn't need a target
- Option 2 (hand): resolve `targetPlayer`, use `targetPlayer.Hand().ShowCards()`

### 3.4 Steal

**From:** `Steal(playerName string, cardPosition int)`
**To:** `Steal(playerName, targetPlayerName string, cardPosition int)`

- Same target validation, `targetPlayer.CardStolenFromHand(cardPosition)`

### 3.5 SpecialPower - ally targeting in 2v2

**File:** `internal/domain/game.go:287`

Extend target search to include ally fields:

```go
var targetCard ports.Card
var ok bool

// Search own field
targetCard, ok = p.GetCardFromField(targetID)
if !ok {
    // Search ally fields (2v2)
    for _, ally := range g.Allies(g.PlayerIndex(playerName)) {
        targetCard, ok = ally.GetCardFromField(targetID)
        if ok { break }
    }
}
if !ok {
    // Search enemy fields
    for _, enemy := range g.Enemies(g.PlayerIndex(playerName)) {
        targetCard, ok = enemy.GetCardFromField(targetID)
        if ok { break }
    }
}
if !ok {
    return status, errors.New("target card not valid: " + targetID)
}
```

Since card IDs are unique (UUIDs), this works without an explicit target player name parameter.

### 3.6 nextAction - multi-enemy checks

**File:** `internal/domain/game.go:663`

Line 672 currently checks single enemy:
```go
canAttackWithCatapult := p.HasCatapult() && enemy.Castle().CanBeAttacked()
```

Change to check ANY enemy:
```go
canAttackWithCatapult := false
if p.HasCatapult() {
    for _, e := range g.Enemies(g.CurrentTurn) {
        if e.Castle().CanBeAttacked() {
            canAttackWithCatapult = true
            break
        }
    }
}
```

---

## Phase 4: GameStatus & GameStatusProvider

### 4.1 New OpponentStatus struct

**File:** `internal/domain/gamestatus.go`

```go
type OpponentStatus struct {
    PlayerName   string
    Field        []gamestatus.FieldCard
    Castle       gamestatus.Castle
    CardsInHand  int
    IsAlly       bool
    IsEliminated bool
}
```

### 4.2 Replace singular enemy fields in GameStatus

**File:** `internal/domain/gamestatus.go:9`

**Remove:**
```go
EnemyField       []gamestatus.FieldCard
EnemyCastle      gamestatus.Castle
CardsInEnemyHand int
```

**Add:**
```go
Opponents []OpponentStatus
GameMode  string
```

### 4.3 Change newGameStatus signature

**From:** `newGameStatus(currentPlayer, enemy ports.Player, game *Game, newCards ...ports.Card)`
**To:** `newGameStatus(viewer ports.Player, game *Game, newCards ...ports.Card)`

Build opponents by iterating all game players, skipping the viewer:

```go
viewerIdx := game.PlayerIndex(viewer.Name())

for i, p := range game.Players {
    if i == viewerIdx { continue }
    opp := OpponentStatus{
        PlayerName:   p.Name(),
        CardsInHand:  p.CardsInHand(),
        Castle:       gamestatus.NewCastle(p.Castle()),
        IsAlly:       game.SameTeam(viewerIdx, i),
        IsEliminated: game.EliminatedPlayers[i],
    }
    for _, warrior := range p.Field().Warriors() {
        opp.Field = append(opp.Field, gamestatus.NewFieldCard(warrior))
    }
    gs.Opponents = append(gs.Opponents, opp)
}
```

### 4.4 Change GameStatusProvider interface

**File:** `internal/domain/gamestatusprovider.go`

**From:**
```go
type GameStatusProvider interface {
    Get(p, e ports.Player, game *Game, newCards ...ports.Card) GameStatus
    GetWithModal(p, e ports.Player, game *Game, modalCards []ports.Card) GameStatus
}
```

**To:**
```go
type GameStatusProvider interface {
    Get(viewer ports.Player, game *Game, newCards ...ports.Card) GameStatus
    GetWithModal(viewer ports.Player, game *Game, modalCards []ports.Card) GameStatus
}
```

**This changes ~15 call sites** in game.go wherever `g.GameStatusProvider.Get(p, e, g, ...)` is called.

### 4.5 Update HandCard builders for multiple fields

**File:** `internal/domain/gamestatus/handcard.go`

**NewWeaponHandCard** - take `enemyFields []ports.Field`:
```go
func NewWeaponHandCard(weapon ports.Weapon, myField ports.Field,
    enemyFields []ports.Field, castleConstructed bool, action types.ActionType) HandCard {
    // ...
    // Build attackableIDs from ALL enemy fields
    for _, ef := range enemyFields {
        for _, v := range ef.Warriors() {
            mults[v.GetID()] = weapon.MultiplierFactor(v)
            attackableIDs = append(attackableIDs, v.GetID())
        }
    }
}
```

**NewSpecialPowerHandCard** - take `allyFields, enemyFields []ports.Field`:
```go
func NewSpecialPowerHandCard(sp ports.SpecialPower, myField ports.Field,
    allyFields []ports.Field, enemyFields []ports.Field, action types.ActionType) HandCard {
    // Archer: targets from ALL enemy fields
    if myField.HasArcher() {
        for _, ef := range enemyFields {
            for _, warrior := range ef.Warriors() {
                canBeUsedOnIDs = append(canBeUsedOnIDs, warrior.GetID())
            }
        }
    }
    // Knight/Mage: targets from own field AND ally fields
    if myField.HasKnight() {
        // Own field warriors...
        for _, af := range allyFields {
            for _, warrior := range af.Warriors() {
                // Same protection/dragon checks...
                canBeUsedOnIDs = append(canBeUsedOnIDs, warrior.GetID())
            }
        }
    }
    // Same for HasMage()...
}
```

**NewCatapultHandCard** - take `anyEnemyCastleAttackable bool`:
```go
func NewCatapultHandCard(cardID string, canBeAttacked bool, action types.ActionType) HandCard
```
Compute `canBeAttacked` from any enemy castle before calling.

---

## Phase 5: Dealer & Deck Scaling

### 5.1 Update Dealer interface

**File:** `internal/domain/ports/dealer.go`

```go
type Dealer interface {
    WarriorsCards(playerCount int) []Card
    OtherCards(playerCount int) []Card
}
```

### 5.2 Scale cards for player count

**File:** `internal/domain/cards/helper.go`

| Players | Warriors per type | Total warriors | Extra weapons/gold |
|---------|-------------------|----------------|--------------------|
| 2       | 5                 | 15             | Current amounts    |
| 3-4     | 7                 | 21             | ~1.5x current      |
| 5       | 9                 | 27             | ~2x current        |

### 5.3 Update game.deal()

**File:** `internal/domain/game.go:67`

Pass `len(g.Players)` to dealer. The existing loop that deals 3 warriors + 4 cards per player already iterates `g.Players`, so it scales naturally.

---

## Phase 6: WebSocket Layer Changes

### 6.1 Update message payloads

**File:** `internal/websocket/message.go`

Add `TargetPlayer string` to:
- `AttackPayload`
- `SpyPayload`
- `StealPayload`
- `CatapultPayload`
- `SpecialPowerPayload` (optional, for explicit targeting)

Add to `JoinGamePayload`:
- `GameMode string` (e.g., "1v1", "2v2", "ffa")
- `MaxPlayers int` (for FFA: 3, 4, or 5)

### 6.2 New OpponentStatusDTO

```go
type OpponentStatusDTO struct {
    PlayerName   string         `json:"player_name"`
    Field        []FieldCardDTO `json:"field"`
    Castle       CastleDTO      `json:"castle"`
    CardsInHand  int            `json:"cards_in_hand"`
    IsAlly       bool           `json:"is_ally"`
    IsEliminated bool           `json:"is_eliminated"`
}
```

Update `GameStatusDTO`: remove `EnemyField`/`EnemyCastle`/`CardsInEnemyHand`, add `Opponents []OpponentStatusDTO` and `GameMode string`.

### 6.3 Update handleJoinGame

**File:** `internal/websocket/hub.go:139`

- Add `GameMode` and `MaxPlayers` fields to `GameRoom` struct
- Change game start condition:
  - 1v1: start when 2 players
  - 2v2: start when 4 players
  - FFA: start when `MaxPlayers` joined (or add `MsgStartGame` for room creator)
- Pass mode to `domain.NewGame(playerNames, mode, ...)`

### 6.4 Update handlers

Pass `p.TargetPlayer` to domain methods for Attack, Spy, Steal, Catapult.

### 6.5 Update sendGameStateWithStatus

Loop all room players, compute each perspective via `GameStatusProvider.Get(viewer, game)`.

### 6.6 Update sendInitialWarriors, sendReconnectState

Use `game.CurrentPlayer()` and iterate all players.

---

## Phase 7: Frontend Changes

### 7.1 Join screen - game mode selection
Add `<select id="game-mode">` with 1v1/2v2/FFA options. Show player count for FFA.

### 7.2 Waiting screen
Show "Players: X / Y" and list of joined player names.

### 7.3 Layout: Focus + Mini Panels

Replace single `enemy-board` with dynamically populated `opponents-container`:

```
┌──────────────────────────────────────────────┐
│  [Opponent 1]  [Opponent 2]  [Opponent 3]... │  <- mini panels at top
│         (click to select as target)          │
├──────────────────────────────────────────────┤
│           Deck | Discard | Cemetery          │
├──────────────────────────────────────────────┤
│  Your Field (large)          Your Castle     │
│  Your Hand                                   │
└──────────────────────────────────────────────┘
```

Each mini panel shows:
- Player name + ally/eliminated badge
- Mini field cards (smaller card elements)
- Mini castle indicator
- Card count in hand

### 7.4 CSS for mini panels
```css
.opponents-container { display: flex; justify-content: space-around; gap: 10px; }
.opponent-panel { flex: 1; max-width: 300px; border: 2px solid rgba(255,255,255,0.1); border-radius: 12px; padding: 8px; cursor: pointer; }
.opponent-panel.selected-target { border-color: #ff4444; box-shadow: 0 0 15px rgba(255,68,68,0.3); }
.opponent-panel.ally { border-color: rgba(76,175,80,0.3); }
.opponent-panel.eliminated { opacity: 0.4; pointer-events: none; }
.mini-card { width: 50px; height: 70px; font-size: 0.6em; }
```

### 7.5 JS changes
- Parse `opponents` array from game state (instead of singular `enemy_field`/`enemy_castle`)
- Add `gameState.selectedTarget` for opponent selection
- Add `gameState.gameMode`
- New `renderOpponents(opponents)` function
- New `selectOpponentTarget(playerName)` function
- Update `findCardById()` to search all opponent fields
- Update `sendAction()` to include `target_player`
- For 1v1: auto-select the single opponent (preserves current UX)
- Turn indicator: show `"PlayerName's Turn"` instead of `"Enemy Turn"`

### 7.6 Target selection flow
1. Player clicks weapon/spy/thief/catapult in hand
2. If 1 opponent: auto-select (current behavior)
3. If multiple: highlight opponent panels, prompt to click one
4. Player clicks opponent panel -> that opponent's field becomes targetable
5. For weapon attacks: click specific warrior on opponent's mini-field
6. For catapult/steal: clicking panel directly opens the modal

---

## Phase 8: Recommended Implementation Order

1. **Phase 1** (game mode type, struct changes, helpers) + **Phase 2** (observer changes, win conditions)
2. **Phase 3** (target player on actions)
3. **Phase 4** (GameStatus restructure) - **highest impact**, do as one focused changeset
4. **Phase 5** (dealer scaling)
5. **Phase 6** (WebSocket layer)
6. **Phase 7** (frontend)
7. Test throughout, with comprehensive testing at end

**Key risk:** Phase 4 changes the `GameStatusProvider` interface used by every action method (~15 call sites). Do this in one pass.

**Backward compatibility:** In 1v1 mode, `Opponents` has length 1, `Enemies()` returns one player, frontend auto-selects the single opponent. No 1v1-specific code paths needed.
