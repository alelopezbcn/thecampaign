# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

TheCampaign is a real-time multiplayer turn-based card game. Go backend with WebSocket communication, vanilla JavaScript frontend. Supports 1v1, 2v2, FFA3, and FFA5 game modes.

## Code Style

After modifying any `.go` file, run `gofmt -w <file>` on it before moving on.

**Never edit mock files manually.** If a mock needs to change (e.g. interface changed, new method added), run `make mocks` to regenerate all mocks.

**Split long functions into smaller, well-named methods.** Keep orchestrating functions (e.g. `execute()`) short and readable by extracting sub-steps into focused methods with doc comments (e.g. `snapshotPreKillHP`, `applyBloodlust`, `applyChampionsBounty`).

## Commands

```bash
# Build
cd backend && go build -o ../server.exe ./cmd/server

# Run (default port 8080, configurable via PORT env var)
./server.exe

# Run all tests
make test

# Run a single test
cd backend && go test ./internal/domain/ -run TestName

# Regenerate mocks (uber/mock)
make mocks

# Docker
make up        # docker-compose up --build
make down      # docker-compose down
```

## Testing

**Always write tests for new code.** When adding a new function, method, or game action:
- Add unit tests in the same package (e.g. `gameaction_foo_test.go` alongside `gameaction_foo.go`)
- Cover the happy path and at least one error/edge case
- Run `make test` after writing tests to confirm they pass

## Architecture

**Server-authoritative design**: all game logic runs on the backend. The client sends actions, the server validates and broadcasts the resulting game state to all players.

### Backend (`backend/`)

Entry point: `cmd/server/main.go`

Three layers inside `internal/`:

- **`server/`** - HTTP server. Serves frontend static files and upgrades `/ws` to WebSocket.
- **`websocket/`** - Connection management. `Hub` dispatches messages and manages game rooms. `Client` runs read/write goroutines per connection. `handlers.go` processes each action type. `dto.go` converts domain objects to JSON.
- **`domain/`** - Core game logic. Uses ports/adapters pattern (`ports/` defines interfaces, implementations alongside). Observer pattern for card movement events. Every game action method returns `(GameStatus, error)`. Card types live in `domain/cards/`. Game modes and action types in `domain/types/`.

### Frontend (`frontend/`)

Single-page app: `index.html` + `static/js/game.js` + `static/css/styles.css`. No build step, no frameworks.

- `gameState` global object holds all client state
- WebSocket connection with auto-reconnect (max 20 attempts)
- Multi-step action pattern: user clicks build up `actionState` fields, final click sends complete action to server
- Card images in `static/img/cards/`

### Communication

WebSocket messages are JSON with `type` and `payload` fields. Key client-to-server types: `join_game`, `draw_card`, `attack`, `move_warrior`, `trade`, `buy`, `construct`, `special_power`, `spy`, `steal`, `catapult`, `end_turn`, `skip_phase`. Server responds with `game_state` (full snapshot per player), `error`, `game_started`, `game_ended`.

### Turn Structure

Each turn follows a phase sequence: `draw` -> `attack` -> `spy/steal` -> `buy` -> `construct` -> `endturn`. The `current_action` field in game state drives which UI buttons are enabled and which cards are marked usable.

## Key Patterns

- **Game actions**: All follow `func (g *Game) Action(playerName string, ...) (GameStatus, error)` returning a full state snapshot.
- **Broadcasting**: Hub locks game room mutex, gets status for each player (each sees their own hand, opponents' card counts), sends personalized `game_state` messages.
- **Card usability**: Backend sets `can_be_used` and `can_be_traded` flags per card. Frontend adds `usable`/`unusable` CSS classes based on these flags and the current action phase.
- **Observers**: Cards notify players via interfaces (`CardMovedToPileObserver`, `WarriorMovedToCemeteryObserver`, `CastleCompletionObserver`) for decoupled state updates.
- **Win conditions**: Castle value reaches goal (25 in 1v1/FFA, 30 in 2v2) OR all opponents eliminated.

## Game Mode Differences

- **1v1**: 2 players, castle goal 25
- **2v2**: 4 players in 2 teams, castle goal 30, allies can construct on each other's castles and move warriors to ally fields
- **FFA3/FFA5**: 3 or 5 players free-for-all, castle goal 25

## Help File (`frontend/static/js/help.js`)

**Keep this file in sync whenever game mechanics change.** It is the single source of truth for the in-game reference modal. Update it when you:

- Add, remove, or change a **card type** (warriors, weapons, ambush effects, spy/steal/sabotage/treason, catapult, resurrection, gold)
- Add, remove, or change a **game event** (`domain/gameevents/`)
- Change a **game limit** (hand limit → `board.MaxCardsInHand`, turn time, trade limit, warrior moves, ambush limit)
- Change **castle goals** or **win conditions**
- Add or modify a **game mode**
- Change **turn phase** sequence or rules
- Add or change **special power** behaviour per warrior type

## Phase Badge Tooltips (`frontend/index.html` — `data-tip` on `.pb-phase` spans)

**Keep these tooltips in sync whenever playable cards or mechanics change per phase.** Update them when you:

- Add or remove a **card type** playable in Attack, Spy/Steal, Buy, or Build phase
- Add a new **card action** (e.g. a new Buy-phase card like Mercenary)
- Change **which phase** a card belongs to
