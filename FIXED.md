# Bug Fix - JSON Field Name Mismatch

## Issue
The game was not progressing past the "waiting for opponent" screen when both players joined.

## Root Cause
The JavaScript frontend was trying to access JSON fields using **PascalCase** names (e.g., `CurrentPlayer`, `CurrentPlayerHand`) but the Go backend was sending JSON with **snake_case** names (e.g., `current_player`, `current_player_hand`).

This is because Go struct tags define the JSON field names:
```go
type GameStatusDTO struct {
    CurrentPlayer string `json:"current_player"`  // ← JSON uses snake_case
    // ...
}
```

## Fix Applied
Updated all JavaScript code to use snake_case field names to match the JSON response:

### Changed Field Names:
- `CurrentPlayer` → `current_player`
- `CurrentPlayerHand` → `current_player_hand`
- `CurrentPlayerField` → `current_player_field`
- `CurrentPlayerCastle` → `current_player_castle`
- `EnemyField` → `enemy_field`
- `EnemyCastle` → `enemy_castle`
- `CardsInEnemyHand` → `cards_in_enemy_hand`
- `WarriorsInHandIDs` → `warriors_in_hand_ids`
- `UsableWeaponIDs` → `usable_weapon_ids`
- `ResourceIDs` → `resource_ids`
- `ConstructionIDs` → `construction_ids`
- `SpyID` → `spy_id`
- `ThiefID` → `thief_id`
- `CatapultID` → `catapult_id`
- `ResourceCardsInEnemyCastle` → `resource_cards_in_enemy_castle`

## Files Modified
- `frontend/static/js/game.js` - All game state access code updated

## Status
✅ **FIXED** - The game should now work properly!

## How to Test
1. Kill any running server: `taskkill /F /IM server.exe`
2. Start the server: `./server.exe` or `start.bat`
3. Open two browser tabs to `http://localhost:8080`
4. Join with same Game ID in both tabs
5. You should now see the warrior selection screen! 🎉

## Developer Note
When working with JSON in Go:
- **Always check the JSON tags** on your structs
- Use browser DevTools Console to inspect actual JSON response
- JSON field names are case-sensitive
