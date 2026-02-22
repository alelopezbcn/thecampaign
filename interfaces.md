# Consumer-Side Interface Pattern: Architecture Document

## 1. The Pattern

In Go, the idiomatic rule is: **interfaces belong to the package that consumes them, not to the package that defines them**. Each consumer declares only the narrow slice of behaviour it actually needs. Concrete types satisfy those interfaces implicitly.

The current code applies this correctly in one place (`gameactions.Game`) but violates it everywhere else: broad interfaces (`board.Player`, `board.Board`) are defined in `board`, then imported wholesale by `domain`, `gamestatus`, and `gameactions`, creating unnecessary coupling and making unit tests needlessly heavy.

---

## 2. Current State

### 2.1 Dependency Graph (before)

```
cards
  ↑
board             ← defines Board, Player, Field, Castle, Hand, Deck, DiscardPile, Cemetery
  ↑         ↑
domain    gamestatus     ← both import board.Player, board.Field, board.Castle directly
  ↑
gameactions               ← defines Game interface (already consumer-side ✓) but imports board.Player
  ↑
websocket                 ← defines HubGame interface (already consumer-side ✓)
```

### 2.2 Problems

| # | Problem | Where |
|---|---------|-------|
| P1 | board.go defines the `Board` aggregate interface that wraps all other board types. `domain.game` stores a `board.Board` field, forcing all tests to mock the entire board object instead of individual components | game.go line 25 |
| P2 | `board.Player` is a very broad interface (26 methods). Any consumer that needs only `Name()` and `Hand()` is still coupled to the full interface | everywhere `board.Player` is accepted |
| P3 | `gameactions.Game` exposes `Board() board.Board`, which means game actions can reach into the full board. Only gameaction_spy.go actually uses it (`g.Board().Deck().Reveal(5)`) | gameaction.go line 30, gameaction_spy.go line 60 |
| P4 | `board.Player` is missing `Attack(target cards.Attackable, weapon cards.Weapon) error` and presence-check helpers (`HasCatapult`, `HasSpy`, `HasThief`, `HasBloodRain`, `HasHarpoon`). The logic lives in `gameactions` (validation) and game.go (`nextAction`), but it belongs on the Player. Tests already expect these methods | player_test.go lines 253–362, game_test.go line 112 |
| P5 | `CastleCompletionObserver.OnCastleCompletion(playerName string)` loses the player object. game.go must then call `g.PlayerIndex(playerName)` to reconstruct the winner index. Tests expect the signature `OnCastleCompletion(player board.Player)` | castle.go line 22, game_test.go line 199 |
| P6 | `WinState` is defined in `domain/types`. It is only ever used in `domain.game`, so it leaks an internal struct into a shared types package | winstate.go |
| P7 | `gameactions.Result` is in the `gameactions` package but `domain.game` stores it as a field. Tests reference it as `Result{...}` (unqualified) from inside `package domain`, which means either the type must move to `domain` or domain must re-export it | game_test.go line 99 |
| P8 | `TurnState` is in `types/` but is only read/written by `domain.game`. It is passed to `gameactions.Game` via `TurnState()` and thence into status generation | turnstate.go |
| P9 | field_test.go is `package board_test` but directly instantiates the unexported `field` struct (`&field{...}` and `&board.field{...}`). This is inconsistent and does not compile | field_test.go lines 14, 27, 39, 51 |
| P10 | `domain.game` calls `board.HasCardTypeInHand[cards.Xxx](p)` in `nextAction`. This generic helper is defined in the `board` package but the domain is the consumer. It creates a dependency on `board` just for a loop | game.go lines 503–549 |
| P11 | `gamestatus` imports `board.Player`, `board.Field`, `board.Castle` directly for its functions `processHandCards`, `processOpponents`, and constructors such as `NewCastle`. It therefore must know the complete board hierarchy even though it only reads a small subset | gamestatus.go, handcard.go, castle.go |
| P12 | `gameactions` tests (`package gameactions_test`) use a local `game` struct defined nowhere, plus `NewMockGameStatusProvider` which also does not exist yet | gameaction_attack_test.go line 37 |

---

## 3. Proposed Architecture

### 3.1 Dependency Graph (after)

```
cards
  ↑
board             ← concrete implementations only; minimal shared interfaces
  ↑
domain            ← defines Game interface (what websocket needs)
  |               ← defines GameAction interface (what domain.game needs from actions)
  |               ← owns WinState, TurnState, Result
  ↑
gameactions       ← defines its own narrow Game interface (what actions need from domain)
  |               ← defines GameStatusProvider interface
  ↑
gamestatus        ← defines its own narrow viewer/player/field/castle interfaces
  ↑
websocket         ← already consumer-side; unchanged
```

Each arrow means "depends on". No package reaches upward.

---

### 3.2 `board` Package

**Role:** Concrete game board components. Provides `Player`, `Hand`, `Field`, `Castle`, `Deck`, `DiscardPile`, and `Cemetery` as interfaces so they can be mocked in the packages that consume them. The key change is that `board` stops defining a broad `Board` aggregate and stops owning observer interfaces that belong to the domain.

**`board.Player` interface additions:**
```go
// New methods the tests already expect
Attack(target cards.Attackable, weapon cards.Weapon) error
HasWarriorsInField() bool   // already implied by Field().Warriors()
HasCatapult() bool          // convenience: checks hand for cards.Catapult
HasSpy() bool               // convenience: checks hand for cards.Spy
HasThief() bool             // convenience: checks hand for cards.Thief
HasBloodRain() bool         // convenience: checks hand for cards.BloodRain
HasHarpoon() bool           // convenience: checks hand for cards.Harpoon
```

`Attack` moves the weapon-type vs warrior-type validation out of `gameactions/gameaction_attack.go.Validate` and onto the player, where it belongs. The `Has*` helpers replace `board.HasCardTypeInHand[T]` and `board.HasCardTypeInHand[cards.Xxx](p)` calls scattered across game.go and `gameactions/`.

**`CastleCompletionObserver` signature change:**
```go
// Before (board/castle.go)
type CastleCompletionObserver interface {
    OnCastleCompletion(playerName string)
}

// After
type CastleCompletionObserver interface {
    OnCastleCompletion(player Player)
}
```
This lets `domain.game.OnCastleCompletion` receive the player directly and derive the winner index without a second lookup.

**Remove:** `board.Board` aggregate interface and board.go's `New` constructor can stay as a factory helper, but `board.Board` itself should not be passed around — consumers that need the deck, discard pile, and cemetery should depend on those interfaces individually.

**Remove:** `board.HasCardTypeInHand[T]` generic function. Each call site is replaced by the new `Has*` methods on `Player`.

---

### 3.3 `domain` Package

**Role:** The `game` struct is the root aggregate. It is the source of truth for turn state, player list, win conditions, and history. It communicates upward to `websocket` via the `Game` interface, and downward to `gameactions` via the `GameAction` interface.

**Struct fields change:** Instead of storing `board board.Board`, the game struct directly holds each board component:

```go
// Before
type game struct {
    board board.Board
    // ...
}

// After
type game struct {
    players     []board.Player
    deck        board.Deck
    discardPile board.DiscardPile
    cemetery    board.Cemetery
    // ...
}
```

This is the single most impactful change. It allows tests to inject individual mocks (`mockDeck`, `mockDiscardPile`, `mockCemetery`) without wrapping them in a `MockBoard`. It matches exactly what game_test.go already does.

**Types that belong here (moved from `types/` or `gameactions/`):**

```go
// domain/winstate.go  (moved from types/winstate.go — only domain uses it)
type WinState struct {
    GameOver  bool
    Winner    string
    WinnerIdx int
}

// domain/turnstate.go  (moved from types/turnstate.go — only domain.game reads/writes it)
type TurnState struct {
    CanMoveWarrior  bool
    HasMovedWarrior bool
    CanTrade        bool
    HasTraded       bool
    StartedAt       time.Time
}

// domain/result.go  (moved from gameactions/result.go — domain.game stores it as lastResult)
type Result struct {
    Action             types.LastActionType
    MovedWarriorID     string
    StolenFrom         string
    StolenCard         cards.Card
    Spy                types.SpyInfo
    AttackWeaponID     string
    AttackTargetID     string
    AttackTargetPlayer string
}
```

**`GameAction` interface (consumer-side: what `domain.game.ExecuteAction` needs):**

```go
// domain/gameaction.go
type GameAction interface {
    PlayerName() string
    Validate(g GameContext) error
    Execute(g GameContext) (*Result, func() gamestatus.GameStatus, error)
    NextPhase() types.PhaseType
}
```

`GameContext` is a narrow interface exposing only what actions need (see §3.4).

**`Game` interface (consumer-side: what `websocket` needs from `domain`):**

The hub.go already defines this inline. It moves into `domain` so the domain package is the canonical source:

```go
// domain/game_interface.go
type Game interface {
    CurrentPlayer() board.Player
    GetPlayer(name string) board.Player
    Status(viewer board.Player, newCards ...cards.Card) gamestatus.GameStatus
    StatusWithModal(viewer board.Player, modalCards []cards.Card) gamestatus.GameStatus
    ExecuteAction(action GameAction) (gamestatus.GameStatus, error)
    IsGameOver() (bool, string)
    DisconnectPlayer(playerName string) error
    ReconnectPlayer(playerName string)
    AutoMoveWarriorToField(playerName string, cardID string) error
}
```

`websocket.HubGame` then embeds or mirrors this interface (it already does). No duplication occurs because `websocket` imports `domain.Game`.

**Mock for tests:**

```go
// domain/gameaction_mock.go  (generated by mockgen)
// source: domain/gameaction.go, interface: GameAction
```

This replaces the mysterious `NewMockGameAction` that game_test.go already calls without a source.

---

### 3.4 `gameactions` Package

**Role:** Implements every game action. Each file defines one action struct implementing the domain-level `GameAction` interface.

**The `Game` interface:** Actions need more from the game than websocket does. This is the consumer-side interface specific to the `gameactions` package:

```go
// gameactions/gameaction.go — renamed from the current broad Game interface
type GameContext interface {
    CurrentPlayer() board.Player
    CurrentAction() types.PhaseType
    TurnState() domain.TurnState
    GetTargetPlayer(playerName, targetPlayerName string) (board.Player, error)
    AddHistory(msg string, cat types.Category)
    Status(viewer board.Player, newCards ...cards.Card) gamestatus.GameStatus
    StatusWithModal(viewer board.Player, modalCards []cards.Card) gamestatus.GameStatus
    OnCardMovedToPile(card cards.Card)
    DrawCards(player board.Player, count int) ([]cards.Card, error)
    SwitchTurn()
    SetHasMovedWarrior(bool)
    SetHasTraded(bool)
    SetCanMoveWarrior(bool)
    SetCanTrade(bool)
    GetPlayer(name string) board.Player
    PlayerIndex(name string) int
    SameTeam(i, j int) bool
    Allies(playerIdx int) []board.Player
    Enemies(playerIdx int) []board.Player
    RevealDeckCards(n int) []cards.Card   // replaces Board().Deck().Reveal(n) in gameaction_spy.go
}
```

`Board() board.Board` is removed. The single caller (gameaction_spy.go) is replaced by `RevealDeckCards(n int) []cards.Card`, which is a method on `GameContext` that `domain.game` implements trivially.

**`GameStatusProvider` interface (for test injection):**

Several `Execute` functions build a status closure. In tests, this closure is hard to intercept. The pattern the tests already use is:

```go
// gameactions/status_provider.go
type GameStatusProvider interface {
    Get(viewer board.Player, g GameContext, newCards ...cards.Card) gamestatus.GameStatus
}
```

Each action struct that needs a status closure holds a `GameStatusProvider`. In production, the provider delegates to `g.Status(viewer, newCards...)`. In tests, it is mocked with `NewMockGameStatusProvider`. This is the approach gameaction_draw_test.go and gameaction_attack_test.go already use.

**Test helper `game` struct:**

Every `*_test.go` file in `gameactions_test` instantiates `&game{players: ..., deck: ..., discardPile: ..., ...}`. This struct is a lightweight fake that lives in a shared test file:

```go
// gameactions/testhelpers_test.go  (package gameactions_test)
type game struct {
    players       []board.Player
    currentTurn   int
    currentAction types.PhaseType
    deck          board.Deck
    discardPile   board.DiscardPile
    history       []types.HistoryLine
    turnState     domain.TurnState
    gameStatusProvider GameStatusProvider

    // callbacks for assertions
    onCardMovedToPile func(cards.Card)
}

func (g *game) CurrentPlayer() board.Player { return g.players[g.currentTurn] }
func (g *game) CurrentAction() types.PhaseType { return g.currentAction }
// ... all other GameContext methods
```

This struct is **not** the production `domain.game` — it is the minimal fake that lets each test control exactly what the action sees, without importing the real domain.

---

### 3.5 `gamestatus` Package

**Role:** Pure projection — takes raw domain data and produces JSON-serialisable `GameStatus`.

**Current problem:** gamestatus.go and handcard.go import `board.Player`, `board.Field`, `board.Castle` directly. This means the gamestatus package knows about every method on those types.

**Consumer-side interfaces in `gamestatus`:**

```go
// gamestatus/interfaces.go

type playerView interface {
    Name() string
    Idx() int
    CardsInHand() int
    Hand() handViewer
    Field() fieldViewer
    Castle() castleViewer
    CanBuyWith(resource cards.Resource) bool
}

type handViewer interface {
    ShowCards() []cards.Card
    Count() int
}

type fieldViewer interface {
    Warriors() []cards.Warrior
}

type castleViewer interface {
    IsConstructed() bool
    Value() int
    ResourceCardsCount() int
    CanBeAttacked() bool
}
```

`GameStatusDTO.Viewer`, `GameStatusDTO.Players`, and the function signatures of `processHandCards`, `processOpponents`, `NewCastle`, `NewWeaponHandCard`, etc. all change from accepting `board.Player` / `board.Field` / `board.Castle` to accepting these narrow local interfaces.

`board.Player` (and all other board types) implicitly satisfy these interfaces, so no implementation changes in `board`. The `gamestatus` package simply stops importing `board` entirely.

---

### 3.6 `websocket` Package

Already consumer-side. hub.go defines `HubGame`:

```go
type HubGame interface {
    CurrentPlayer() board.Player
    GetPlayer(name string) board.Player
    Status(viewer board.Player, newCards ...cards.Card) gamestatus.GameStatus
    // ...
}
```

After the refactor, `HubGame` is replaced by (or becomes an alias for) `domain.Game`. The import of `board` in hub.go remains only because `board.Player` appears in the interface signature. If `websocket` were also to define its own narrow viewer interface, it could drop the `board` import too — but that is a lower-priority step.

---

## 4. Summary Table

| Package | Defines interfaces for whom | Imports |
|---------|----------------------------|---------|
| `cards` | `Card`, `Warrior`, `Weapon`, `Resource`, `Dealer`, etc. | nothing domain-specific |
| `board` | `Player`, `Hand`, `Field`, `Castle`, `Deck`, `DiscardPile`, `Cemetery`, observer interfaces | `cards` |
| `domain` | `Game` (for websocket), `GameAction` (for itself), `WinState`, `TurnState`, `Result` | `board`, `cards`, `gameactions` (for `GameAction` implementation types), `gamestatus` |
| `gameactions` | `GameContext` (for actions), `GameStatusProvider` (for testability) | `board`, `cards`, `gamestatus`, `domain` (for `TurnState`, `Result`) |
| `gamestatus` | `playerView`, `handViewer`, `fieldViewer`, `castleViewer` (for its own functions) | `cards`, `types` — **no longer imports `board`** |
| `websocket` | `HubGame` (mirrors `domain.Game`) | `domain`, `gamestatus`, `board` (minimal) |

---

## 5. Key Design Decisions

### 5.1 Why flatten `board.Board`?
The aggregate interface creates a false sense that "a board" is a thing that moves around in the system. In reality, only `domain.game` owns a board — sub-components (`Deck`, `DiscardPile`, `Cemetery`) are what other objects actually interact with. Flattening them into `game` fields makes construction simpler, mocking in tests surgical, and the code honest about what each method needs.

### 5.2 Why move `Attack` onto `Player`?
The attack validation in gameaction_attack.go (`if !a.currentPlayer.Field().HasKnight()...`) duplicates weapon-type logic. `Player.Attack(target, weapon)` encapsulates it. The game action's `Validate` becomes "does this player have a weapon in hand?" and `Execute` becomes "player, attack this target with this weapon". The player is the right authority on whether it can fire a given weapon type.

### 5.3 Why `GameStatusProvider` in `gameactions`?
The `statusFn := func() gamestatus.GameStatus { return g.Status(p, cards...) }` pattern makes each `Execute` hard to test in isolation because it closes over the real `g`. Injecting a `GameStatusProvider` breaks that closure and lets tests assert on the status call without instantiating a full game object.

### 5.4 Why keep `board.Player` as a broad interface rather than splitting it?
`board.Player` is already the primary abstraction for the player entity. Splitting it into many sub-interfaces would make the mock setup in tests more verbose. The right split is between **what external consumers need** (the narrow `gamestatus.playerView`) versus **what the domain needs** (the full `board.Player`). Consumers outside `domain` never receive a `board.Player` raw — they receive a narrow view.

### 5.5 Why move `WinState` and `TurnState` out of `types`?
`types` is a leaf package. Its purpose is to hold pure enumerations and value types used across many packages (`PhaseType`, `WarriorType`, `HistoryLine`, etc.). `WinState` and `TurnState` are internal state of `domain.game` — nothing outside `domain` needs them except as values passed into `GameStatusDTO`. Moving them to `domain` makes their scope explicit.

---

## 6. Migration Path

1. **`board`** — Add `Attack` and `Has*` methods to `Player` interface AND `player` implementation. Change `CastleCompletionObserver` signature. Remove `HasCardTypeInHand` generic helper.
2. **`domain`** — Add `WinState`, `TurnState`, `Result` as domain-level types. Change `game` struct to hold `players`, `deck`, `discardPile`, `cemetery` directly. Move `validatePlayers` to accept `[]string`. Create `GameAction` interface and mock.
3. **`gamestatus`** — Replace all `board.Player`/`board.Field`/`board.Castle` parameter types with the local narrow interfaces. Remove the `board` import.
4. **`gameactions`** — Rename `Game` → `GameContext`. Add `RevealDeckCards`. Add `GameStatusProvider` interface and mock. Create `testhelpers_test.go`. Update `Result` references to `domain.Result`.
5. **`mocks`** — Regenerate player_mocks.go (new methods), `castle_mocks.go` (observer signature).
6. **`websocket`** — Point `HubGame` at `domain.Game`. Remove duplicated interface.

User: can you write ir down in a plan.md document?

GitHub Copilot: Using "Create File"