# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with the backend.

## Module

`github.com/alelopezbcn/thecampaign` (Go 1.25.5)

## Adding a New Card Type

1. Define the interface in `internal/domain/ports/` (or use an existing one: `Warrior`, `Weapon`, `Resource`, `Spy`, `Thief`, `Catapult`, `SpecialPower`)
2. Create the struct in `internal/domain/cards/`, embedding the appropriate base:
   - `cardBase` — all cards (ID + observer management)
   - `warriorBase` — warriors (HP, combat, damage tracking)
   - `weaponBase` — weapons (damage amount, type, multiplier)
   - `attackableBase` — anything with health
3. Implement the constructor as `NewXxx(id string) ports.Xxx`
4. Register in dealer (`internal/domain/cards/helper.go`): add to `WarriorsCards()` for warriors or `OtherCards()` for other types. Card counts scale by `playerCount`.
5. Add usability logic in `internal/domain/gamestatus/handcard.go` — create a `NewXxxHandCard()` function that sets `CanBeUsed`, `CanBeUsedOnIDs`, and `CanBeTraded` based on the current action phase and field state.

## Adding a New Game Action

Five touch points, in order:

1. **Message type** (`internal/websocket/message.go`): Add `MsgXxx MessageType = "xxx"` constant and payload struct
2. **Domain method** (`internal/domain/gameactions.go`): Add `func (g *Game) Xxx(playerName string, ...) (GameStatus, error)`. Must validate current player and action phase, call `g.nextAction()` to advance phase, return status.
3. **Handler** (`internal/websocket/handlers.go`): Add `func (h *Hub) handleXxx(client *Client, payload interface{})`. Unmarshal payload, call `h.executeGameAction()` with the domain method.
4. **Routing** (`internal/websocket/hub.go`): Add `case MsgXxx:` to `processMessage()` switch
5. **DTO** (`internal/websocket/dto.go`): Update `ConvertGameStatus()` if the action introduces new state fields

All domain action methods follow this pattern:
```go
func (g *Game) Xxx(playerName string, ...) (status GameStatus, err error) {
    p := g.CurrentPlayer()
    if p.Name() != playerName {
        return status, fmt.Errorf("%s not your turn", playerName)
    }
    if g.currentAction != types.ActionTypeXxx {
        return status, fmt.Errorf("cannot xxx in the %s phase", g.currentAction)
    }
    // ... action logic ...
    g.addToHistory("...", types.CategoryXxx)
    status = g.nextAction(types.ActionTypeNextPhase, func() GameStatus {
        return g.GameStatusProvider.Get(p, g)
    })
    return status, nil
}
```

## Testing

```bash
go test ./...                                    # all tests
go test ./internal/domain/ -run TestGame_Buy     # single test
go test ./internal/domain/ -v                    # verbose
```

- Uses `github.com/stretchr/testify/assert` for assertions
- Uses `go.uber.org/mock/gomock` for mocking
- Generated mocks live in `test/mocks/`, regenerate with `make mocks` from repo root
- Test setup pattern: `ctrl := gomock.NewController(t)` + `defer ctrl.Finish()` + `mocks.NewMockXxx(ctrl)`
- `game_test.go` is the main integration test file; smaller unit tests in `castle_test.go`, `gamestatus/castle_test.go`, etc.

## Observer Interfaces (`internal/domain/ports/observers.go`)

Cards and game components communicate via observers. When adding new card behaviors that affect other entities, implement the relevant observer:

- `CardMovedToPileObserver` — card removed from hand/field
- `WarriorDeadObserver` — warrior HP reaches 0
- `WarriorMovedToCemeteryObserver` — warrior sent to cemetery
- `CastleCompletionObserver` — castle value reaches win threshold
- `FieldWithoutWarriorsObserver` — all warriors removed from a player's field

## Phase Flow

Turn phases progress in order: `draw` -> `attack` -> `spy/steal` -> `buy` -> `construct` -> `endturn`. The `nextAction()` method advances to the next phase. Phases can be skipped with `SkipPhase()`.

## Card Usability Flags

`handcard.go` sets per-card flags sent to the frontend:
- `can_be_used` — whether the card is playable in the current phase (depends on field warriors, castle state, action type)
- `can_be_traded` — whether the card can be included in a trade
- `use_on` — list of target IDs this card can be used on
- `dmg_mult` — map of target ID to damage multiplier (weapons only)

Weapon usability depends on having the matching warrior on field (Sword->Knight/Dragon, Arrow->Archer/Dragon, Poison->Mage/Dragon).
