# Command Pattern for Game Actions — Implementation Plan

## Goal

Refactor the 13 action methods on `Game` (640 lines in `gameactions.go`) into standalone command structs that implement a `GameAction` interface. `Game` becomes an orchestrator that dispatches actions, not the implementor of every action.

**Why**: New card types (Harpoon, Wall, Plague, etc.) should each be a new file, not more methods bolted onto `Game`. Each action becomes independently testable and self-contained.

---

## Current State

- **`gameactions.go`** — 13 methods on `Game`, each following the same pattern:
  1. Validate it's the player's turn (`CurrentPlayer().Name() != playerName`)
  2. Validate the current phase (`g.currentAction != expected`)
  3. Look up player/cards, validate types
  4. Execute the logic (attack, buy, construct, etc.)
  5. Write to `g.lastResult` (action-specific fields)
  6. Add to `g.history`
  7. Call `g.nextAction(nextPhase, statusFn)` to advance the phase
  8. Return `GameStatus`

- **`handlers.go`** — Each handler unmarshals a payload struct, then calls the corresponding `Game` method inside `executeGameAction`.

- **`ActionResult`** — Flat struct with fields for all possible actions (attack fields, steal fields, spy fields, etc.). Only a subset is populated per action.

---

## Design

### The Interface

```go
// File: backend/internal/domain/action.go

type GameAction interface {
    Validate(g *Game) error
    Execute(g *Game) (*ActionResult, func() gamestatus.GameStatus, error)
    NextPhase() types.ActionType
}
```

- **`Validate`** — All precondition checks (turn, phase, card lookups, type assertions). Returns descriptive error or nil.
- **`Execute`** — The actual game mutation. Returns an `ActionResult` (pure data for animations/notifications), a `statusFn` closure (controls `Get` vs `GetWithModal` and which extra args to pass), and an error. Must only be called after `Validate` succeeds.
- **`NextPhase`** — Returns which phase to advance to after this action. Every action always goes through `nextAction` — there are no special cases. Actions that don't conceptually "advance" (MoveWarrior, Trade) simply return their current phase (e.g., `ActionTypeAttack`, `ActionTypeBuy`), and `nextAction` re-evaluates and stays on that phase.

### The Dispatcher

```go
// File: backend/internal/domain/action.go

func (g *Game) ExecuteAction(action GameAction) (GameStatus, error) {
    if err := action.Validate(g); err != nil {
        return GameStatus{}, err
    }

    result, statusFn, err := action.Execute(g)
    if err != nil {
        return GameStatus{}, err
    }

    g.lastResult = *result

    return g.nextAction(action.NextPhase(), statusFn), nil
}
```

No conditionals, no special cases. Every action flows through `nextAction`. The dispatcher is 100% generic.

### How Each Action Controls Its Status

`Execute` returns the `statusFn` closure alongside the `ActionResult`. This keeps `ActionResult` as a pure data struct (no function fields) while letting each action control its own status generation:

```go
// Attack — simple Get:
statusFn := func() gamestatus.GameStatus {
    return g.GameStatusProvider.Get(p, g)
}

// Buy — passes newCards for animation:
statusFn := func() gamestatus.GameStatus {
    return g.GameStatusProvider.Get(p, g, cards...)
}

// Spy — uses GetWithModal:
statusFn := func() gamestatus.GameStatus {
    return g.GameStatusProvider.GetWithModal(p, g, spiedCards)
}
```

The closure naturally captures the local variables from `Execute`, so everything stays in scope.

---

## Actions to Extract (13 total)

Each becomes its own file under `backend/internal/domain/actions/`. The file contains the action struct + `Validate` + `Execute` + `NextPhase`.

| # | Action | File | Phase Check | Next Phase | Special Notes |
|---|--------|------|-------------|------------|---------------|
| 1 | DrawCard | `action_draw.go` | (none — draw is first phase) | `Attack` | Hand limit → skip to Attack without error. Passes `newCards` to `StatusFn`. |
| 2 | Attack | `action_attack.go` | `Attack` | `SpySteal` | Sets `AttackWeaponID/TargetID/TargetPlayer` on result. |
| 3 | SpecialPower | `action_specialpower.go` | `Attack` | `SpySteal` | Complex target resolution: own field → ally fields → enemy fields. Side validation (archer=enemy, knight/mage=ally). |
| 4 | Catapult | `action_catapult.go` | `Attack` | `SpySteal` | Uses `p.Catapult()` and attacks castle. |
| 5 | MoveWarrior | `action_movewarrior.go` | (during Attack phase, but no phase check — uses `hasMovedWarrior` guard) | `Attack` (stays on current phase) | Returns `NextPhase() = ActionTypeAttack`. `nextAction` re-evaluates, sees player can still act, stays on Attack phase. Sets `hasMovedWarrior = true`. |
| 6 | Trade | `action_trade.go` | (during Buy phase, but no explicit phase check — uses `hasTraded` guard) | `Buy` (stays on current phase) | Returns `NextPhase() = ActionTypeBuy`. `nextAction` re-evaluates, sees player can still buy/trade, stays on Buy phase. Sets `hasTraded = true`. |
| 7 | Spy | `action_spy.go` | `SpySteal` | `Buy` | Two options (deck/player). Uses `GetWithModal` for status. Sets `Spy` field on result. |
| 8 | Steal | `action_steal.go` | `SpySteal` | `Buy` | Uses `GetWithModal` for status. Sets `StolenFrom/StolenCard` on result. |
| 9 | Buy | `action_buy.go` | `Buy` | `Construct` | Draws cards, passes `newCards` to status. Reverts on hand limit error. |
| 10 | Construct | `action_construct.go` | `Construct` | `EndTurn` | Two paths: own castle vs ally castle (2v2). |
| 11 | SkipPhase | `action_skip.go` | Attack/SpySteal/Buy/Construct | (dynamic — depends on current phase) | `NextPhase` must read `g.currentAction` to determine where to skip to. |
| 12 | EndTurn | `action_endturn.go` | (none) | `DrawCard` | Calls `g.switchTurn()`. Special: `expired` flag affects history message. |
| 13 | AutoMoveWarrior | — | — | — | **Do NOT extract.** This is a setup-only method (returns `error`, not `GameStatus`). Keep as-is on `Game`. |

### MoveWarrior and Trade — No Special Cases

These two actions don't conceptually "advance" the phase, but they still go through `nextAction` like everything else. They simply return their **current phase** as `NextPhase()`:

```go
func (a *MoveWarriorAction) NextPhase() types.ActionType { return types.ActionTypeAttack }
func (a *TradeAction) NextPhase() types.ActionType        { return types.ActionTypeBuy }
```

`nextAction` re-evaluates the phase conditions. After moving a warrior, the player might still be able to attack — `nextAction(ActionTypeAttack, ...)` checks and stays on Attack. After trading, the player might still be able to buy — `nextAction(ActionTypeBuy, ...)` checks and stays on Buy. If they can't, it naturally falls through to the next phase.

This means the dispatcher has zero special cases — every action is treated identically.

---

## Task Breakdown

### Phase 1: Foundation (no behavior change)

**Task 1.1 — Create the `GameAction` interface and `ExecuteAction` dispatcher**
- Create file `backend/internal/domain/action.go`
- Define the `GameAction` interface (`Validate`, `Execute`, `NextPhase`)
- Implement `ExecuteAction` on `Game` — no conditionals, just: validate → execute → set lastResult → nextAction
- No existing code changes yet — this is additive only

For each task below, the process is the same:
1. Create the action file (`action_*.go`) with the action struct implementing `GameAction`
2. Delete the old `Game.Method()` from `gameactions.go`
3. Update tests in `game_test.go` to use `g.ExecuteAction(NewXxxAction(...))`
4. Update the handler in `handlers.go` to use `g.ExecuteAction(NewXxxAction(...))`
5. Run the relevant tests

**Task 1.2 — Extract `SkipPhase`**
- Create `backend/internal/domain/action_skip.go` with `SkipPhaseAction`
- Dynamic `NextPhase` — computed during `Validate` from `g.currentAction`
- Delete `Game.SkipPhase` from `gameactions.go`
- Update tests: `go test ./internal/domain/ -run TestGame_SkipPhase`
- This validates the entire pattern works end-to-end before touching anything else

**Task 1.3 — Extract `DrawCard`**
- Create `backend/internal/domain/action_draw.go`
- Handle the hand-limit edge case (skip to Attack without error)
- Run tests: `go test ./internal/domain/ -run TestGame_DrawCard`

**Task 1.4 — Extract `EndTurn`**
- Create `backend/internal/domain/action_endturn.go`
- Handle `expired` flag and `switchTurn()` call
- **Note**: `handleEndTurn` in hub stays special — it still calls `autoDrawAndBroadcast` + `startTurnTimer` after `ExecuteAction`
- Run tests: `go test ./internal/domain/ -run TestGame_EndTurn`

### Phase 2: Combat actions

**Task 2.1 — Extract `Attack`**
- Create `backend/internal/domain/action_attack.go`
- Run tests: `go test ./internal/domain/ -run TestGame_Attack` and `TestAttacks`

**Task 2.2 — Extract `SpecialPower`**
- Create `backend/internal/domain/action_specialpower.go`
- Most complex validation (target resolution across own/ally/enemy fields + side validation per warrior type)
- Run tests: `go test ./internal/domain/ -run TestGame_SpecialPower`

**Task 2.3 — Extract `Catapult`**
- Create `backend/internal/domain/action_catapult.go`
- Run tests: `go test ./internal/domain/ -run TestGame_Catapult`

### Phase 3: Spy/Steal actions

**Task 3.1 — Extract `Spy`**
- Create `backend/internal/domain/action_spy.go`
- Handle two options (deck reveal vs player hand) and `GetWithModal` status
- Run tests: `go test ./internal/domain/ -run TestGame_Spy`

**Task 3.2 — Extract `Steal`**
- Create `backend/internal/domain/action_steal.go`
- Handle `GetWithModal` status and `StolenFrom/StolenCard` result fields
- Run tests: `go test ./internal/domain/ -run TestGame_Steal`

### Phase 4: Economy actions

**Task 4.1 — Extract `Buy`**
- Create `backend/internal/domain/action_buy.go`
- Handle the revert-on-hand-limit logic (give card back if draw fails)
- Run tests: `go test ./internal/domain/ -run TestGame_Buy`

**Task 4.2 — Extract `Construct`**
- Create `backend/internal/domain/action_construct.go`
- Handle own castle vs ally castle (2v2) paths
- Run tests: `go test ./internal/domain/ -run TestGame_Construct`

### Phase 5: Stay-on-phase actions

**Task 5.1 — Extract `MoveWarrior`**
- Create `backend/internal/domain/action_movewarrior.go`
- `NextPhase()` returns `ActionTypeAttack` (current phase) — `nextAction` re-evaluates and stays
- Handle own field vs ally field (2v2) paths
- Run tests: `go test ./internal/domain/ -run TestGame_MoveWarrior`

**Task 5.2 — Extract `Trade`**
- Create `backend/internal/domain/action_trade.go`
- `NextPhase()` returns `ActionTypeBuy` (current phase) — `nextAction` re-evaluates and stays
- Run tests: `go test ./internal/domain/ -run TestGame_Trade`

### Phase 6: Cleanup & validation

**Task 6.1 — Delete `gameactions.go`**
- By this point all actions are extracted, all tests and handlers updated — `gameactions.go` should be empty
- Delete the file

**Task 6.2 — Run full test suite**
- `cd backend && go test ./...`
- Fix any failures

**Task 6.3 — Verify with a smoke test**
- Start the server, play a full game through the UI
- Verify all actions work: draw, attack, special power, catapult, spy, steal, buy, construct, move warrior, trade, skip, end turn

---

## File Structure (after refactoring)

```
backend/internal/domain/
├── action.go                    # GameAction interface + ExecuteAction + ActionResult
├── action_draw.go               # DrawCardAction
├── action_attack.go             # AttackAction
├── action_specialpower.go       # SpecialPowerAction
├── action_catapult.go           # CatapultAction
├── action_movewarrior.go        # MoveWarriorAction
├── action_trade.go              # TradeAction
├── action_spy.go                # SpyAction
├── action_steal.go              # StealAction
├── action_buy.go                # BuyAction
├── action_construct.go          # ConstructAction
├── action_skip.go               # SkipPhaseAction
├── action_endturn.go            # EndTurnAction
├── game.go                      # Game struct + nextAction + helpers (unchanged)
├── game_test.go                 # Tests use g.ExecuteAction(NewXxxAction(...))
└── ...                          # gameactions.go is deleted
```

---

## Risks and Mitigations

| Risk | Mitigation |
|------|-----------|
| **MoveWarrior/Trade don't advance the phase** | They return their current phase as `NextPhase()` — `nextAction` re-evaluates and stays. No special cases in the dispatcher. |
| **DrawCard's hand-limit skip is an edge case** | Handled within DrawCardAction's `Execute` — returns appropriate statusFn and nextPhase |
| **Spy/Steal need `GetWithModal` not `Get`** | Each action's `Execute` returns its own `statusFn` closure — dispatcher is agnostic to which status method is used |
| **4000+ lines of existing tests** | Update tests per action as you extract — mechanical replacement of `g.Method(args)` → `g.ExecuteAction(NewXxxAction(args))` |
| **EndTurn has side effects beyond the action (auto-draw, timer)** | `EndTurn` as an action only handles turn logic. Hub still calls `autoDrawAndBroadcast` + `startTurnTimer` after. |
| **Actions need access to Game internals (unexported fields)** | Action structs live in the same `domain` package — they have access to all unexported fields |

---

## Order of Implementation

```
1.1  Interface + dispatcher         ← Foundation, no behavior change
1.2  SkipPhase                      ← Simplest action, validates the pattern
1.3  DrawCard                       ← Tests the edge case handling
1.4  EndTurn                        ← Tests switchTurn integration
2.1  Attack                         ← First combat action
2.2  SpecialPower                   ← Most complex validation
2.3  Catapult                       ← Simple combat variant
3.1  Spy                            ← Tests GetWithModal path
3.2  Steal                          ← Tests GetWithModal + stolen card
4.1  Buy                            ← Tests revert logic
4.2  Construct                      ← Tests 2v2 ally path
5.1  MoveWarrior                    ← Tests stay-on-phase path (NextPhase = current phase)
5.2  Trade                          ← Tests stay-on-phase path (NextPhase = current phase)
6.1  Delete gameactions.go          ← Should be empty by now
6.2  Full test suite                ← Validate everything
6.3  Smoke test                     ← Manual verification
```

For each task (1.2–5.2): create action file, delete old Game method, update tests + handler, run tests.
After each phase, run `go test ./...` to catch regressions.
