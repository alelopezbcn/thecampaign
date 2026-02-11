# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with the frontend.

## Stack

Vanilla HTML/CSS/JS. No build step, no frameworks, no dependencies. Single-page app served by the Go backend.

## File Overview

- `index.html` — all screens (join, waiting, game, game-over) and modal markup
- `static/js/game.js` (~2,800 lines) — all game logic, WebSocket handling, rendering
- `static/css/styles.css` (~3,000 lines) — all styling
- `static/img/cards/` — card artwork (webp format)

## CSS Section Map (`styles.css`)

| Section | Lines | Notes |
|---------|-------|-------|
| Base & screens | 1-112 | Reset, body, screen visibility |
| Buttons | 151-219 | `.btn-primary`, `.btn-danger`, `.btn-action`, `.btn-skip` |
| Game layout | 515-576 | Main game screen, `padding-left: 250px` for sidebar |
| Phase tracker | 622-802 | Phase pills, current/completed/skipped states |
| Player list panel | 802-900 | Right sidebar |
| History panel | 900-990 | Game event log |
| Opponent boards | 990-1270 | Opponent layout, eliminated states |
| Player area (field + castle) | 1274-1418 | Grid `3fr 1fr`, castle image/progress, field container |
| Cemetery & discard pile | 1419-1512 | Graveyard and discard styling |
| Deck & card backs | 1513-1656 | Deck display, card-back design |
| Hand area | 1657-1684 | Player hand container |
| Cards | 1685-1955 | Base card, hover, selected, disabled, unusable, usable, valid-target, newly-drawn, damage feedback |
| Card images & info | 2174-2233 | `.card.has-image`, `.card-image`, `.card-info`, stat badges |
| Action panel | 2234-2259 | Action buttons and prompt |
| Responsive | 2297-2336 | Breakpoints at 1024px and 768px |
| Modals | 2338-2877 | Game modal, action confirm, face-down cards, catapult, game-over |
| Error toast | 2877-2976 | Error notification styling |

## Color Conventions

- Warriors: `#ff6b6b` (red)
- Weapons: `#00d4ff` / `#4ecdc4` (cyan/teal)
- Resources: `#ffd700` / `#ffd93d` (gold/yellow)
- Special cards: `#c74dff` (purple)
- Success/usable: `#4cd964` (green)
- Constructed castle border: `rgba(76, 217, 100, 0.4)`

## JavaScript Architecture (`game.js`)

### Global State

```javascript
gameState = {
    playerName, gameID, gameMode,    // session identity
    isYourTurn, currentState,        // from server
    selectedCards: [],                // multi-select (trade)
    currentAction: null,             // active action type string
    actionState: {                   // multi-step action tracking
        type, weaponId, userId, targetId, warriorId, targetPlayer
    }
}
```

### Action State Machine

Multi-step actions build up `actionState` fields across multiple clicks:

1. `startAction(type)` — sets `currentAction`, re-renders board (applies usable/unusable classes)
2. User clicks cards — `handleCardClick(cardID, cardType, context, card)` routes based on `currentAction` and `context`
3. Each click fills an `actionState` field and may show a confirmation modal
4. `sendAction(type, payload)` — sends completed action to server via WebSocket
5. `resetActionState()` — clears all state, removes visual selections
6. `cancelAction()` — resets state AND re-renders board (important for restoring usable/unusable classes)

### Card Rendering

`createCardElement(card, context)` creates card DOM elements. The `context` parameter controls interactivity:

| Context | Where | Click behavior |
|---------|-------|---------------|
| `'player-hand'` | Player's hand | Full action handling based on current phase |
| `'player-field'` | Player's field | Only during special power ally selection |
| `'opponent-field:NAME'` | Opponent fields | Target selection during attack/special power |
| `'cemetery'` | Cemetery display | No click handler |
| `'discard-pile'` | Discard pile | No click handler |

Cards without a stat badge (SpecialPower, Spy, Thief) have a shorter `.card-info` bar — the `min-height: 24px` on `.card-info` keeps all cards the same height.

### Phase-Specific Card Click Handlers

Each action phase has a dedicated handler called from `handleCardClick()`:

- `handleAttackPhaseHandClick()` — select weapon, then target opponent, then target warrior
- `handleSpyStealPhaseHandClick()` — select spy/thief card, choose target
- `handleBuyPhaseHandClick()` — select resource card to spend
- `handleConstructPhaseHandClick()` — select resource/weapon for castle

### Modal Patterns

Two modal systems:

**`showGameModal(title, subtitle, contentHtml)`** — general decision modals (target selection, spy options, steal choices). Content is arbitrary HTML. Close button resets action state.

**`showActionConfirmModal({ title, cardsHtml, description, onConfirm })`** — Yes/No confirmation. Stores `onConfirm` callback, invoked on Yes. No resets action state.

### WebSocket

`connectWebSocket()` establishes connection to `/ws`. Auto-reconnects with exponential backoff (max 20 attempts, capped at 10s delay). `handleMessage()` routes on `message.type`:

- `game_state` → `handleGameState()` — main update, re-renders entire board
- `game_started` → show game screen
- `player_joined` → update waiting room
- `error` → show error toast

### Key Constants

- `CARD_IMAGES` — maps card sub_type to image filename (e.g., `'knight': 'knight.webp'`)
- `MAX_RECONNECT_ATTEMPTS` = 20
- Castle goals: 25 (1v1/FFA) or 30 (2v2), derived from `gameState.gameMode`
- Turn timer: server-provided via `turn_time_limit_secs`
