# Debugging Guide - The Campaign Server

## VS Code Debugging Setup ✅

The project is now configured for debugging in VS Code!

## Available Debug Configurations

### 1. Debug Server (Default)
- **Port:** 8080
- **Working Directory:** Project root
- **Verbose logging:** Enabled

### 2. Debug Server (Alternative Port)
- **Port:** 8081
- **Working Directory:** Project root
- **Use this if:** Port 8080 is already in use

## How to Debug

### Method 1: Using VS Code UI

1. **Open the file you want to debug**
   - Example: [backend/cmd/server/main.go](backend/cmd/server/main.go)
   - Or: [backend/internal/websocket/hub.go](backend/internal/websocket/hub.go)

2. **Set Breakpoints**
   - Click in the left margin next to line numbers
   - Red dot appears = breakpoint set ●

3. **Start Debugging**
   - Press `F5` or
   - Click "Run and Debug" in sidebar (Ctrl+Shift+D)
   - Select "Debug Server" from dropdown
   - Click green play button ▶

4. **Debug Controls**
   - `F5` - Continue
   - `F10` - Step Over
   - `F11` - Step Into
   - `Shift+F11` - Step Out
   - `Shift+F5` - Stop Debugging

### Method 2: Using Command Palette

1. `Ctrl+Shift+P` → Type "Debug: Start Debugging"
2. Select "Debug Server"

## Key Files to Debug

### Server Initialization
- [backend/cmd/server/main.go](backend/cmd/server/main.go) - Entry point
- [backend/internal/server/server.go](backend/internal/server/server.go) - HTTP server setup

### WebSocket Handling
- [backend/internal/websocket/hub.go](backend/internal/websocket/hub.go) - Game room management
  - `handleJoinGame` - Player joining (line ~128)
  - `sendGameState` - Broadcasting state (line ~219)
  - `Run` - Main hub loop (line ~53)

- [backend/internal/websocket/client.go](backend/internal/websocket/client.go) - Client connections
  - `readPump` - Reading messages (line ~43)
  - `writePump` - Writing messages (line ~77)

- [backend/internal/websocket/handlers.go](backend/internal/websocket/handlers.go) - Game actions
  - All game action handlers (Draw, Attack, Move, etc.)

### Game Logic
- [backend/internal/domain/game.go](backend/internal/domain/game.go) - Core game rules
  - All public methods for game operations

## Useful Breakpoints

### For Debugging Connection Issues
```go
// backend/internal/server/server.go
func (s *Server) handleWebSocket(w http.ResponseWriter, r *http.Request) {
    conn, err := upgrader.Upgrade(w, r, nil)  // ← Set breakpoint here
    // ...
}
```

### For Debugging Game Join
```go
// backend/internal/websocket/hub.go
func (h *Hub) handleJoinGame(client *Client, payload interface{}) {
    // Line 141 - Set breakpoint at start
    h.mutex.Lock()
    gameID := joinPayload.GameID
    playerName := joinPayload.PlayerName  // ← Inspect values here
    // ...
}
```

### For Debugging Game State
```go
// backend/internal/websocket/hub.go
func (h *Hub) sendGameState(gameID string) {
    // Line 232 - Check who is current player
    currentPlayer, enemyPlayer := room.Game.WhoIsCurrent()  // ← Breakpoint
    // ...
}
```

### For Debugging Game Actions
```go
// backend/internal/websocket/handlers.go
func (h *Hub) handleAttack(client *Client, payload interface{}) {
    // Set breakpoint to see attack details
    var p AttackPayload
    // ...
}
```

## Debug Variables to Watch

Add these to the **Watch** panel (right-click → Add to Watch):

```
client.PlayerName
client.GameID
room.Players
len(room.Players)
currentPlayerName
gameState.isYourTurn
payload
```

## Debug Console Commands

When paused at a breakpoint, you can use the Debug Console:

```go
// Print variable
p client.PlayerName

// Print struct
p client

// Print map
p room.Players

// Check length
p len(room.Players)

// Call function
p currentPlayer.Name()
```

## Debugging Tips

### 1. Log Everything Important
The server already has logging, but you can add more:

```go
log.Printf("DEBUG: playerName=%s, gameID=%s", playerName, gameID)
log.Printf("DEBUG: Game state sent to %s (isYourTurn=%v)", playerName, isYourTurn)
```

### 2. Check WebSocket Messages
Open browser DevTools (F12) → Console → Look for:
```javascript
console.log('Received message:', message);
console.log('Sending message:', message);
```

### 3. Use Conditional Breakpoints
Right-click breakpoint → Edit Breakpoint → Add condition:
```
playerName == "John"
len(room.Players) == 2
gameID == "test123"
```

### 4. Check Mutex Deadlocks
If server hangs:
1. Pause debugger
2. Check goroutines panel
3. Look for goroutines waiting on mutex.Lock()

## Common Debugging Scenarios

### Scenario 1: Players Can't Join
**Breakpoint:** `handleJoinGame` line 141
**Check:**
- Is WebSocket connection established?
- Is message reaching the server?
- Are player names correct?

### Scenario 2: Game State Not Updating
**Breakpoint:** `sendGameState` line 232
**Check:**
- Is function being called?
- Are both players in `room.Players`?
- Is `SendMessage` succeeding?

### Scenario 3: Wrong Perspective Shown
**Breakpoint:** `sendGameState` line 237-243
**Check:**
- Value of `playerName` vs `currentPlayerName`
- Which branch is taken (if/else)?
- Parameters passed to `NewGameStatus`

### Scenario 4: Actions Don't Work
**Breakpoint:** Respective handler in `handlers.go`
**Check:**
- Is message reaching handler?
- Is payload correctly parsed?
- Does game action succeed?

## Environment Variables

You can modify debug settings:

```json
// In launch.json
"env": {
    "PORT": "8080",           // Change server port
    "DEBUG": "true",          // Custom debug flag
    "LOG_LEVEL": "verbose"    // Custom log level
}
```

## Troubleshooting Debugger

### Problem: Breakpoints Not Hit
**Solutions:**
1. Make sure you're running "Debug Server" not regular build
2. Check file path is correct
3. Rebuild: `go build ./cmd/server`
4. Restart VS Code

### Problem: Variables Show "<optimized out>"
**Solution:**
Add to `launch.json`:
```json
"buildFlags": "-gcflags='all=-N -l'"
```

### Problem: Debugger Won't Start
**Solutions:**
1. Install Delve: `go install github.com/go-delve/delve/cmd/dlv@latest`
2. Check Go extension is installed
3. Try running: `dlv version`

## Performance Profiling

To profile the server:

```bash
# CPU profile
go test -cpuprofile=cpu.prof ./...

# Memory profile
go test -memprofile=mem.prof ./...

# View with pprof
go tool pprof cpu.prof
```

## Remote Debugging

To debug a running server:

```bash
# Start server with Delve
dlv exec ./server.exe --headless --listen=:2345 --api-version=2

# In VS Code launch.json, add:
{
    "name": "Attach to Server",
    "type": "go",
    "request": "attach",
    "mode": "remote",
    "remotePath": "${workspaceFolder}",
    "port": 2345,
    "host": "localhost"
}
```

## Quick Reference

| Action | Shortcut |
|--------|----------|
| Start Debug | F5 |
| Toggle Breakpoint | F9 |
| Step Over | F10 |
| Step Into | F11 |
| Step Out | Shift+F11 |
| Continue | F5 |
| Stop | Shift+F5 |
| Restart | Ctrl+Shift+F5 |

## Additional Resources

- [VS Code Go Debugging](https://code.visualstudio.com/docs/languages/go#_debugging)
- [Delve Documentation](https://github.com/go-delve/delve)
- [Go Debugging Guide](https://go.dev/doc/diagnostics)

---

Happy Debugging! 🐛🔍
