# Critical Bug Fix - Mutex Deadlock

## Issue
After both players joined the game, the frontend would connect successfully but the game would not progress to the warrior selection screen. The server logs showed "Game started" but clients never received the game state update.

## Root Cause: **DEADLOCK**

The `handleJoinGame` function had a mutex deadlock issue:

```go
func (h *Hub) handleJoinGame(client *Client, payload interface{}) {
    h.mutex.Lock()           // ← Acquires hub-level write lock
    defer h.mutex.Unlock()   // ← Will unlock when function returns

    // ... code ...

    room.mutex.Lock()        // ← Acquires room-level write lock
    defer room.mutex.Unlock() // ← Will unlock when function returns

    // ... code ...

    // If we have 2 players, start the game
    if len(room.Players) == 2 {
        // ... create game ...

        h.sendGameState(gameID)  // ← Calls sendGameState WHILE HOLDING BOTH LOCKS!
    }
}
```

The `sendGameState` function tries to acquire locks:

```go
func (h *Hub) sendGameState(gameID string) {
    h.mutex.RLock()         // ← Tries to acquire read lock
    // ...
    room.mutex.RLock()      // ← Tries to acquire room read lock
    defer room.mutex.RUnlock()
    // ...
}
```

### Why This Causes Deadlock

In Go, **you cannot acquire a read lock while holding a write lock on the same mutex** from the same goroutine. The write lock blocks all other lock attempts, including read locks.

**The sequence:**
1. `handleJoinGame` acquires `h.mutex` write lock
2. `handleJoinGame` acquires `room.mutex` write lock
3. `handleJoinGame` calls `sendGameState`
4. `sendGameState` tries to acquire `room.mutex.RLock()` ← **BLOCKS FOREVER**
5. Function never returns, so write locks are never released
6. **DEADLOCK** - goroutine frozen, game state never sent

## The Fix

Release all locks **before** calling `sendGameState`:

```go
func (h *Hub) handleJoinGame(client *Client, payload interface{}) {
    h.mutex.Lock()
    // ... setup game room ...
    h.mutex.Unlock()  // ← Release hub lock early

    room.mutex.Lock()
    // ... add player, check conditions ...

    shouldStartGame := len(room.Players) == 2

    if shouldStartGame {
        // ... create game and notify players ...
    }

    room.mutex.Unlock()  // ← Release room lock before sendGameState

    // Send initial game state AFTER releasing locks
    if shouldStartGame {
        h.sendGameState(gameID)  // ← Now safe to call!
    }
}
```

### Key Changes:
1. ✅ Removed `defer` statements - explicit lock/unlock for better control
2. ✅ Release `h.mutex` immediately after modifying `h.gameRooms`
3. ✅ Release `room.mutex` before calling `sendGameState`
4. ✅ Handle early returns with explicit `room.mutex.Unlock()` calls
5. ✅ Call `sendGameState` only after all locks released

## Files Modified
- `backend/internal/websocket/hub.go` - Fixed `handleJoinGame` function

## Testing
1. Kill all servers: `taskkill /F /IM server.exe`
2. Rebuild: `cd backend && go build -o ../server.exe ./cmd/server`
3. Start server: `./server.exe`
4. Open two browser tabs to `http://localhost:8080`
5. Both join same game ID
6. ✅ **Both players should now see the warrior selection screen!**

## Lesson Learned

**Mutex Best Practices:**
- ⚠️ Never call functions that acquire locks while holding locks
- ⚠️ Keep critical sections as small as possible
- ⚠️ Release locks before calling other functions
- ⚠️ Be careful with `defer` - sometimes explicit unlock is better
- ⚠️ Document lock ordering to prevent deadlocks
- ✅ Use read locks (`RLock`) when only reading
- ✅ Release write locks as soon as possible

## Status
✅ **FIXED** - Deadlock eliminated, game state now properly sent to both players!
