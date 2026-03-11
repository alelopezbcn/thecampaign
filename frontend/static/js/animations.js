// Extract HP values from field cards for damage detection
function extractFieldHP(status) {
    const hpMap = {};
    if (status) {
        (status.current_player_field || []).forEach(card => {
            hpMap[card.id] = card.value;
        });
        (status.opponents || []).forEach(opp => {
            (opp.field || []).forEach(card => {
                hpMap[card.id] = card.value;
            });
        });
    }
    return hpMap;
}

function showDamageFeedback(previousState, newState) {
    const previousHP = extractFieldHP(previousState);
    const newHP = extractFieldHP(newState);

    // Reset screen flash flag so only one flash per damage batch
    screenFlashShown = false;

    // Check for HP changes
    const skipSlash = newState.last_action === 'blood_rain';
    for (const cardId in previousHP) {
        if (newHP[cardId] !== undefined && newHP[cardId] < previousHP[cardId]) {
            const damage = previousHP[cardId] - newHP[cardId];
            showFloatingDamage(cardId, damage, skipSlash);
        }
    }

    // Check for healed warriors (HP increased)
    for (const cardId in previousHP) {
        if (newHP[cardId] !== undefined && newHP[cardId] > previousHP[cardId]) {
            showFloatingHeal(cardId, previousHP[cardId], newHP[cardId]);
        }
    }
}

// Display attack impact animation on a card
function showAttackAnimation(cardElement) {
    // Add shake + red glow
    cardElement.classList.add('taking-damage');

    // Create slash overlay
    const slashContainer = document.createElement('div');
    slashContainer.className = 'attack-slash-container';

    const slash1 = document.createElement('div');
    slash1.className = 'slash-line slash-line-1';

    const slash2 = document.createElement('div');
    slash2.className = 'slash-line slash-line-2';

    slashContainer.appendChild(slash1);
    slashContainer.appendChild(slash2);
    cardElement.appendChild(slashContainer);

    // Screen flash (once per damage batch)
    if (!screenFlashShown) {
        screenFlashShown = true;
        showScreenFlash();
    }

    // Cleanup after animations complete
    setTimeout(() => {
        cardElement.classList.remove('taking-damage');
        slashContainer.remove();
    }, 2000);
}

// Brief flash across the screen (red for attacks, orange for fire kills)
function showScreenFlash(isFire) {
    const flash = document.createElement('div');
    flash.className = 'screen-flash-overlay' + (isFire ? ' fire-flash' : '');
    document.body.appendChild(flash);
    setTimeout(() => flash.remove(), 600);
}

// Detect killed warriors and clone their elements before DOM re-render
function prepareDeathAnimations(previousState, newState) {
    const previousHP = extractFieldHP(previousState);
    const newHP = extractFieldHP(newState);
    const isInstantKill = newState.last_action === 'special_power';
    const killed = [];

    for (const cardId in previousHP) {
        // Card existed before but is gone now = killed
        if (newHP[cardId] === undefined) {
            const cardElement = document.querySelector(`.card[data-card-id="${cardId}"]`);
            if (!cardElement) continue;

            const rect = cardElement.getBoundingClientRect();
            const clone = cardElement.cloneNode(true);
            const damage = previousHP[cardId]; // full HP was the lethal damage

            killed.push({ clone, rect, cardId, damage, isInstantKill });
        }
    }

    return killed;
}

// Play death animations for killed warriors using cloned elements
function playDeathAnimations(killedWarriors) {
    // Screen flash for the kills
    const hasFireKill = killedWarriors.some(k => k.isInstantKill);
    if (!screenFlashShown) {
        screenFlashShown = true;
        showScreenFlash(hasFireKill);
    }

    killedWarriors.forEach(({ clone, rect, damage, isInstantKill }) => {
        // Create ghost container positioned where the card was
        const ghost = document.createElement('div');
        ghost.className = 'death-ghost' + (isInstantKill ? ' fire-kill' : '');
        ghost.style.left = rect.left + 'px';
        ghost.style.top = rect.top + 'px';
        ghost.style.width = rect.width + 'px';
        ghost.style.height = rect.height + 'px';

        // Style the clone to fill the ghost container
        clone.style.width = '100%';
        clone.style.height = '100%';
        clone.style.margin = '0';
        clone.classList.add('death-slash-phase');
        ghost.appendChild(clone);

        // Add slash marks on the ghost
        const slashContainer = document.createElement('div');
        slashContainer.className = 'attack-slash-container';
        const slash1 = document.createElement('div');
        slash1.className = 'slash-line slash-line-1';
        const slash2 = document.createElement('div');
        slash2.className = 'slash-line slash-line-2';
        slashContainer.appendChild(slash1);
        slashContainer.appendChild(slash2);
        ghost.appendChild(slashContainer);

        // Add floating damage number
        const floatingNum = document.createElement('div');
        floatingNum.className = 'floating-damage';
        floatingNum.textContent = `-${damage}`;
        ghost.appendChild(floatingNum);

        // Add fire particles for instant kill
        if (isInstantKill) {
            for (let i = 0; i < 10; i++) {
                const particle = document.createElement('div');
                particle.className = 'fire-particle';
                const xMid = (Math.random() - 0.5) * 40;
                const xEnd = xMid + (Math.random() - 0.5) * 30;
                const yMid = -20 - Math.random() * 40;
                const yEnd = -50 - Math.random() * 60;
                particle.style.setProperty('--particle-x', xMid + 'px');
                particle.style.setProperty('--particle-x-end', xEnd + 'px');
                particle.style.setProperty('--particle-y', yMid + 'px');
                particle.style.setProperty('--particle-y-end', yEnd + 'px');
                particle.style.setProperty('--particle-delay', (Math.random() * 0.5) + 's');
                particle.style.setProperty('--particle-duration', (0.8 + Math.random() * 0.8) + 's');
                particle.style.left = (20 + Math.random() * 60) + '%';
                particle.style.top = (20 + Math.random() * 60) + '%';
                ghost.appendChild(particle);
            }
        }

        document.body.appendChild(ghost);

        // After attack animation, start death animation
        setTimeout(() => {
            slashContainer.remove();
            ghost.classList.add('dying');

            // Add skull overlay
            const skull = document.createElement('div');
            skull.className = 'death-skull';
            skull.textContent = isInstantKill ? '\u{1F525}' : '\u{1F480}';
            ghost.appendChild(skull);
        }, 1000);

        // Cleanup everything (fire kills get extra time for particles)
        setTimeout(() => ghost.remove(), isInstantKill ? 3500 : 2500);
    });
}

// Display floating damage number on a card
function showFloatingDamage(cardId, damage, skipSlash = false) {
    const cardElement = document.querySelector(`.card[data-card-id="${cardId}"]`);
    if (!cardElement) return;

    // Play attack impact animation
    if (!skipSlash) showAttackAnimation(cardElement);

    const floatingNum = document.createElement('div');
    floatingNum.className = 'floating-damage';
    floatingNum.textContent = `-${damage}`;
    cardElement.appendChild(floatingNum);

    // Remove after animation completes
    setTimeout(() => floatingNum.remove(), 3500);
}

// Display heal animation on a card with green cross and floating +amount bonus
function showFloatingHeal(cardId, fromHp, toHp) {
    const cardElement = document.querySelector(`.card[data-card-id="${cardId}"]`);
    if (!cardElement) return;

    // Green glow on card
    cardElement.classList.add('healing');
    setTimeout(() => cardElement.classList.remove('healing'), 2500);

    // Green cross overlay
    const cross = document.createElement('div');
    cross.className = 'heal-cross';
    cross.textContent = '\u271A';
    cardElement.appendChild(cross);
    setTimeout(() => cross.remove(), 2000);

    // Floating +amount bonus (shows the heal delta, not the final HP)
    const amount = toHp - fromHp;
    const floatingBonus = document.createElement('div');
    floatingBonus.className = 'floating-heal';
    floatingBonus.textContent = `+${amount}`;
    cardElement.appendChild(floatingBonus);
    setTimeout(() => floatingBonus.remove(), 3000);
}

// Extract protection state from game status
function extractProtectionState(status) {
    const protMap = {};
    if (status) {
        (status.current_player_field || []).forEach(card => {
            if (card.protected_by && card.protected_by.id) {
                protMap[card.id] = card.protected_by.id;
            }
        });
        (status.opponents || []).forEach(opp => {
            (opp.field || []).forEach(card => {
                if (card.protected_by && card.protected_by.id) {
                    protMap[card.id] = card.protected_by.id;
                }
            });
        });
    }
    return protMap;
}

// Detect warriors that just gained protection
function detectNewProtections(previousState, newState) {
    const prevProt = extractProtectionState(previousState);
    const newProt = extractProtectionState(newState);
    const newlyProtected = [];

    for (const cardId in newProt) {
        if (!prevProt[cardId]) {
            newlyProtected.push(cardId);
        }
    }
    return newlyProtected;
}

// Show shield activation animation on newly protected cards
function showProtectionAnimations(cardIds) {
    cardIds.forEach(cardId => {
        const cardElement = document.querySelector(`.card[data-card-id="${cardId}"]`);
        if (!cardElement) return;

        // Teal glow on card
        cardElement.classList.add('shield-activating');
        setTimeout(() => cardElement.classList.remove('shield-activating'), 2000);

        // Expanding shield circle
        const circle = document.createElement('div');
        circle.className = 'shield-expand-overlay';
        cardElement.appendChild(circle);
        setTimeout(() => circle.remove(), 1600);
    });
}

// Detect warriors that lost their shield (was protected, now not, warrior still alive)
function detectBrokenShields(previousState, newState) {
    const prevProt = extractProtectionState(previousState);
    const newProt = extractProtectionState(newState);
    const newFieldIds = new Set();
    (newState.current_player_field || []).forEach(c => newFieldIds.add(c.id));
    (newState.opponents || []).forEach(opp => {
        (opp.field || []).forEach(c => newFieldIds.add(c.id));
    });

    const broken = [];
    for (const cardId in prevProt) {
        // Had a shield before, doesn't have one now, but warrior still on field
        if (!newProt[cardId] && newFieldIds.has(cardId)) {
            broken.push(cardId);
        }
    }
    return broken;
}

// Show shield break animation on cards that lost protection
function showShieldBreakAnimations(cardIds) {
    cardIds.forEach(cardId => {
        const cardElement = document.querySelector(`.card[data-card-id="${cardId}"]`);
        if (!cardElement) return;

        cardElement.classList.add('shield-breaking');
        setTimeout(() => cardElement.classList.remove('shield-breaking'), 2200);

        // Screen flash for shield break
        const flash = document.createElement('div');
        flash.className = 'screen-flash-overlay';
        flash.style.background = 'radial-gradient(ellipse at center, rgba(78, 205, 196, 0.3) 0%, rgba(78, 205, 196, 0.05) 70%, transparent 100%)';
        document.body.appendChild(flash);
        setTimeout(() => flash.remove(), 600);

        // Shattering shield fragments
        for (let i = 0; i < 10; i++) {
            const angle = (i * 36 + Math.random() * 20) * Math.PI / 180;
            const dist = 50 + Math.random() * 50;
            const frag = document.createElement('div');
            frag.className = 'shield-fragment';
            frag.textContent = '🛡️';
            frag.style.setProperty('--tx', (Math.cos(angle) * dist) + 'px');
            frag.style.setProperty('--ty', (Math.sin(angle) * dist) + 'px');
            frag.style.animationDelay = (Math.random() * 0.15) + 's';
            cardElement.appendChild(frag);
            setTimeout(() => frag.remove(), 1600);
        }
    });
}

// Detect hand cards that will vanish on re-render (used/consumed cards)
function prepareHandCardVanish(previousState, newState) {
    if (!previousState || !previousState.current_player_hand) return [];

    const prevHandIds = new Set(previousState.current_player_hand.map(c => c.id));
    const newHandIds = new Set((newState.current_player_hand || []).map(c => c.id));
    const newCardIds = new Set(newState.new_cards || []);
    const vanished = [];

    for (const cardId of prevHandIds) {
        // Card was in hand before, not in hand now, and not a newly-received card
        if (!newHandIds.has(cardId) && !newCardIds.has(cardId)) {
            const cardElement = document.querySelector(`#player-hand .card[data-card-id="${cardId}"]`);
            if (!cardElement) continue;

            const rect = cardElement.getBoundingClientRect();
            const clone = cardElement.cloneNode(true);
            vanished.push({ clone, rect });
        }
    }

    return vanished;
}

// Play vanish animations for consumed hand cards
function playCardVanishAnimations(vanishedCards) {
    vanishedCards.forEach(({ clone, rect }) => {
        const ghost = document.createElement('div');
        ghost.className = 'card-vanish-ghost';
        ghost.style.left = rect.left + 'px';
        ghost.style.top = rect.top + 'px';
        ghost.style.width = rect.width + 'px';
        ghost.style.height = rect.height + 'px';

        clone.style.width = '100%';
        clone.style.height = '100%';
        clone.style.margin = '0';
        ghost.appendChild(clone);

        document.body.appendChild(ghost);
        setTimeout(() => ghost.remove(), 900);
    });
}

// Detect castle construction and gold additions
function detectCastleChanges(previousState, newState) {
    const constructions = [];
    const goldAdded = [];

    // Player's own castle
    const prevCastle = previousState.current_player_castle || {};
    const newCastle = newState.current_player_castle || {};

    if (!prevCastle.constructed && newCastle.constructed) {
        constructions.push({ containerId: 'player-castle' });
    } else if (newCastle.constructed && (newCastle.value || 0) > (prevCastle.value || 0)) {
        goldAdded.push({ containerId: 'player-castle', amount: (newCastle.value || 0) - (prevCastle.value || 0) });
    }

    // Opponent castles
    const prevOpponents = previousState.opponents || [];
    const newOpponents = newState.opponents || [];

    newOpponents.forEach((newOpp, idx) => {
        const prevOpp = prevOpponents.find(p => p.player_name === newOpp.player_name);
        if (!prevOpp) return;

        const prevC = prevOpp.castle || {};
        const newC = newOpp.castle || {};

        // Find opponent castle container by player name
        const oppArea = document.querySelector(`[data-opponent-name="${newOpp.player_name}"]`);
        if (!oppArea) return;
        const castleContainer = oppArea.querySelector('.castle');
        if (!castleContainer) return;

        if (!prevC.constructed && newC.constructed) {
            constructions.push({ container: castleContainer });
        } else if (newC.constructed && (newC.value || 0) > (prevC.value || 0)) {
            goldAdded.push({ container: castleContainer, amount: (newC.value || 0) - (prevC.value || 0) });
        }
    });

    // Detect gold removed (catapult attack)
    const goldRemoved = [];

    // Player's own castle attacked
    if (newCastle.constructed && (newCastle.value || 0) < (prevCastle.value || 0)) {
        goldRemoved.push({ containerId: 'player-castle', amount: (prevCastle.value || 0) - (newCastle.value || 0) });
    }

    // Opponent castles attacked
    newOpponents.forEach(newOpp => {
        const prevOpp = prevOpponents.find(p => p.player_name === newOpp.player_name);
        if (!prevOpp) return;
        const prevC = prevOpp.castle || {};
        const newC = newOpp.castle || {};
        if (newC.constructed && (newC.value || 0) < (prevC.value || 0)) {
            const oppArea = document.querySelector(`[data-opponent-name="${newOpp.player_name}"]`);
            if (!oppArea) return;
            const castleContainer = oppArea.querySelector('.castle');
            if (!castleContainer) return;
            goldRemoved.push({ container: castleContainer, amount: (prevC.value || 0) - (newC.value || 0) });
        }
    });

    // Detect fortress destroyed (is_protected went from true to false)
    const fortressDestroyed = [];

    if (prevCastle.is_protected && !newCastle.is_protected) {
        fortressDestroyed.push({ containerId: 'player-castle' });
    }

    newOpponents.forEach(newOpp => {
        const prevOpp = prevOpponents.find(p => p.player_name === newOpp.player_name);
        if (!prevOpp) return;
        const prevC = prevOpp.castle || {};
        const newC = newOpp.castle || {};
        if (prevC.is_protected && !newC.is_protected) {
            const oppArea = document.querySelector(`[data-opponent-name="${newOpp.player_name}"]`);
            if (!oppArea) return;
            const castleContainer = oppArea.querySelector('.castle');
            if (!castleContainer) return;
            fortressDestroyed.push({ container: castleContainer });
        }
    });

    return { constructions, goldAdded, goldRemoved, fortressDestroyed };
}

// Castle construction celebration animation
function showCastleConstructionAnimation(change) {
    const container = change.container || document.getElementById(change.containerId);
    if (!container) return;

    container.style.position = 'relative';
    container.classList.add('castle-just-constructed');

    const text = document.createElement('div');
    text.className = 'castle-construct-text';
    text.textContent = 'Castle Built!';
    container.appendChild(text);

    setTimeout(() => {
        container.classList.remove('castle-just-constructed');
        text.remove();
    }, 3000);
}

// Castle gold addition animation
function showCastleGoldAnimation(change) {
    const container = change.container || document.getElementById(change.containerId);
    if (!container) return;

    container.style.position = 'relative';
    container.classList.add('castle-gold-added');

    const floatingGold = document.createElement('div');
    floatingGold.className = 'castle-gold-floating';
    floatingGold.textContent = `+${change.amount}`;
    container.appendChild(floatingGold);

    setTimeout(() => {
        container.classList.remove('castle-gold-added');
        floatingGold.remove();
    }, 2500);
}

// Castle attack animation (catapult - gold removed)
function showCastleAttackAnimation(change) {
    const container = change.container || document.getElementById(change.containerId);
    if (!container) return;

    container.style.position = 'relative';
    container.classList.add('castle-attacked');

    // Floating damage number
    const floatingDmg = document.createElement('div');
    floatingDmg.className = 'castle-attack-floating';
    floatingDmg.textContent = `-${change.amount}`;
    container.appendChild(floatingDmg);

    // Impact flash
    const flash = document.createElement('div');
    flash.className = 'castle-attack-flash';
    container.appendChild(flash);

    setTimeout(() => {
        container.classList.remove('castle-attacked');
        floatingDmg.remove();
        flash.remove();
    }, 3000);
}

// Fortress wall destroyed by catapult animation
function showFortressDestroyedAnimation(change) {
    const container = change.container || document.getElementById(change.containerId);
    if (!container) return;

    container.style.position = 'relative';
    container.classList.add('fortress-destroyed');

    const text = document.createElement('div');
    text.className = 'fortress-destroyed-text';
    text.textContent = '🛡 Wall Destroyed!';
    container.appendChild(text);

    setTimeout(() => {
        container.classList.remove('fortress-destroyed');
        text.remove();
    }, 3000);
}

// Detect steal action and clone a card-back from the victim's hand before re-render
// Prepare warrior move animation: capture source position before re-render
function prepareWarriorMoveAnimation(previousState, newState) {
    if (newState.last_action !== 'move_warrior') return null;
    const warriorID = newState.last_moved_warrior_id;
    if (!warriorID) return null;

    const turnPlayer = newState.turn_player;
    const isMyMove = turnPlayer === newState.current_player;

    if (isMyMove) {
        // Active player: clone the warrior card from hand
        const handCard = document.querySelector(`#player-hand .card[data-card-id="${warriorID}"]`);
        if (!handCard) return null;
        const rect = handCard.getBoundingClientRect();
        const clone = handCard.cloneNode(true);
        return { type: 'self', clone, rect, warriorID };
    } else {
        // Opponent view: capture their hand area position
        const oppBoard = document.querySelector(`[data-opponent-name="${turnPlayer}"]`);
        if (!oppBoard) return null;
        const cardBacks = oppBoard.querySelectorAll('.opponent-hand-card');
        if (cardBacks.length === 0) return null;
        const lastCard = cardBacks[cardBacks.length - 1];
        const rect = lastCard.getBoundingClientRect();
        const clone = lastCard.cloneNode(true);
        return { type: 'opponent', clone, rect, warriorID, turnPlayer };
    }
}

// Play warrior move animation: animate from hand to field after re-render
function playWarriorMoveAnimation(data, status) {
    if (!data) return;

    let targetEl;
    if (data.type === 'self') {
        // Find the warrior now in player's field
        targetEl = document.querySelector(`#player-field .card[data-card-id="${data.warriorID}"]`);
    } else {
        // Find the new warrior in opponent's field
        const oppBoard = document.querySelector(`[data-opponent-name="${data.turnPlayer}"]`);
        if (oppBoard) {
            targetEl = oppBoard.querySelector(`.card[data-card-id="${data.warriorID}"]`);
        }
    }

    if (!targetEl) return;
    const targetRect = targetEl.getBoundingClientRect();

    const dx = (targetRect.left + targetRect.width / 2) - (data.rect.left + data.rect.width / 2);
    const dy = (targetRect.top + targetRect.height / 2) - (data.rect.top + data.rect.height / 2);

    const ghost = document.createElement('div');
    ghost.className = 'warrior-move-ghost';
    ghost.style.left = data.rect.left + 'px';
    ghost.style.top = data.rect.top + 'px';
    ghost.style.width = data.rect.width + 'px';
    ghost.style.height = data.rect.height + 'px';
    ghost.style.setProperty('--dx', dx + 'px');
    ghost.style.setProperty('--dy', dy + 'px');

    data.clone.style.width = '100%';
    data.clone.style.height = '100%';
    data.clone.style.margin = '0';
    ghost.appendChild(data.clone);

    document.body.appendChild(ghost);
    setTimeout(() => ghost.remove(), 1200);
}

// Prepare attack animation: capture weapon card position before re-render
function prepareAttackAnimation(previousState, newState) {
    if (newState.last_action !== 'attack' && newState.last_action !== 'harpoon') return null;
    const weaponID = newState.last_attack_weapon_id;
    const targetID = newState.last_attack_target_id;
    const targetPlayer = newState.last_attack_target_player;
    if (!weaponID || !targetID || !targetPlayer) return null;

    const turnPlayer = newState.turn_player;
    const isMyAttack = turnPlayer === newState.current_player;

    if (isMyAttack) {
        // Attacker view: clone weapon card from hand (it was just used)
        const handCard = document.querySelector(`#player-hand .card[data-card-id="${weaponID}"]`);
        if (!handCard) return null;
        const rect = handCard.getBoundingClientRect();
        const clone = handCard.cloneNode(true);
        return { type: 'self', clone, rect, targetID, targetPlayer };
    } else {
        // Opponent view: capture opponent hand area for card-back
        const oppBoard = document.querySelector(`[data-opponent-name="${turnPlayer}"]`);
        if (!oppBoard) return null;
        const cardBacks = oppBoard.querySelectorAll('.opponent-hand-card');
        if (cardBacks.length === 0) return null;
        const lastCard = cardBacks[cardBacks.length - 1];
        const rect = lastCard.getBoundingClientRect();
        const clone = lastCard.cloneNode(true);
        return { type: 'opponent', clone, rect, targetID, targetPlayer, turnPlayer };
    }
}

// Play attack animation: animate weapon card flying from hand to target warrior
function playAttackAnimation(data, status) {
    if (!data) return;

    let targetEl;
    const targetPlayer = data.targetPlayer;
    const isTargetMe = targetPlayer === status.current_player;

    if (isTargetMe) {
        // Target is on my field
        targetEl = document.querySelector(`#player-field .card[data-card-id="${data.targetID}"]`);
    } else {
        // Target is on an opponent's field
        const oppBoard = document.querySelector(`[data-opponent-name="${targetPlayer}"]`);
        if (oppBoard) {
            targetEl = oppBoard.querySelector(`.card[data-card-id="${data.targetID}"]`);
        }
    }

    if (!targetEl) return;
    const targetRect = targetEl.getBoundingClientRect();

    const dx = (targetRect.left + targetRect.width / 2) - (data.rect.left + data.rect.width / 2);
    const dy = (targetRect.top + targetRect.height / 2) - (data.rect.top + data.rect.height / 2);

    const ghost = document.createElement('div');
    ghost.className = 'attack-fly-ghost';
    ghost.style.left = data.rect.left + 'px';
    ghost.style.top = data.rect.top + 'px';
    ghost.style.width = data.rect.width + 'px';
    ghost.style.height = data.rect.height + 'px';
    ghost.style.setProperty('--dx', dx + 'px');
    ghost.style.setProperty('--dy', dy + 'px');

    data.clone.style.width = '100%';
    data.clone.style.height = '100%';
    data.clone.style.margin = '0';
    ghost.appendChild(data.clone);

    document.body.appendChild(ghost);

    // On impact, trigger a flash on the target card
    setTimeout(() => {
        targetEl.classList.add('attack-impact');
        setTimeout(() => targetEl.classList.remove('attack-impact'), 600);
    }, 900);

    setTimeout(() => ghost.remove(), 1300);
}

function prepareBloodRainAnimation(previousState, newState) {
    if (newState.last_action !== 'blood_rain') return null;
    const targetPlayer = newState.last_attack_target_player;
    if (!targetPlayer) return null;
    return { targetPlayer };
}

function playBloodRainAnimation(data, status) {
    if (!data) return;

    const isTargetMe = data.targetPlayer === status.current_player;
    let fieldEl;
    if (isTargetMe) {
        fieldEl = document.getElementById('player-field');
    } else {
        const oppBoard = document.querySelector(`[data-opponent-name="${data.targetPlayer}"]`);
        if (oppBoard) fieldEl = oppBoard.querySelector('.opponent-field');
    }
    if (!fieldEl) return;

    const rect = fieldEl.getBoundingClientRect();

    // Full-screen red vignette
    const vignette = document.createElement('div');
    vignette.className = 'blood-rain-vignette';
    document.body.appendChild(vignette);
    setTimeout(() => vignette.remove(), 2300);

    // Drop overlay — padded beyond field edges so drops spill over
    const PAD = 40;
    const overlay = document.createElement('div');
    overlay.className = 'blood-rain-overlay';
    overlay.style.left = (rect.left - PAD) + 'px';
    overlay.style.top = (rect.top - PAD) + 'px';
    overlay.style.width = (rect.width + PAD * 2) + 'px';
    overlay.style.height = (rect.height + PAD * 2) + 'px';

    const DROP_COUNT = 55;
    for (let i = 0; i < DROP_COUNT; i++) {
        const drop = document.createElement('div');
        drop.className = 'blood-drop';
        const w = 2 + Math.random() * 6;       // 2–8 px wide
        const h = 22 + Math.random() * 28;      // 22–50 px tall
        drop.style.width = w + 'px';
        drop.style.height = h + 'px';
        drop.style.left = (Math.random() * 110 - 5) + '%';
        drop.style.animationDelay = (Math.random() * 0.4) + 's';
        drop.style.animationDuration = (0.28 + Math.random() * 0.28) + 's';
        overlay.appendChild(drop);
    }

    document.body.appendChild(overlay);

    // Impact: splatter particles + blood glow on each card
    setTimeout(() => {
        for (let i = 0; i < 18; i++) {
            const splat = document.createElement('div');
            splat.className = 'blood-splatter';
            const size = 10 + Math.random() * 28;
            splat.style.width = size + 'px';
            splat.style.height = size + 'px';
            splat.style.left = (Math.random() * 95) + '%';
            splat.style.top = (Math.random() * 85) + '%';
            splat.style.animationDuration = (0.5 + Math.random() * 0.3) + 's';
            overlay.appendChild(splat);
        }

        fieldEl.querySelectorAll('.card').forEach(card => {
            card.classList.add('blood-card-hit');
            setTimeout(() => card.classList.remove('blood-card-hit'), 1400);
        });
    }, 350);

    setTimeout(() => overlay.remove(), 2200);
}

function prepareStealAnimation(previousState, newState) {
    if (newState.last_action !== 'steal') return null;
    // Only animate for the thief (the player whose turn it is)
    if (newState.turn_player !== newState.current_player) return null;

    const prevOpponents = previousState.opponents || [];
    const newOpponents = newState.opponents || [];

    for (const newOpp of newOpponents) {
        const prevOpp = prevOpponents.find(o => o.player_name === newOpp.player_name);
        if (!prevOpp) continue;
        if ((prevOpp.cards_in_hand || 0) > (newOpp.cards_in_hand || 0)) {
            const oppBoard = document.querySelector(`[data-opponent-name="${newOpp.player_name}"]`);
            if (!oppBoard) continue;
            const cardBacks = oppBoard.querySelectorAll('.opponent-hand-card');
            if (cardBacks.length === 0) continue;

            const lastCard = cardBacks[cardBacks.length - 1];
            const rect = lastCard.getBoundingClientRect();
            const clone = lastCard.cloneNode(true);

            return { clone, rect };
        }
    }
    return null;
}

// Play steal animation: card-back flies from victim's hand to player's hand
function playStealAnimation(stealData) {
    if (!stealData) return;

    const { clone, rect } = stealData;
    const playerHand = document.getElementById('player-hand');
    if (!playerHand) return;
    const targetRect = playerHand.getBoundingClientRect();

    const dx = (targetRect.left + targetRect.width / 2) - (rect.left + rect.width / 2);
    const dy = (targetRect.top + targetRect.height / 2) - (rect.top + rect.height / 2);

    const ghost = document.createElement('div');
    ghost.className = 'steal-card-ghost';
    ghost.style.left = rect.left + 'px';
    ghost.style.top = rect.top + 'px';
    ghost.style.width = rect.width + 'px';
    ghost.style.height = rect.height + 'px';
    ghost.style.setProperty('--dx', dx + 'px');
    ghost.style.setProperty('--dy', dy + 'px');

    clone.style.width = '100%';
    clone.style.height = '100%';
    clone.style.margin = '0';
    ghost.appendChild(clone);

    document.body.appendChild(ghost);
    setTimeout(() => ghost.remove(), 1300);
}

function prepareSabotageAnimation(previousState, newState) {
    if (newState.last_action !== 'sabotage') return null;

    const prevOpponents = previousState.opponents || [];
    const newOpponents = newState.opponents || [];

    for (const newOpp of newOpponents) {
        const prevOpp = prevOpponents.find(o => o.player_name === newOpp.player_name);
        if (!prevOpp) continue;
        if ((prevOpp.cards_in_hand || 0) > (newOpp.cards_in_hand || 0)) {
            const oppBoard = document.querySelector(`[data-opponent-name="${newOpp.player_name}"]`);
            if (!oppBoard) continue;
            const cardBacks = oppBoard.querySelectorAll('.opponent-hand-card');
            if (cardBacks.length === 0) continue;

            const lastCard = cardBacks[cardBacks.length - 1];
            const rect = lastCard.getBoundingClientRect();
            const clone = lastCard.cloneNode(true);

            return { clone, rect };
        }
    }
    return null;
}

// Play sabotage animation: card-back flies from victim's hand to discard pile
function playSabotageAnimation(sabotageData) {
    if (!sabotageData) return;

    const { clone, rect } = sabotageData;
    const pile = document.getElementById('discard-pile');
    if (!pile) return;
    const targetRect = pile.getBoundingClientRect();

    const dx = (targetRect.left + targetRect.width / 2) - (rect.left + rect.width / 2);
    const dy = (targetRect.top + targetRect.height / 2) - (rect.top + rect.height / 2);

    const ghost = document.createElement('div');
    ghost.className = 'sabotage-card-ghost';
    ghost.style.left = rect.left + 'px';
    ghost.style.top = rect.top + 'px';
    ghost.style.width = rect.width + 'px';
    ghost.style.height = rect.height + 'px';
    ghost.style.setProperty('--dx', dx + 'px');
    ghost.style.setProperty('--dy', dy + 'px');

    clone.style.width = '100%';
    clone.style.height = '100%';
    clone.style.margin = '0';
    ghost.appendChild(clone);

    document.body.appendChild(ghost);
    setTimeout(() => ghost.remove(), 1200);
}

// ── Treason Animation ──────────────────────────────────────────────────────

// Captures the warrior card element in the enemy field before re-render.
function prepareTreasonAnimation(previousState, newState) {
    if (newState.last_action !== 'treason') return null;
    // Only animate from the attacker's perspective
    if (newState.turn_player !== newState.current_player) return null;

    const prevOpponents = previousState.opponents || [];
    const newOpponents = newState.opponents || [];

    for (const prevOpp of prevOpponents) {
        const newOpp = newOpponents.find(o => o.player_name === prevOpp.player_name);
        if (!newOpp) continue;
        const prevCount = (prevOpp.field || []).length;
        const newCount = (newOpp.field || []).length;
        if (prevCount > newCount) {
            // The warrior left this opponent's field
            const oppBoard = document.querySelector(`[data-opponent-name="${prevOpp.player_name}"]`);
            if (!oppBoard) continue;
            const fieldCards = oppBoard.querySelectorAll('.opponent-field .card');
            if (fieldCards.length === 0) continue;
            const lastCard = fieldCards[fieldCards.length - 1];
            const rect = lastCard.getBoundingClientRect();
            const clone = lastCard.cloneNode(true);
            return { clone, rect, fromPlayer: prevOpp.player_name };
        }
    }
    return null;
}

// Warrior ghost flies from the enemy field to the player's own field with a flip + glow.
function playTreasonAnimation(data) {
    if (!data) return;

    const { clone, rect } = data;
    const playerField = document.getElementById('player-field');
    if (!playerField) return;
    const targetRect = playerField.getBoundingClientRect();

    const dx = (targetRect.left + targetRect.width / 2) - (rect.left + rect.width / 2);
    const dy = (targetRect.top + targetRect.height / 2) - (rect.top + rect.height / 2);

    const ghost = document.createElement('div');
    ghost.className = 'treason-ghost';
    ghost.style.left = rect.left + 'px';
    ghost.style.top = rect.top + 'px';
    ghost.style.width = rect.width + 'px';
    ghost.style.height = rect.height + 'px';
    ghost.style.setProperty('--dx', dx + 'px');
    ghost.style.setProperty('--dy', dy + 'px');

    clone.style.width = '100%';
    clone.style.height = '100%';
    clone.style.margin = '0';
    ghost.appendChild(clone);
    document.body.appendChild(ghost);

    // After the warrior lands, flash the player-field green
    setTimeout(() => {
        ghost.remove();
        playerField.classList.add('treason-landing-flash');
        setTimeout(() => playerField.classList.remove('treason-landing-flash'), 700);
    }, 1300);
}

// Detect deck count decrease for draw animation
function detectDeckDraw(previousState, newState) {
    if (!previousState) return null;
    const prevDeckCount = previousState.cards_in_deck || 0;
    const newDeckCount = newState.cards_in_deck || 0;
    const newCards = newState.new_cards || [];

    // Deck decreased and we received new cards (we're the one who drew/bought/traded)
    if (prevDeckCount > newDeckCount && newCards.length > 0) {
        return { count: Math.min(prevDeckCount - newDeckCount, 3) };
    }
    return null;
}

// Play deck draw animation: card-backs fly from deck to player's hand
function playDeckDrawAnimation(drawInfo) {
    if (!drawInfo) return;

    const deckElement = document.getElementById('deck');
    if (!deckElement) return;
    const deckRect = deckElement.getBoundingClientRect();

    const playerHand = document.getElementById('player-hand');
    if (!playerHand) return;
    const handRect = playerHand.getBoundingClientRect();

    const dx = (handRect.left + handRect.width / 2) - (deckRect.left + deckRect.width / 2);
    const dy = (handRect.top + handRect.height / 2) - (deckRect.top + deckRect.height / 2);

    for (let i = 0; i < drawInfo.count; i++) {
        setTimeout(() => {
            const ghost = document.createElement('div');
            ghost.className = 'deck-draw-ghost';
            ghost.style.left = deckRect.left + 'px';
            ghost.style.top = deckRect.top + 'px';
            ghost.style.width = (deckRect.width * 0.6) + 'px';
            ghost.style.height = (deckRect.height * 0.6) + 'px';
            ghost.style.setProperty('--dx', dx + 'px');
            ghost.style.setProperty('--dy', dy + 'px');

            ghost.innerHTML = `
                <div class="card-back" style="width:100%;height:100%;">
                    <div class="card-back-design">
                        <div class="card-back-border">
                            <div class="card-back-inner">
                                <div class="card-back-pattern"></div>
                                <div class="card-back-emblem">\u2694</div>
                            </div>
                        </div>
                    </div>
                </div>
            `;

            document.body.appendChild(ghost);
            setTimeout(() => ghost.remove(), 900);
        }, i * 150);
    }
}

// Detect discard pile and cemetery count changes
function detectPileAndCemeteryChanges(previousState, newState) {
    const changes = { pileAdded: false, cemeteryAdded: false };

    const prevPile = previousState.discard_pile || {};
    const newPile = newState.discard_pile || {};
    if ((newPile.cards || 0) > (prevPile.cards || 0)) {
        changes.pileAdded = true;
    }

    const prevCem = previousState.cemetery || {};
    const newCem = newState.cemetery || {};
    if ((newCem.corps || 0) > (prevCem.corps || 0)) {
        changes.cemeteryAdded = true;
    }

    return changes;
}

// Show discard pile flash and card appear animation
function showPileAnimation() {
    const pileElement = document.getElementById('discard-pile');
    if (!pileElement) return;

    pileElement.classList.add('pile-flash');
    setTimeout(() => pileElement.classList.remove('pile-flash'), 1500);

    const lastCard = document.querySelector('#discard-pile-last-card .card');
    if (lastCard) {
        lastCard.classList.add('pile-new-card');
        setTimeout(() => lastCard.classList.remove('pile-new-card'), 1200);
    }
}

// Show cemetery flash and corpse appear animation
function showCemeteryAnimation() {
    const cemElement = document.getElementById('cemetery');
    if (!cemElement) return;

    cemElement.classList.add('cemetery-flash');
    setTimeout(() => cemElement.classList.remove('cemetery-flash'), 1500);

    const lastCorp = document.querySelector('#cemetery-last-corp .card');
    if (lastCorp) {
        lastCorp.classList.add('cemetery-new-corp');
        setTimeout(() => lastCorp.classList.remove('cemetery-new-corp'), 1200);
    }
}

function showResurrectionAnimation() {
    const cemElement = document.getElementById('cemetery');
    if (!cemElement) return;

    cemElement.classList.add('cemetery-resurrection');
    setTimeout(() => cemElement.classList.remove('cemetery-resurrection'), 2800);

    const overlay = document.createElement('div');
    overlay.className = 'resurrection-overlay';
    overlay.textContent = '✨ RESURRECTED ✨';
    cemElement.appendChild(overlay);
    setTimeout(() => overlay.remove(), 2400);
}

function showMercenaryHiredAnimation() {
    const fieldEl = document.getElementById('player-field');
    if (!fieldEl) return;

    // Flash the field border
    fieldEl.classList.add('field-hired-flash');
    setTimeout(() => fieldEl.classList.remove('field-hired-flash'), 1200);

    // Find the newly added mercenary card (last warrior card in the field)
    const mercenaryCard = fieldEl.querySelector('.card.warrior:last-child');
    if (mercenaryCard) {
        mercenaryCard.classList.add('mercenary-entrance');
        setTimeout(() => mercenaryCard.classList.remove('mercenary-entrance'), 900);
    }

    // Floating "Mercenary Hired!" text
    const text = document.createElement('div');
    text.className = 'mercenary-hired-text';
    text.textContent = '⚔️ Mercenary Hired!';
    fieldEl.style.position = 'relative';
    fieldEl.appendChild(text);
    setTimeout(() => text.remove(), 1600);
}

// Ambush placed animation — shown to all players when a player places an ambush card
function showAmbushPlacedAnimation(gameStatus) {
    // ambush_placed_on identifies whose field received the ambush (handles ally placement in 2v2)
    const targetPlayer = gameStatus.ambush_placed_on || gameStatus.turn_player;

    let fieldEl;
    if (targetPlayer === gameStatus.current_player) {
        fieldEl = document.getElementById('player-field');
    } else {
        const opponentFields = document.querySelectorAll('.opponent-field');
        for (const f of opponentFields) {
            if (f.dataset.opponentName === targetPlayer) {
                fieldEl = f;
                break;
            }
        }
    }
    if (!fieldEl) return;

    // Ominous purple flash on the field
    fieldEl.classList.add('ambush-placed-flash');
    setTimeout(() => fieldEl.classList.remove('ambush-placed-flash'), 1500);

    // Card slams in from above
    const ambushCard = fieldEl.querySelector('.card.ambush-face-down');
    if (ambushCard) {
        ambushCard.classList.add('ambush-card-entrance');
        setTimeout(() => ambushCard.classList.remove('ambush-card-entrance'), 1000);
    }

    // Floating label
    const text = document.createElement('div');
    text.className = 'ambush-placed-text';
    text.textContent = '⚠ Ambush Set!';
    fieldEl.style.position = 'relative';
    fieldEl.appendChild(text);
    setTimeout(() => text.remove(), 1800);
}

// Elimination detection
function checkForEliminations(previousState, newState) {
    const prevOpponents = previousState.opponents || [];
    const newOpponents = newState.opponents || [];

    for (const newOpp of newOpponents) {
        if (!newOpp.is_eliminated) continue;
        const prevOpp = prevOpponents.find(o => o.player_name === newOpp.player_name);
        if (prevOpp && !prevOpp.is_eliminated) {
            // This player was just eliminated
            const isSelf = newOpp.player_name === gameState.playerName;
            setTimeout(() => showEliminationModal(newOpp.player_name, isSelf), 200);
            return; // Show one at a time
        }
    }
}
