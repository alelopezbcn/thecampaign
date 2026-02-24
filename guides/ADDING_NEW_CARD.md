# Adding a New Card Type

This guide documents every file that must be created or modified when adding a new card type. It was written after adding the **Fortress** card and covers both simple cards (resource/warrior variants) and complex cards with their own game action.

---

## Overview of layers

```
cards/          — card type definition (interface + struct)
board/          — player/castle changes if the card affects them
gameactions/    — game action that handles using the card
gamestatus/     — serialization: CardType, HandCard, GameStatusDTO inputs
game.go         — phase-skip logic, extractCastle if needed
websocket/      — message type, handler, hub routing, DTO
server/         — card config (image + description for the UI)
frontend/       — rendering, click handler, animations, CSS
```

---

## Step 1 — Define the card (`domain/cards/`)

**Create `cards/<name>.go`**

```go
package cards

type Fortress interface {
    Card
    IsFortressCard() bool // ⚠️ REQUIRED: unique marker method — see note below
}

type fortress struct {
    *cardBase
}

func NewFortress(id string) *fortress {
    return &fortress{cardBase: newCardBase(id, "Fortress")}
}

func (f *fortress) IsFortressCard() bool { return true }
```

> ⚠️ **Every card interface MUST have at least one unique method beyond `Card`.**
>
> `Card` defines `GetID()`, `Name()`, and observer methods. If your interface only embeds `Card` with nothing else, then **every card type satisfies your interface** (Go structural typing). This breaks two things:
> - `board.HasCardTypeInHand[YourType](p)` returns the first card in hand — any card — not just yours.
> - The `case YourType:` branch in the `processHandCards` type switch catches all remaining cards, including Resource/Gold cards, before `case cards.Resource:` runs.
>
> Add a marker method like `Is<Name>Card() bool` (return `true` in the implementation). See how `Spy` uses `CanSpy()` and `Catapult` uses `Attack(...)` as their unique methods.

- Embed `*cardBase` (provides `GetID()`, `Name()`, observer wiring).
- If the card is a weapon, embed `*weaponBase` instead and implement `DamageAmount()`, `MultiplierFactor()`, `CanBeUsedWith()`.
- If the card is a resource, implement `Value()` and `CanConstruct()`.

**Create `cards/<name>_test.go`** — verify `GetID()`, `Name()`, interface satisfaction.

---

## Step 2 — Add to the deck (`cards/helper.go`)

In `OtherCards()` (or `WarriorCards()` / `ResourceCards()`), add the new card. Follow the same quantity pattern as similar cards:

```go
// base deck
cards.NewFortress("fw1"),

// for games with > 3 players (inside the playerCount > 3 block)
cards.NewFortress("fw2"),
```

---

## Step 3 — Board changes (if the card affects player/castle state)

If the card modifies the **castle** (like Fortress adding protection):

**`board/castle.go`**
- Add field to `castle` struct: `protection cards.Fortress`
- Add method(s) to the `CastleReader` interface: `IsProtected() bool`
- Add method(s) to the `CastleMutator` interface: `SetProtection(f cards.Fortress)`, `ConsumeProtection() cards.Card`
- Implement those methods on `*castle`

If the card modifies the **player** (e.g., a stat or flag on the player), modify `board/player.go` and `ports/player.go` interfaces accordingly.

---

## Step 4 — Game action (`gameactions/`)

**Create `gameactions/gameaction_<name>.go`**

Every game action must implement the `GameAction` interface:

```go
func (a *fortressAction) PlayerName() string { return a.playerName }
func (a *fortressAction) Validate(g Game) error { ... }
func (a *fortressAction) Execute(g Game) (*Result, func() gamestatus.GameStatus, error) { return a.execute(g) }
func (a *fortressAction) NextPhase() types.PhaseType { return types.PhaseTypeEndTurn }
```

Key points:
- `Validate` checks phase, hand contents, and preconditions. Store resolved references (`a.targetPlayer`, `a.fortressCard`) on the struct so `Execute` doesn't need to re-look them up.
- `execute` does the mutation, calls `g.AddHistory(...)`, and returns `&Result{Action: types.LastAction<Name>}`.
- Cache `p.Name()` and target names in local variables before using them — mock expectations count each call.
- `NextPhase()` returns the phase to transition to after this action completes (usually `types.PhaseTypeEndTurn`).

**Create `gameactions/gameaction_<name>_test.go`**

Test all `Validate` error paths and all `Execute` success paths with `gomock`.

---

## Step 5 — LastActionType (`types/lastaction.go`)

Add a constant for the new action:

```go
LastActionFortress       LastActionType = "fortress"
LastActionCatapultBlocked LastActionType = "catapult_blocked"  // if the card can block another action
```

---

## Step 6 — Gamestatus serialization (`gamestatus/`)

### 6a. `gamestatus/cardtype.go`

Add a `CardType` constant and register it in `zeroValueCardTypes`:

```go
CardTypeFortress = CardType{"Fortress", "", "#8B4513"}

// in zeroValueCardTypes map:
"Fortress": CardTypeFortress,
```

### 6b. `gamestatus/handcard.go`

Add a constructor `NewFortressHandCard(...)` that computes `canBeUsed` based on game state, then add a case in `processHandCards`:

```go
case cards.Fortress:
    gs.CurrentPlayerHand = append(gs.CurrentPlayerHand,
        NewFortressHandCard(ct.GetID(), viewer.Castle.IsConstructed,
            game.AllyHasCastleConstructed, action))
```

### 6c. `gamestatus/dto_inputs.go` (if card affects castle/field snapshots)

Add any new fields to `CastleInput` or `FieldInput` that the gamestatus layer needs:

```go
type CastleInput struct {
    ...
    IsProtected bool
}
```

### 6d. `gamestatus/castle.go` (if castle serialization changes)

Add the new field with its JSON tag:

```go
type Castle struct {
    ...
    IsProtected bool `json:"is_protected"`
}

func NewCastle(c CastleInput) Castle {
    return Castle{
        ...
        IsProtected: c.IsProtected,
    }
}
```

---

## Step 7 — `game.go`

### 7a. `extractCastle` (if castle gained new state)

```go
func extractCastle(c board.Castle) gamestatus.CastleInput {
    return gamestatus.CastleInput{
        ...
        IsProtected: c.IsProtected(),
    }
}
```

### 7b. `nextAction` phase-skip logic (if the card unlocks a phase)

The construct phase is only entered if `p.CanConstruct()` is true. If your card is usable in the construct phase but is not a resource, add a check:

```go
if !canConstruct {
    if _, hasFortress := board.HasCardTypeInHand[cards.Fortress](p); hasFortress {
        if p.Castle().IsConstructed() && !p.Castle().IsProtected() {
            canConstruct = true
        }
        // check ally castles for 2v2...
    }
}
```

Do the same for `spy/steal`, `buy`, or any other phase that has a skip condition.

---

## Step 8 — WebSocket layer (`websocket/`)

### 8a. `websocket/dto.go`

If the card affects the castle snapshot, add the new field to `CastleDTO` **and** to `convertCastle`:

```go
type CastleDTO struct {
    ...
    IsProtected bool `json:"is_protected"`
}

func convertCastle(castle gamestatus.Castle) CastleDTO {
    return CastleDTO{
        ...
        IsProtected: castle.IsProtected,
    }
}
```

> ⚠️ Easy to forget — `CastleDTO` is the final wire format. If it's missing a field, the frontend never sees it even if the domain and gamestatus layers are correct.

### 8b. `websocket/message.go`

Add the message type and payload struct:

```go
MsgFortress MessageType = "fortress"

type FortressPayload struct {
    TargetPlayer string `json:"target_player,omitempty"`
}
```

### 8c. `websocket/handlers.go`

Add a handler function:

```go
func (h *Hub) handleFortress(client *Client, payload interface{}) {
    data, _ := json.Marshal(payload)
    var p FortressPayload
    json.Unmarshal(data, &p)
    h.executeGameAction(client, func(g HubGame) (gamestatus.GameStatus, error) {
        return g.ExecuteAction(gameactions.NewFortressAction(client.PlayerName, p.TargetPlayer))
    })
}
```

### 8d. `websocket/hub.go`

Register the handler in the dispatch table:

```go
MsgFortress: func(h *Hub, c *Client, p interface{}) { h.handleFortress(c, p) },
```

---

## Step 9 — Server card config (`server/server.go`)

Add an entry in the `cardConfig` map so the frontend can show the image and description:

```go
"fortress": {
    Description: "Fortify your castle (or an ally's) to block the next catapult attack. The wall is destroyed instead of gold.",
    Image:       "fortress.webp",
},
```

The key must match the `CardType.Name` (lowercase) used in `CardTypeFortress`.

---

## Step 10 — Frontend (`frontend/static/js/game.js`)

### 10a. Card image map

```js
const CARD_IMAGES = {
    ...
    'Fortress': 'fortress.webp',
};
```

### 10b. Card type classification (for CSS)

In `getCardType(type)`:
```js
case 'Fortress': return 'special';
```

### 10c. Click handler (construct phase example)

Add a handler called from `handleConstructPhaseHandClick` (or the relevant phase handler):

```js
if (card.type === 'Fortress') {
    handleFortressPhaseHandClick(card);
    return;
}
```

Implement `handleFortressPhaseHandClick`, and if the card requires target selection, also `showFortressTargetModal` / `showFortressConfirmModal`.

### 10d. WebSocket message send

```js
sendMessage('fortress', { target_player: targetPlayer });
```

### 10e. Phase mapping

```js
const actionToPhase = {
    ...
    'fortress': 'construct',
};
```

### 10f. Rendering state changes (castle, field, etc.)

If the card affects visual state (e.g., adding a shield indicator to the castle):

> ⚠️ There are **two separate castle render functions**: `renderCastle` (player's own castle, uses `getElementById`) and `renderCastleInto` (opponent castles, receives a DOM element). You must update **both**.

```js
// In BOTH renderCastle() and renderCastleInto():
const isProtected = castle.is_protected || false;
if (isProtected) container.classList.add('fortified');
const fortressIndicator = isProtected
    ? '<div class="castle-fortress-indicator">🛡 Fortress Wall</div>'
    : '';
```

### 10g. Change detection for animations

In `detectCastleChanges()`, detect the new state transition:

```js
if (prevCastle.is_protected && !newCastle.is_protected) {
    fortressDestroyed.push({ containerId: 'player-castle' });
}
```

Return it alongside the existing arrays and consume it in the game state handler.

### 10h. Animation function

```js
function showFortressDestroyedAnimation(change) {
    const container = change.container || document.getElementById(change.containerId);
    container.classList.add('fortress-destroyed');
    // add floating text, remove after timeout
}
```

---

## Step 11 — CSS (`frontend/static/css/styles.css`)

Add styles for any new visual states, indicators, and animation keyframes:

```css
.castle.fortified { box-shadow: 0 0 0 3px #8B4513, ...; }
.castle-fortress-indicator { position: absolute; ... }
.fortress-destroyed { animation: fortressShake 0.7s ease-in-out; }
@keyframes fortressShake { ... }
.fortress-destroyed-text { ... }
@keyframes fortressDestroyedFloat { ... }
```

---

## Step 12 — Mocks (if new interfaces were added)

If you added methods to `CastleReader`, `CastleMutator`, `Castle`, `Player`, or any other mocked interface:

```bash
make mocks
```

Or manually update `backend/test/mocks/castle_mocks.go` (and any other mock file) to implement the new interface methods. All mock structs that implement the interface must include the new methods.

For the **card interface marker method** (e.g., `IsFortressCard() bool`), add it directly to the mock without the EXPECT machinery — it's a pure marker with no test behavior:

```go
// IsFortressCard mocks base method.
func (m *MockFortress) IsFortressCard() bool { return true }
```

Do NOT wire it through `m.ctrl.Call(...)` — that would require every test to set up an expectation for a method that's never meaningfully asserted.

---

## Step 13 — Verify

```bash
cd backend && go test ./...
cd backend && go build -o ../server.exe ./cmd/server
```

---

## Checklist summary

| # | File(s) | What to do |
|---|---------|-----------|
| 1 | `cards/<name>.go` + `_test.go` | Define interface + struct |
| 2 | `cards/helper.go` | Add to deck |
| 3 | `board/castle.go` or `board/player.go` | Add state fields + interface methods |
| 4 | `gameactions/gameaction_<name>.go` + `_test.go` | Game action (Validate / Execute / NextPhase) |
| 5 | `types/lastaction.go` | Add `LastAction<Name>` constant |
| 6a | `gamestatus/cardtype.go` | Add `CardType<Name>` + register in map |
| 6b | `gamestatus/handcard.go` | Add `New<Name>HandCard` + case in `processHandCards` |
| 6c | `gamestatus/dto_inputs.go` | Add fields to `CastleInput` / `FieldInput` if needed |
| 6d | `gamestatus/castle.go` | Add field + JSON tag if castle changed |
| 7a | `game.go` — `extractCastle` | Include new field |
| 7b | `game.go` — `nextAction` | Add phase-skip check if card unlocks a phase |
| 8a | `websocket/dto.go` — `CastleDTO` | Add field + `convertCastle` ⚠️ easy to forget |
| 8b | `websocket/message.go` | Add `Msg<Name>` + payload struct |
| 8c | `websocket/handlers.go` | Add `handle<Name>` function |
| 8d | `websocket/hub.go` | Register in dispatch table |
| 9  | `server/server.go` | Add card config entry |
| 10 | `game.js` | Image map, click handler, WS send, rendering, animations |
| 11 | `styles.css` | Visual state styles + animation keyframes |
| 12 | `test/mocks/*.go` | Update mocks for new interface methods |
| 13 | — | `go test ./...` + `go build` |
