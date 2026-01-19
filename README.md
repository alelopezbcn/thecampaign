# The Campaign - Multiplayer Card Game 🎮

A real-time multiplayer turn-based card game built with Go WebSockets and vanilla JavaScript.

## 🚀 Quick Start

```bash
./server.exe
```

Open two browser tabs at `http://localhost:8080` and play!

## 📖 Documentation

- **[QUICK_START.md](QUICK_START.md)** - Get playing in 30 seconds
- **[GAME_GUIDE.md](GAME_GUIDE.md)** - Complete gameplay instructions
- **[DEBUG_GUIDE.md](DEBUG_GUIDE.md)** - VS Code debugging setup ✨ NEW!
- **[BUGS_FIXED.md](BUGS_FIXED.md)** - All bugs that were fixed
- **[IMPLEMENTATION_SUMMARY.md](IMPLEMENTATION_SUMMARY.md)** - Technical deep dive

## ✨ Features

- ✅ Real-time multiplayer with WebSockets
- ✅ Turn-based card game mechanics
- ✅ Multiple concurrent games support
- ✅ Beautiful animated UI
- ✅ No framework dependencies (vanilla JS)
- ✅ Full VS Code debugging support

## 🎮 How to Play

1. **Start Server:** `./server.exe`
2. **Open two tabs:** `http://localhost:8080`
3. **Join:** Use same Game ID in both tabs
4. **Select Warriors:** Choose 1-3 warrior cards
5. **Play:** Attack, build castle, defeat opponent!

## 🐛 Debugging in VS Code

Press **F5** to start debugging!

The project is fully configured with two debug configurations:
- **Debug Server** - Port 8080 (default)
- **Debug Server (Alternative Port)** - Port 8081

See [DEBUG_GUIDE.md](DEBUG_GUIDE.md) for complete debugging instructions.

## 🏗️ Project Structure

```
TheCampaign/
├── backend/
│   ├── cmd/server/         # Server entry point
│   ├── internal/
│   │   ├── domain/         # Game logic
│   │   ├── server/         # HTTP server
│   │   └── websocket/      # WebSocket layer
│   └── go.mod
├── frontend/
│   ├── index.html          # Main UI
│   └── static/
│       ├── css/           # Styling
│       └── js/            # Game client
├── .vscode/
│   └── launch.json        # Debug configurations
├── server.exe              # Compiled server
└── start.bat              # Quick start script
```

## 🛠️ Building from Source

```bash
cd backend
go build -o ../server.exe ./cmd/server
cd ..
./server.exe
```

## 🎯 Game Mechanics

### Card Types
- 🔴 **Warriors** - Knight, Archer, Mage, Dragon
- 🔵 **Weapons** - Sword, Arrow, Poison
- 🟡 **Resources** - Gold cards for buying/building
- 🟣 **Special** - Spy, Thief, Catapult

### Win Conditions
- Complete your castle with resources
- Eliminate all enemy warriors

### Actions
- Draw cards
- Move warriors to field
- Attack with weapons
- Use special powers (Heal/Protect/Kill)
- Buy more cards
- Construct castle
- Trade cards

## 🔧 Technology Stack

**Backend:**
- Go 1.25.5
- gorilla/websocket
- Standard library

**Frontend:**
- HTML5
- CSS3 (Grid, Flexbox, Animations)
- Vanilla JavaScript (ES6+)

## 📊 Status

✅ Fully functional and ready to play!

**All bugs fixed:**
1. ✅ JSON field name mismatch
2. ✅ Mutex deadlock in game join
3. ✅ Wrong player perspective

## 🤝 Multiplayer Support

- Unlimited concurrent games
- 2 players per game
- Real-time state synchronization
- Sub-100ms latency

## 📝 Development Commands

```bash
# Build
cd backend && go build -o ../server.exe ./cmd/server

# Run
./server.exe

# Debug in VS Code
# Press F5

# Test
cd backend && go test ./...
```

## 🐛 Troubleshooting

**Server won't start:**
```bash
taskkill /F /IM server.exe
./server.exe
```

**Can't connect:**
- Check firewall settings
- Verify server is running on port 8080
- Open browser DevTools console for errors

**Game won't start:**
- Ensure BOTH players use EXACT same Game ID
- Check both clicked "Join Game"

## 📚 Learn More

- Game logic: [backend/internal/domain/](backend/internal/domain/)
- WebSocket protocol: [backend/internal/websocket/](backend/internal/websocket/)
- Frontend client: [frontend/static/js/game.js](frontend/static/js/game.js)

## 🎉 Have Fun!

Enjoy playing The Campaign! Deploy warriors, build your castle, and defeat your opponent!

---

**Built with ❤️ using Go and Vanilla JavaScript**