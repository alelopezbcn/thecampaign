# Global Event System — Implementation Plan

## Overview
Round-by-round global events that add variability between games. Each round, one event is drawn and affects all players until the next round.

## Events

| Type | Display | Effect | Range |
|------|---------|--------|-------|
| `curse` | Curse | 2 of 3 weapons (Sword/Arrow/Poison) get damage modifier; 1 random weapon is unaffected | [-3,+3] excl. 0 |
| `harvest` | Harvest | Construction resources get value modifier | [-4,+4] excl. 0 |
| `plague` | Plague | Active player's warriors gain/lose HP at turn start | [-3,+3] excl. 0 |
| `abundance` | Abundance | Active player draws 1 extra card during draw phase | fixed |
| `` | Calm | No effect | — |

Pool: 20% each. Active from round 1. Changes at the start of each new round.

---

## Phase 1 — Types & EventHandler interface

**Goal:** Define all data types and the `EventHandler` interface. No game logic changes yet.

### Files to create
- `backend/internal/domain/types/eventtype.go`
  - `EventType` constants: `EventTypeNone`, `EventTypeCurse`, `EventTypeHarvest`, `EventTypePlague`, `EventTypeAbundance`
  - `ActiveEvent` struct with all random parameters (CurseExcludedWeapon, CurseModifier, HarvestModifier, PlagueModifier)
  - Curse affects **2 of 3** weapons: store the one weapon that is NOT affected (`CurseExcludedWeapon`)

- `backend/internal/domain/gameevents/event_handler.go`
  - `EventHandler` interface:
    ```go
    ExtraDrawCards() int
    WeaponDamageModifier(weaponType types.WeaponType) int
    ConstructionValueModifier() int
    TurnStartWarriorHPModifier() int
    Display() (name, description string)
    ```
  - `NewHandler(event types.ActiveEvent) EventHandler` factory

- `backend/internal/domain/gameevents/handler_calm.go` — all methods return 0/""
- `backend/internal/domain/gameevents/handler_curse.go` — returns modifier for the 2 affected weapons (all except `CurseExcludedWeapon`)
- `backend/internal/domain/gameevents/handler_harvest.go` — returns HarvestModifier
- `backend/internal/domain/gameevents/handler_plague.go` — returns PlagueModifier
- `backend/internal/domain/gameevents/handler_abundance.go` — ExtraDrawCards returns 1

### Verification
- `cd backend && go build ./...` — compiles cleanly
- Unit tests for each handler: modifier returned for correct inputs, 0 for unrelated methods

---

## Phase 2 — Game core: draw event, round detection

**Goal:** The game draws a random event at start and changes it each round. Expose it via interface.

### Files to modify
- `backend/internal/domain/game.go`
  - Add field: `currentEvent types.ActiveEvent`
  - Add `drawRandomEvent() types.ActiveEvent` (uses `rand.Intn`, picks random params per type)
  - `NewGame()`: initialize `g.currentEvent = g.drawRandomEvent()`
  - `SwitchTurn()`: at end, detect round wrap (`newTurn <= oldTurn`), call `drawRandomEvent()`, add to history
  - Add method: `EventHandler() gameevents.EventHandler`
  - Add method: `CurrentEvent() types.ActiveEvent`

- `backend/internal/domain/game_interface.go`
  - Add to `GameTurn`: `EventHandler() gameevents.EventHandler` and `CurrentEvent() types.ActiveEvent`

### Verification
- Unit test: after all players complete their turn, `CurrentEvent` changes
- Unit test: `drawRandomEvent` never returns zero for Curse/Harvest/Plague modifiers
- `cd backend && go test ./internal/domain/...`

---

## Phase 3 — Mechanical effects in game actions

**Goal:** Apply the event effects at the right moment in each game action.

### Files to modify

**`gameaction_attack.go`** — Curse effect
```go
mod := g.EventHandler().WeaponDamageModifier(a.weapon.WeaponType())
// if mod != 0, wrap weapon with eventWeaponWrapper{original, effectiveDmg}
```
Add `eventWeaponWrapper` struct implementing `WeaponCard` (overrides only `DamageAmount()`).

**`gameaction_draw.go`** — Plague effect + Abundance effect
```go
// Plague: apply TurnStartWarriorHPModifier to active player's warriors (min HP = 1)
// Abundance: add ExtraDrawCards() to draw count
```

**`backend/internal/domain/board/castle.go`** — Harvest effect
- Change `Construct(card, valueModifier int) error` signature
- Internal: `value := max(1, card.Value() + valueModifier)`

**`backend/internal/domain/ports/`** — Update Castle interface signature.

**`gameaction_construct.go`** — pass `g.EventHandler().ConstructionValueModifier()` to `castle.Construct`.

All other callers of `castle.Construct` pass `0` as modifier.

### Verification
- Unit tests for each event effect:
  - Curse: Arrow with modifier -2 → effective damage reduced
  - Harvest: resource value 4 with modifier -3 → castle gets 1
  - Plague: warrior with 10 HP + modifier -5 → stays at 1 (doesn't die)
  - Plague: modifier +2 → warrior gains HP (up to max health)
  - Abundance: draw phase draws N+1 cards
- `cd backend && go test ./...`

---

## Phase 4 — GameStatus: expose event to frontend

**Goal:** Include the active event in every `game_state` message sent to clients.

### Files to modify
- `backend/internal/domain/gamestatus/gamestatus.go`
  - Add to `GameStatus` struct:
    ```go
    CurrentEvent            string `json:"current_event"`
    CurrentEventDisplay     string `json:"current_event_display"`
    CurrentEventDescription string `json:"current_event_description"`
    ```
- Wherever `GameStatus` is constructed (BuildInput or equivalent):
  - Accept `CurrentEvent types.ActiveEvent`
  - Populate the three fields via `gameevents.NewHandler(event).Display()`

### Verification
- Print or log the JSON game_state message — confirm `current_event`, `current_event_display`, `current_event_description` appear
- `cd backend && go test ./...`

---

## Phase 5 — Frontend: banner, turn modal, toast

**Goal:** Display the active event prominently to all players.

### Files to modify

**`frontend/index.html`**
- Add `<div id="event-banner">` with: icon, event name, short description, `(?)` tooltip trigger
- Add `<div id="event-turn-modal">` (hidden, shown at turn start)
- Add `<div id="event-toast">` (hidden, non-blocking corner notification)

**`frontend/static/css/styles.css`**
- Banner always visible, colored by `data-event` attribute
- Turn modal: large centered overlay
- Toast: top-right corner, slides in/out

**`frontend/static/js/game.js`** — in `handleGameState()`:

1. `renderEventBanner(gs)` — updates banner with name + description + tooltip (called every state update)
2. `showEventTurnModal(gs)` — triggered when `isNowYourTurn` transitions to `true` (turn start)
3. `showEventChangeToast(gs)` — triggered when `current_event` changes between two consecutive states

### Verification
- Start a game, verify banner shows event with tooltip on hover
- When it's your turn, verify modal appears with event description
- Complete all turns (full round), verify toast appears for all players when event changes

---

## Summary of all new/modified files

### New files
- `backend/internal/domain/types/eventtype.go`
- `backend/internal/domain/gameevents/event_handler.go`
- `backend/internal/domain/gameevents/handler_calm.go`
- `backend/internal/domain/gameevents/handler_curse.go`
- `backend/internal/domain/gameevents/handler_harvest.go`
- `backend/internal/domain/gameevents/handler_plague.go`
- `backend/internal/domain/gameevents/handler_abundance.go`

### Modified files
- `backend/internal/domain/game.go`
- `backend/internal/domain/game_interface.go`
- `backend/internal/domain/gamestatus/gamestatus.go`
- `backend/internal/domain/gamestatus/` (BuildInput / construction function)
- `backend/internal/domain/gameactions/gameaction_attack.go`
- `backend/internal/domain/gameactions/gameaction_draw.go`
- `backend/internal/domain/gameactions/gameaction_construct.go`
- `backend/internal/domain/board/castle.go`
- `backend/internal/domain/ports/` (Castle interface)
- `frontend/index.html`
- `frontend/static/css/styles.css`
- `frontend/static/js/game.js`
