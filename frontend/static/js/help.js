// help.js — standalone help/reference system for The Campaign
// Injects the help button, modal, styles, and all static content into the DOM.
//
// ── KEEP IN SYNC WITH THE BACKEND ──────────────────────────────────────────
// Update this file whenever any of the following change:
//
//  Modes tab:
//    - Game modes (1v1 / 2v2 / FFA3 / FFA5) or their castle goals
//    - Win conditions (castle complete, elimination rules)
//
//  Turn tab:
//    - Phase sequence (draw → attack → spy/steal → buy → construct → endturn)
//    - Phase rules (what's allowed in each phase)
//    - Limits: hand limit (board.MaxCardsInHand), turn time, trades/turn,
//              warrior moves/turn, ambush traps/field
//
//  Warriors tab:
//    - Warrior types, HP, weapon affinities, weaknesses (×2 damage rules)
//    - Dragon / Mercenary special rules
//    - Weapon types and damage ranges
//    - Special power behaviour per warrior type
//
//  Cards tab:
//    - Ambush effects (reflect, cancel, steal weapon, drain life, instant kill)
//    - Spy / Steal / Sabotage / Treason rules
//    - Other cards: Catapult, Resurrection, Gold
//
//  Events tab:
//    - Event types and their effects (domain/gameevents/)
// ───────────────────────────────────────────────────────────────────────────

(function () {
    'use strict';

    // ── Styles ──────────────────────────────────────────────────────────────
    const style = document.createElement('style');
    style.textContent = `
        /* Help button — fixed top-right, visible on all screens */
        .help-open-btn {
            position: fixed;
            top: 14px;
            right: 14px;
            z-index: 1200;
            width: 36px;
            height: 36px;
            border-radius: 50%;
            border: 1.5px solid rgba(102, 126, 234, 0.45);
            background: rgba(18, 18, 36, 0.92);
            color: #99a;
            font-size: 1em;
            font-weight: 700;
            cursor: pointer;
            display: flex;
            align-items: center;
            justify-content: center;
            backdrop-filter: blur(6px);
            transition: color 0.2s, border-color 0.2s, background 0.2s, box-shadow 0.2s;
            box-shadow: 0 2px 10px rgba(0, 0, 0, 0.5);
            line-height: 1;
        }
        .help-open-btn:hover {
            color: #fff;
            border-color: rgba(102, 126, 234, 0.9);
            background: rgba(102, 126, 234, 0.18);
            box-shadow: 0 0 14px rgba(102, 126, 234, 0.45);
        }

        /* Modal overlay */
        .help-modal-overlay {
            position: fixed;
            inset: 0;
            background: rgba(0, 0, 0, 0.62);
            display: flex;
            align-items: center;
            justify-content: center;
            z-index: 1800;
        }
        .help-modal-overlay.hidden { display: none !important; }

        /* Modal panel */
        .help-modal-panel {
            background: rgba(16, 16, 30, 0.99);
            border: 1.5px solid rgba(102, 126, 234, 0.32);
            border-radius: 16px;
            width: 94vw;
            max-width: 720px;
            height: 82vh;
            display: flex;
            flex-direction: column;
            box-shadow: 0 16px 56px rgba(0, 0, 0, 0.75);
            overflow: hidden;
        }

        /* Header */
        .help-modal-head {
            display: flex;
            align-items: center;
            justify-content: space-between;
            padding: 18px 24px 0;
            flex-shrink: 0;
        }
        .help-modal-head-title {
            font-size: 0.82em;
            color: #aaa;
            text-transform: uppercase;
            letter-spacing: 1.8px;
        }
        .help-modal-head-close {
            background: none;
            border: none;
            color: #888;
            font-size: 1.15em;
            cursor: pointer;
            padding: 4px 8px;
            border-radius: 6px;
            transition: color 0.2s;
            line-height: 1;
        }
        .help-modal-head-close:hover { color: #e0e0e0; }

        /* Tabs */
        .help-modal-tabs {
            display: flex;
            gap: 2px;
            padding: 12px 24px 0;
            flex-shrink: 0;
            border-bottom: 1px solid rgba(255, 255, 255, 0.06);
            overflow-x: auto;
            scrollbar-width: none;
        }
        .help-modal-tabs::-webkit-scrollbar { display: none; }
        .help-tab-btn {
            background: none;
            border: none;
            border-bottom: 2px solid transparent;
            color: #888;
            font-size: 0.9em;
            font-weight: 600;
            padding: 8px 12px;
            cursor: pointer;
            white-space: nowrap;
            transition: color 0.2s, border-color 0.2s;
            margin-bottom: -1px;
        }
        .help-tab-btn:hover { color: #bbb; }
        .help-tab-btn.active {
            color: #fff;
            border-bottom-color: rgba(102, 126, 234, 0.8);
        }

        /* Scrollable content area */
        .help-modal-body {
            flex: 1;
            overflow-y: auto;
            padding: 22px 24px 28px;
            scrollbar-width: thin;
            scrollbar-color: rgba(102, 126, 234, 0.3) transparent;
            /* Fixed height so all tabs feel the same size */
            height: 0;
        }
        .help-modal-body::-webkit-scrollbar { width: 5px; }
        .help-modal-body::-webkit-scrollbar-thumb {
            background: rgba(102, 126, 234, 0.3);
            border-radius: 4px;
        }
        .help-tab-content { display: none; }
        .help-tab-content.active { display: block; }

        /* Section */
        .help-section { margin-bottom: 26px; }
        .help-section:last-child { margin-bottom: 0; }
        .help-section-title {
            font-size: 0.78em;
            color: #667eea;
            text-transform: uppercase;
            letter-spacing: 1.6px;
            margin-bottom: 12px;
            display: flex;
            align-items: center;
            gap: 8px;
        }
        .help-section-title::after {
            content: '';
            flex: 1;
            height: 1px;
            background: rgba(102, 126, 234, 0.18);
        }

        /* Tables */
        .help-table {
            width: 100%;
            border-collapse: collapse;
            font-size: 0.95em;
        }
        .help-table th {
            color: #999;
            font-size: 0.8em;
            text-transform: uppercase;
            letter-spacing: 0.8px;
            padding: 6px 10px;
            text-align: left;
            border-bottom: 1px solid rgba(255, 255, 255, 0.07);
            font-weight: 600;
        }
        .help-table td {
            color: #d4d4d4;
            padding: 8px 10px;
            border-bottom: 1px solid rgba(255, 255, 255, 0.04);
            vertical-align: top;
            line-height: 1.4;
        }
        .help-table tr:last-child td { border-bottom: none; }
        .help-table tr:hover td { background: rgba(255, 255, 255, 0.015); }

        /* Chips */
        .hc { display: inline-block; padding: 2px 8px; border-radius: 10px; font-size: 0.9em; font-weight: 600; background: rgba(255,255,255,0.06); color: #ccc; }
        .hc-red    { background: rgba(231,76,60,0.15);   color: #e74c3c; }
        .hc-gold   { background: rgba(241,196,15,0.15);  color: #f1c40f; }
        .hc-purple { background: rgba(155,89,182,0.15);  color: #9b59b6; }
        .hc-green  { background: rgba(46,213,115,0.15);  color: #2ed573; }
        .hc-orange { background: rgba(230,126,34,0.15);  color: #e67e22; }
        .hc-blue   { background: rgba(102,126,234,0.15); color: #667eea; }

        /* Colored values */
        .hp { color: #2ed573; font-weight: 700; }
        .hn { color: #e74c3c; font-weight: 700; }

        /* Phase list */
        .help-phase-list {
            list-style: none;
            padding: 0;
            margin: 0;
            display: flex;
            flex-direction: column;
            gap: 8px;
        }
        .help-phase-item {
            display: flex;
            gap: 14px;
            align-items: flex-start;
            padding: 11px 14px;
            background: rgba(28, 28, 48, 0.6);
            border: 1px solid rgba(255, 255, 255, 0.06);
            border-radius: 9px;
        }
        .help-phase-icon { font-size: 1.15em; flex-shrink: 0; width: 22px; text-align: center; margin-top: 1px; }
        .help-phase-name { font-weight: 600; color: #eef; font-size: 1em; margin-bottom: 3px; }
        .help-phase-desc { color: #bbb; font-size: 0.9em; line-height: 1.45; }

        /* Event items */
        .help-event-item {
            display: flex;
            align-items: baseline;
            gap: 10px;
            background: rgba(28, 28, 48, 0.7);
            border: 1px solid rgba(102, 126, 234, 0.18);
            border-radius: 8px;
            padding: 11px 14px;
            font-size: 1em;
            margin-bottom: 8px;
        }
        .help-event-item:last-child { margin-bottom: 0; }
        .help-event-name { font-weight: 700; flex-shrink: 0; min-width: 100px; color: #eef; }
        .help-event-sep { color: #666; flex-shrink: 0; }
        .help-event-desc { color: #c0c0c0; line-height: 1.4; }
        .help-event-item[data-event="curse"]     { border-color: rgba(231,76,60,0.35); }
        .help-event-item[data-event="curse"]     .help-event-name { color: #e74c3c; }
        .help-event-item[data-event="harvest"]   { border-color: rgba(241,196,15,0.35); }
        .help-event-item[data-event="harvest"]   .help-event-name { color: #f1c40f; }
        .help-event-item[data-event="plague"]    { border-color: rgba(155,89,182,0.35); }
        .help-event-item[data-event="plague"]    .help-event-name { color: #9b59b6; }
        .help-event-item[data-event="abundance"] { border-color: rgba(46,213,115,0.35); }
        .help-event-item[data-event="abundance"] .help-event-name { color: #2ed573; }
        .help-event-item[data-event="bloodlust"]         { border-color: rgba(230,126,34,0.35); }
        .help-event-item[data-event="bloodlust"]         .help-event-name { color: #e67e22; }
        .help-event-item[data-event="champions_bounty"]  { border-color: rgba(241,196,15,0.35); }
        .help-event-item[data-event="champions_bounty"]  .help-event-name { color: #f1c40f; }

        /* Mode grid */
        .help-mode-grid {
            display: grid;
            grid-template-columns: repeat(2, 1fr);
            gap: 10px;
            margin-bottom: 4px;
        }
        .help-mode-card {
            background: rgba(28, 28, 48, 0.6);
            border: 1px solid rgba(102, 126, 234, 0.18);
            border-radius: 10px;
            padding: 14px 16px;
        }
        .help-mode-name { font-size: 1.05em; font-weight: 700; color: #eef; margin-bottom: 4px; }
        .help-mode-meta { font-size: 0.87em; color: #aaa; margin-bottom: 6px; }
        .help-mode-rule { font-size: 0.9em; color: #c0c0c0; line-height: 1.45; }

        /* Note text */
        .help-note { font-size: 0.88em; color: #999; margin-top: 8px; line-height: 1.5; }
        .help-note strong { color: #bbb; }
    `;
    document.head.appendChild(style);

    // ── Tab content ─────────────────────────────────────────────────────────

    const MODES_HTML = `
        <div class="help-section">
            <div class="help-section-title">Game Modes</div>
            <div class="help-mode-grid">
                <div class="help-mode-card">
                    <div class="help-mode-name">1v1 &mdash; Duel</div>
                    <div class="help-mode-meta">2 players &middot; Castle goal: <span class="hp">25</span></div>
                    <div class="help-mode-rule">Classic head-to-head. Eliminate your opponent or complete your castle first.</div>
                </div>
                <div class="help-mode-card">
                    <div class="help-mode-name">2v2 &mdash; Teams</div>
                    <div class="help-mode-meta">4 players &middot; Castle goal: <span class="hp">30</span></div>
                    <div class="help-mode-rule">Teams of 2. Allies can build on each other&apos;s castles and send warriors to ally fields. Both teammates must fall to be eliminated.</div>
                </div>
                <div class="help-mode-card">
                    <div class="help-mode-name">FFA 3</div>
                    <div class="help-mode-meta">3 players &middot; Castle goal: <span class="hp">25</span></div>
                    <div class="help-mode-rule">Free-for-all. Attack anyone. Last player standing or first to reach the castle goal wins.</div>
                </div>
                <div class="help-mode-card">
                    <div class="help-mode-name">FFA 5</div>
                    <div class="help-mode-meta">5 players &middot; Castle goal: <span class="hp">25</span></div>
                    <div class="help-mode-rule">Five-player chaos. Same rules as FFA&nbsp;3 but with more opponents and more targets.</div>
                </div>
            </div>
        </div>
        <div class="help-section">
            <div class="help-section-title">Win Conditions</div>
            <table class="help-table">
                <tr><th>Condition</th><th>How It Works</th></tr>
                <tr>
                    <td><span class="hc hc-gold">Castle Complete</span></td>
                    <td>Your castle&apos;s total resource value reaches the mode&apos;s goal. You (or your team in 2v2) win instantly.</td>
                </tr>
                <tr>
                    <td><span class="hc hc-red">Elimination</span></td>
                    <td>A player with no warriors on the field and no warriors in hand is eliminated. In 2v2 both teammates must be eliminated. In FFA the last surviving player wins.</td>
                </tr>
            </table>
        </div>
    `;

    const TURN_HTML = `
        <div class="help-section">
            <div class="help-section-title">Phase Sequence (every turn)</div>
            <ul class="help-phase-list">
                <li class="help-phase-item">
                    <span class="help-phase-icon">🎴</span>
                    <div>
                        <div class="help-phase-name">Draw</div>
                        <div class="help-phase-desc">Draw 1 card from the deck. Cannot be skipped.</div>
                    </div>
                </li>
                <li class="help-phase-item">
                    <span class="help-phase-icon">⚔️</span>
                    <div>
                        <div class="help-phase-name">Attack</div>
                        <div class="help-phase-desc">Play weapon cards to attack enemy warriors, use Special Power cards, move a warrior to another field (once per turn, 2v2 only for ally fields), launch a Catapult, or Resurrect a fallen warrior. Multiple attacks per turn are allowed. Skip when done.</div>
                    </div>
                </li>
                <li class="help-phase-item">
                    <span class="help-phase-icon">🎭</span>
                    <div>
                        <div class="help-phase-name">Spy / Steal</div>
                        <div class="help-phase-desc">Use Spy, Steal, Sabotage, or Treason cards against opponents. Multiple uses per turn are allowed. Skip when done.</div>
                    </div>
                </li>
                <li class="help-phase-item">
                    <span class="help-phase-icon">💰</span>
                    <div>
                        <div class="help-phase-name">Buy / Trade</div>
                        <div class="help-phase-desc">Spend Gold cards to draw new cards (value &divide; 2, rounded down). Trade 3 cards for 1 (once per turn). Place an Ambush trap in your field (max 1 per field). Skip when done.</div>
                    </div>
                </li>
                <li class="help-phase-item">
                    <span class="help-phase-icon">🏰</span>
                    <div>
                        <div class="help-phase-name">Construct</div>
                        <div class="help-phase-desc">Play resource cards to build your castle. Each card adds its face value to your castle total. In 2v2 you can also build on an ally&apos;s castle. Skip if you have nothing to play.</div>
                    </div>
                </li>
                <li class="help-phase-item">
                    <span class="help-phase-icon">🔄</span>
                    <div>
                        <div class="help-phase-name">End Turn</div>
                        <div class="help-phase-desc">Pass to the next player. A short countdown gives you a moment to review the board before the turn ends automatically.</div>
                    </div>
                </li>
            </ul>
        </div>
        <div class="help-section">
            <div class="help-section-title">Limits</div>
            <table class="help-table">
                <tr><th>Rule</th><th>Value</th></tr>
                <tr><td>Hand limit</td><td>7 cards</td></tr>
                <tr><td>Turn time limit</td><td>2 minutes</td></tr>
                <tr><td>Trades per turn</td><td>1 (3 cards &rarr; 1 card)</td></tr>
                <tr><td>Warrior moves per turn</td><td>1</td></tr>
                <tr><td>Ambush traps per field</td><td>1</td></tr>
            </table>
        </div>
    `;

    const WARRIORS_HTML = `
        <div class="help-section">
            <div class="help-section-title">Warriors</div>
            <table class="help-table">
                <tr>
                    <th>Warrior</th>
                    <th>HP</th>
                    <th>Weapon</th>
                    <th>Weak vs.</th>
                    <th>Notes</th>
                </tr>
                <tr>
                    <td><span class="hc hc-blue">Knight</span></td>
                    <td class="hp">20</td>
                    <td>Sword</td>
                    <td class="hn">Poison &times;2</td>
                    <td>&mdash;</td>
                </tr>
                <tr>
                    <td><span class="hc hc-green">Archer</span></td>
                    <td class="hp">20</td>
                    <td>Arrow</td>
                    <td class="hn">Sword &times;2</td>
                    <td>&mdash;</td>
                </tr>
                <tr>
                    <td><span class="hc hc-purple">Mage</span></td>
                    <td class="hp">20</td>
                    <td>Poison</td>
                    <td class="hn">Arrow &times;2</td>
                    <td>&mdash;</td>
                </tr>
                <tr>
                    <td><span class="hc hc-gold">Dragon</span></td>
                    <td class="hp">20</td>
                    <td>Any</td>
                    <td>None</td>
                    <td>No &times;2 damage from any weapon. Immune to Instant Kill Special Power (takes damage instead). Only Harpoon or Ambush Instant Kill can bypass immunity.</td>
                </tr>
                <tr>
                    <td><span class="hc">Mercenary</span></td>
                    <td class="hn">10</td>
                    <td>Any</td>
                    <td>None</td>
                    <td>No &times;2 damage. Lower max HP. Can carry any weapon.</td>
                </tr>
            </table>
            <p class="help-note"><strong>Field requirement:</strong> You can only play a weapon if you have a compatible warrior on your field (e.g. Sword requires a Knight, Dragon, or Mercenary).</p>
        </div>
        <div class="help-section">
            <div class="help-section-title">Weapons</div>
            <table class="help-table">
                <tr><th>Weapon</th><th>Damage</th><th>Requires on field</th><th>Effect</th></tr>
                <tr>
                    <td><span class="hc hc-blue">Sword</span></td>
                    <td>1&ndash;9</td>
                    <td>Knight / Dragon / Mercenary</td>
                    <td>Deals <span class="hp">&times;2</span> damage to Archer.</td>
                </tr>
                <tr>
                    <td><span class="hc hc-green">Arrow</span></td>
                    <td>1&ndash;9</td>
                    <td>Archer / Dragon / Mercenary</td>
                    <td>Deals <span class="hp">&times;2</span> damage to Mage.</td>
                </tr>
                <tr>
                    <td><span class="hc hc-purple">Poison</span></td>
                    <td>1&ndash;9</td>
                    <td>Mage / Dragon / Mercenary</td>
                    <td>Deals <span class="hp">&times;2</span> damage to Knight.</td>
                </tr>
                <tr>
                    <td><span class="hc hc-gold">Harpoon</span></td>
                    <td class="hp">20</td>
                    <td>Any warrior</td>
                    <td>Can <em>only</em> target Dragons. Bypasses all Dragon immunities.</td>
                </tr>
                <tr>
                    <td><span class="hc hc-red">Blood Rain</span></td>
                    <td>4 each</td>
                    <td>Any warrior</td>
                    <td>Hits <em>all</em> enemy warriors simultaneously. Each takes 4 damage (no multipliers).</td>
                </tr>
            </table>
            <p class="help-note"><strong>Damage values:</strong> Standard weapons (Sword, Arrow, Poison) come in strengths 1&ndash;9. The number on the card is the base damage before multipliers and event modifiers.</p>
        </div>
        <div class="help-section">
            <div class="help-section-title">Special Powers</div>
            <p class="help-note" style="margin-bottom:10px">Special Power cards are used during the Attack phase. Each warrior type activates a different power. Dragon cannot use Special Powers.</p>
            <table class="help-table">
                <tr><th>Played alongside</th><th>Effect</th><th>Target</th></tr>
                <tr>
                    <td><span class="hc hc-blue">Knight</span></td>
                    <td>Shields an ally warrior. The shield absorbs the next hit before the warrior takes any damage, then breaks.</td>
                    <td>Ally warrior</td>
                </tr>
                <tr>
                    <td><span class="hc hc-green">Archer</span></td>
                    <td>Instantly kills an enemy warrior regardless of HP. Dragon takes damage instead of dying.</td>
                    <td>Enemy warrior</td>
                </tr>
                <tr>
                    <td><span class="hc hc-purple">Mage</span></td>
                    <td>Restores a target ally warrior to full HP. Removes all damage and used weapons from the warrior.</td>
                    <td>Ally warrior</td>
                </tr>
                <tr>
                    <td><span class="hc hc-gold">Dragon</span></td>
                    <td class="hn">Cannot use Special Powers.</td>
                    <td>&mdash;</td>
                </tr>
            </table>
        </div>
    `;

    const CARDS_HTML = `
        <div class="help-section">
            <div class="help-section-title">Ambush</div>
            <p class="help-note" style="margin-bottom:10px">Play an Ambush card face-down in your field during the Buy phase. When any enemy attacks a warrior in your field the Ambush triggers automatically. Only one Ambush per field at a time.</p>
            <table class="help-table">
                <tr><th>Effect</th><th>What Happens When Triggered</th></tr>
                <tr>
                    <td><span class="hc">Reflect</span></td>
                    <td>The weapon&apos;s full damage is redirected to the <em>attacker&apos;s</em> warrior instead of the defender.</td>
                </tr>
                <tr>
                    <td><span class="hc">Cancel</span></td>
                    <td>The attack is completely blocked. The weapon is discarded with no damage dealt.</td>
                </tr>
                <tr>
                    <td><span class="hc">Steal Weapon</span></td>
                    <td>The attacking player loses their weapon to the defending player&apos;s hand.</td>
                </tr>
                <tr>
                    <td><span class="hc">Drain Life</span></td>
                    <td>No damage is dealt. Instead the defending warrior heals by the weapon&apos;s damage value.</td>
                </tr>
                <tr>
                    <td><span class="hc hc-red">Instant Kill</span></td>
                    <td>A random warrior from the <em>attacker&apos;s</em> field is instantly killed. Bypasses Dragon&apos;s immunity.</td>
                </tr>
            </table>
        </div>
        <div class="help-section">
            <div class="help-section-title">Spy / Steal / Sabotage</div>
            <p class="help-note" style="margin-bottom:10px">Played during the Spy &amp; Steal phase. Multiple cards can be used per turn.</p>
            <table class="help-table">
                <tr><th>Card</th><th>Effect</th></tr>
                <tr>
                    <td><span class="hc hc-blue">Spy</span></td>
                    <td>Choose one: peek at the top 5 cards of the deck, or reveal a target player&apos;s entire hand (only visible to you).</td>
                </tr>
                <tr>
                    <td><span class="hc hc-orange">Steal</span></td>
                    <td>Take a random card from a target player&apos;s hand into your own.</td>
                </tr>
                <tr>
                    <td><span class="hc hc-red">Sabotage</span></td>
                    <td>Destroy a random card from a target player&apos;s hand (sent to discard).</td>
                </tr>
                <tr>
                    <td><span class="hc hc-purple">Treason</span></td>
                    <td>Recruit a weakened enemy warrior (&le;5&nbsp;HP) from an opponent&apos;s field directly onto your field.</td>
                </tr>
            </table>
        </div>
        <div class="help-section">
            <div class="help-section-title">Other Cards</div>
            <table class="help-table">
                <tr><th>Card</th><th>Phase</th><th>Effect</th></tr>
                <tr>
                    <td><span class="hc hc-gold">Catapult</span></td>
                    <td>Attack</td>
                    <td>Removes a resource from a target player&apos;s castle at a chosen position, reducing their castle value.</td>
                </tr>
                <tr>
                    <td><span class="hc hc-green">Resurrection</span></td>
                    <td>Attack</td>
                    <td>Revives a random fallen warrior from your cemetery and places it back on your field (or an ally&apos;s field in 2v2).</td>
                </tr>
                <tr>
                    <td><span class="hc">Gold</span></td>
                    <td>Buy / Construct</td>
                    <td>Spend in Buy phase to draw cards (value &divide; 2 = cards drawn, rounded down). Or play in Construct phase to add its face value to your castle.</td>
                </tr>
            </table>
        </div>
    `;

    const EVENTS_HTML = `
        <p class="help-note" style="margin-bottom:16px">At the start of each round a random event becomes active for all players simultaneously. The current event is shown in the banner at all times. When your turn begins you&apos;ll see a reminder of the active event.</p>
        <div class="help-event-item" data-event="">
            <span class="help-event-name">Calm</span>
            <span class="help-event-sep">&#8212;</span>
            <span class="help-event-desc">No special effects this round.</span>
        </div>
        <div class="help-event-item" data-event="abundance">
            <span class="help-event-name">Abundance</span>
            <span class="help-event-sep">&#8212;</span>
            <span class="help-event-desc">Draw <span class="hp">1 extra card</span> at the start of your turn.</span>
        </div>
        <div class="help-event-item" data-event="plague">
            <span class="help-event-name">Plague</span>
            <span class="help-event-sep">&#8212;</span>
            <span class="help-event-desc">All your warriors <span class="hp">gain</span> or <span class="hn">lose</span> HP at the start of your turn. The modifier is the same for all warriors and is revealed in the event description. Warriors <strong>cannot die</strong> from this effect.</span>
        </div>
        <div class="help-event-item" data-event="harvest">
            <span class="help-event-name">Harvest</span>
            <span class="help-event-sep">&#8212;</span>
            <span class="help-event-desc">Each resource card contributes <span class="hp">more</span> or <span class="hn">less</span> value to castle construction than its face value. The modifier is shown in the event description and applies to every card played that round.</span>
        </div>
        <div class="help-event-item" data-event="curse">
            <span class="help-event-name">Curse</span>
            <span class="help-event-sep">&#8212;</span>
            <span class="help-event-desc">Two of the three basic weapon types (Sword, Arrow, Poison) deal <span class="hp">increased</span> or <span class="hn">reduced</span> damage. One weapon type is randomly excluded from the effect and deals normal damage. The exact modifier and excluded weapon are shown in the event description.</span>
        </div>
        <div class="help-event-item" data-event="bloodlust">
            <span class="help-event-name">Bloodlust</span>
            <span class="help-event-sep">&#8212;</span>
            <span class="help-event-desc">Whenever one of your warriors kills an enemy, that warrior is immediately restored <span class="hp">2&nbsp;HP</span>. Killing multiple enemies in one turn stacks the healing.</span>
        </div>
        <div class="help-event-item" data-event="champions_bounty">
            <span class="help-event-name">Champion's Bounty</span>
            <span class="help-event-sep">&#8212;</span>
            <span class="help-event-desc"><em>(FFA3 / FFA5 only)</em> When kills any warrior belonging to the enemy whose warriors have the <span class="hp">highest combined HP</span> on the field, you immediately draw <span class="hp">2&nbsp;cards</span>. Ties count — if multiple enemies share the top HP total, killing any of their warriors grants the reward.</span>
        </div>
    `;

    // ── Tab definitions ──────────────────────────────────────────────────────
    const TABS = [
        { id: 'modes',    label: '🎮 Modes',    html: MODES_HTML    },
        { id: 'turn',     label: '🔄 Turn',     html: TURN_HTML     },
        { id: 'warriors', label: '⚔️ Warriors', html: WARRIORS_HTML },
        { id: 'cards',    label: '🃏 Cards',    html: CARDS_HTML    },
        { id: 'events',   label: '🌪️ Events',  html: EVENTS_HTML   },
    ];

    // ── DOM Injection ────────────────────────────────────────────────────────
    function buildHTML() {
        const firstTab = TABS[0].id;
        const tabBtns = TABS.map(t =>
            `<button class="help-tab-btn${t.id === firstTab ? ' active' : ''}" data-tab="${t.id}">${t.label}</button>`
        ).join('');
        const tabPanes = TABS.map(t =>
            `<div class="help-tab-content${t.id === firstTab ? ' active' : ''}" id="help-tab-${t.id}">${t.html}</div>`
        ).join('');

        return `
            <div id="help-modal-overlay" class="help-modal-overlay hidden">
                <div class="help-modal-panel">
                    <div class="help-modal-head">
                        <span class="help-modal-head-title">How to Play &mdash; The Campaign</span>
                        <button class="help-modal-head-close" id="help-close-btn">&#x2715;</button>
                    </div>
                    <div class="help-modal-tabs" id="help-modal-tabs">${tabBtns}</div>
                    <div class="help-modal-body">${tabPanes}</div>
                </div>
            </div>
            <button class="help-open-btn" id="help-open-btn" title="How to Play">?</button>
        `;
    }

    function init() {
        const wrapper = document.createElement('div');
        wrapper.innerHTML = buildHTML();
        while (wrapper.firstChild) {
            document.body.appendChild(wrapper.firstChild);
        }

        document.getElementById('help-open-btn').addEventListener('click', openHelp);
        document.getElementById('help-close-btn').addEventListener('click', closeHelp);

        document.getElementById('help-modal-overlay').addEventListener('click', (e) => {
            if (e.target === e.currentTarget) closeHelp();
        });

        document.getElementById('help-modal-tabs').addEventListener('click', (e) => {
            const btn = e.target.closest('.help-tab-btn');
            if (!btn) return;
            const tabId = btn.dataset.tab;
            document.querySelectorAll('.help-tab-btn').forEach(b => b.classList.toggle('active', b === btn));
            document.querySelectorAll('.help-tab-content').forEach(c =>
                c.classList.toggle('active', c.id === 'help-tab-' + tabId)
            );
        });

        document.addEventListener('keydown', (e) => {
            if (e.key === 'Escape') closeHelp();
        });
    }

    function openHelp(tabId) {
        document.getElementById('help-modal-overlay').classList.remove('hidden');
        if (typeof tabId === 'string') {
            document.querySelectorAll('.help-tab-btn').forEach(b =>
                b.classList.toggle('active', b.dataset.tab === tabId)
            );
            document.querySelectorAll('.help-tab-content').forEach(c =>
                c.classList.toggle('active', c.id === 'help-tab-' + tabId)
            );
        }
    }

    function closeHelp() {
        document.getElementById('help-modal-overlay').classList.add('hidden');
    }

    // Expose globally so other scripts can open the modal directly
    window.openHelp = openHelp;
    window.closeHelp = closeHelp;

    // Script is at bottom of <body> — DOM is ready
    init();
}());
