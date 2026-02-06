# The Campaign - Multiplayer Card Game

A real-time multiplayer card game built with Go WebSockets and vanilla JavaScript.

## Quick Start

### Running the Server

1. Build the server (already done):
   ```bash
   cd backend
   go build -o ../server.exe ./cmd/server
   ```

2. Start the server from the project root:
   ```bash
   ./server.exe
   ```

3. The server will start on `http://localhost:8080`

### Playing the Game

1. **Open two browser tabs** (or two different browsers) and navigate to `http://localhost:8080`

2. **In the first tab (Player 1):**
   - Enter your name (e.g., "Alice")
   - Enter a Game ID (e.g., "game123")
   - Click "Join Game"
   - You'll see "Waiting for opponent..."

3. **In the second tab (Player 2):**
   - Enter your name (e.g., "Bob")
   - Enter the **same Game ID** ("game123")
   - Click "Join Game"
   - The game will start!

## Game Flow

### 1. Initial Setup Phase
- Each player selects 1-3 warrior cards from their hand to place on the field
- Click cards to select them (they'll highlight)
- Click "Confirm Selection" when ready

### 2. Main Game Loop

**On Your Turn:**
1. **Draw a Card** - Automatically draw one card at the start of your turn
2. **Perform Actions** (you can do multiple actions per turn):
   - **Move Warrior to Field** - Deploy warriors from your hand
   - **Attack** - Select: Your warrior → Enemy target → Weapon card
   - **Special Power** - Use special abilities (Heal, Protect, Instant Kill)
   - **Buy** - Trade gold cards for more cards
   - **Construct** - Build your castle or add resources to it
   - **Trade** - Exchange 3 cards for 1 new card
   - **Spy** - Peek at deck or enemy hand
   - **Steal** - Take a card from enemy hand (requires Thief)
   - **Catapult** - Attack enemy castle resources
3. **End Turn** - Pass control to your opponent

## Game Mechanics

### Card Types

- **Warriors** (Red) - Knight, Archer, Mage, Dragon
  - Each has hit points (HP)
  - Can attack enemy warriors
  - Must be protected when HP reaches 0

- **Weapons** (Cyan) - Sword, Arrow, Poison
  - Used to attack with warriors
  - Different warriors can use different weapons:
    - Knight uses Swords
    - Archer uses Arrows
    - Mage uses Poison
    - Dragon can use all weapons

- **Resources** (Yellow) - Gold cards
  - Value 1: Can construct castle
  - Value 2+: Can buy cards or construct

- **Special Cards** (Purple)
  - Spy: Reveal enemy cards or deck
  - Thief: Steal from enemy hand
  - Catapult: Attack enemy castle
  - Special Power: Heal/Protect/Kill warriors

### Win Conditions

**You WIN if:**
- You complete your castle (enough resource cards)
- Your enemy has no warriors left on the field

**You LOSE if:**
- Your enemy completes their castle first
- You have no warriors left on the field

## UI Guide

### Game Board Layout

```
┌─────────────────────────────────────┐
│ Enemy Castle    │  Enemy Field      │  ← Enemy Area
├─────────────────┴───────────────────┤
│ ═════ BATTLEFIELD ═════             │  ← Divider
├─────────────────┬───────────────────┤
│ Your Field      │  Your Castle      │  ← Your Area
├─────────────────────────────────────┤
│ Your Hand (Cards)                   │  ← Your Hand
├─────────────────────────────────────┤
│ Action Buttons                      │  ← Actions
└─────────────────────────────────────┘
```

### Turn Indicator
- **Green "YOUR TURN"** - You can perform actions
- **Gray "ENEMY TURN"** - Wait for opponent

### Card Selection
- Cards you can interact with will **highlight on hover**
- **Click** to select cards
- **Selected cards** show a blue border and lift up
- Follow the **action prompt** at the bottom for multi-step actions

## Tips & Strategy

1. **Protect Your Warriors** - Keep at least one warrior on the field
2. **Resource Management** - Balance between buying cards and building your castle
3. **Card Advantage** - Use Trade action to cycle through your deck
4. **Special Cards** - Spy and Thief can give you strategic information
5. **Dragon Power** - Dragons can use any weapon type
6. **Castle Race** - Sometimes it's better to rush castle completion

## Technical Details

### Architecture
- **Backend**: Go with gorilla/websocket
- **Frontend**: Vanilla JavaScript with modern CSS
- **Communication**: Real-time WebSocket messages
- **State Management**: Server-authoritative game state

### WebSocket Messages
The game uses JSON messages for real-time communication:
- `join_game` - Join a game room
- `set_initial_warriors` - Setup phase
- `draw_card`, `attack`, `move_warrior`, etc. - Game actions
- `game_state` - Server broadcasts updated state
- `end_turn` - Pass turn to opponent

### Development Mode
The server allows all CORS origins for development. For production:
1. Update the `CheckOrigin` function in `backend/internal/server/server.go`
2. Use HTTPS with WSS protocol
3. Add proper authentication

## Troubleshooting

**Connection Issues:**
- Ensure the server is running
- Check browser console for WebSocket errors
- Try refreshing the page

**Game Not Starting:**
- Make sure both players use the **exact same Game ID**
- Game IDs are case-sensitive

**Cards Not Appearing:**
- Check browser console for JSON parsing errors
- Refresh both browser tabs

**Actions Not Working:**
- Verify it's your turn (green indicator)
- Some actions require specific cards in hand
- Read the action prompt for requirements

## Future Enhancements

Potential features to add:
- Player authentication
- Game history/replay
- Multiple concurrent games per browser
- Card animations and sound effects
- Mobile-responsive design
- AI opponent for single-player
- Tournament mode

---

Enjoy playing The Campaign! 🎮🃏
