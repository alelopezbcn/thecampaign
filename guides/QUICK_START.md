# Quick Start Guide 🚀

## TL;DR - Start Playing NOW!

```bash
./server.exe
```

Then open **two browser tabs** to: `http://localhost:8080`

Both join with the **SAME Game ID** → Play! 🎮

---

## Three Bugs Were Fixed

1. ✅ **JSON field names** - Fixed snake_case mismatch
2. ✅ **Mutex deadlock** - Fixed lock management
3. ✅ **Wrong perspective** - Each player now sees their own cards

---

## What You'll See

### After Joining
Both players see: **"Select Your Initial Warriors (1-3)"**

Click 1-3 warrior cards → Click "Confirm Selection"

### During Game

**Your View (example: John's turn)**
```
┌─────────────────────────────────┐
│ Enemy (Ryan): Castle + Field    │  ← Ryan's warriors (you can attack)
├─────────────────────────────────┤
│ ═══════ BATTLEFIELD ═══════     │
├─────────────────────────────────┤
│ Your (John): Field + Castle     │  ← Your warriors
├─────────────────────────────────┤
│ Your Hand: [Your Cards]         │  ← Only YOU see your hand
├─────────────────────────────────┤
│ Actions: [DRAW CARD] [ATTACK]   │  ← Your turn (enabled)
└─────────────────────────────────┘
```

**Ryan's View (waiting for turn)**
```
┌─────────────────────────────────┐
│ Enemy (John): Castle + Field    │  ← John's warriors
├─────────────────────────────────┤
│ ═══════ BATTLEFIELD ═══════     │
├─────────────────────────────────┤
│ Your (Ryan): Field + Castle     │  ← Your warriors
├─────────────────────────────────┤
│ Your Hand: [Your Cards]         │  ← Only YOU see your hand
├─────────────────────────────────┤
│ Actions: [All disabled]         │  ← Not your turn
└─────────────────────────────────┘
```

---

## Game Actions

**On Your Turn:**
1. **Draw Card** - Draw 1 card (happens automatically)
2. **Move Warrior** - Deploy warrior from hand to field
3. **Attack** - Select: Your warrior → Enemy target → Weapon
4. **Special Power** - Heal/Protect/InstantKill abilities
5. **Buy** - Trade gold for more cards
6. **Construct** - Build castle or add resources
7. **Trade** - Exchange 3 cards for 1 new card
8. **End Turn** - Pass to opponent

---

## Win Conditions

### You WIN if:
✅ Complete your castle (enough resources)
✅ Enemy has no warriors on field

### You LOSE if:
❌ Enemy completes castle first
❌ You have no warriors on field

---

## Card Types

- 🔴 **Warriors** (Red) - Knight, Archer, Mage, Dragon (HP shown)
- 🔵 **Weapons** (Cyan) - Sword, Arrow, Poison (Damage shown)
- 🟡 **Resources** (Yellow) - Gold cards (Value shown)
- 🟣 **Special** (Purple) - Spy, Thief, Catapult

---

## Tips

1. **Keep warriors alive** - You lose if field is empty!
2. **Balance resources** - Don't just spend, also build castle
3. **Use special cards** - Spy reveals opponent's strategy
4. **Attack strategically** - Weaken enemies before killing
5. **Dragon power** - Dragons can use ANY weapon type

---

## Troubleshooting

**Problem:** Cards don't show
- **Fix:** Open browser console (F12), check for errors

**Problem:** Can't join game
- **Fix:** Make sure BOTH players use EXACT same Game ID

**Problem:** Stuck on waiting screen
- **Fix:** Second player must also click "Join Game"

**Problem:** Server won't start
- **Fix:** `taskkill /F /IM server.exe` then restart

**Problem:** Actions don't work
- **Fix:** Check turn indicator - only active on YOUR turn

---

## Advanced: Build From Source

```bash
cd backend
go build -o ../server.exe ./cmd/server
cd ..
./server.exe
```

---

## Documentation

- [BUGS_FIXED.md](BUGS_FIXED.md) - All bugs that were fixed
- [GAME_GUIDE.md](GAME_GUIDE.md) - Detailed gameplay guide
- [IMPLEMENTATION_SUMMARY.md](IMPLEMENTATION_SUMMARY.md) - Technical details
- [BUG3_PERSPECTIVE_FIX.md](BUG3_PERSPECTIVE_FIX.md) - Latest bug fix details

---

## Project Stats

- **Language:** Go + JavaScript
- **Framework:** None (vanilla!)
- **Real-time:** WebSockets
- **Lines of Code:** ~3,000
- **Build Time:** < 3 seconds
- **Ready to Play:** ✅ YES!

---

## Have Fun! 🎉

The game is **fully working** and ready to play!

Enjoy battling with warriors, building castles, and defeating your opponent! ⚔️🏰
