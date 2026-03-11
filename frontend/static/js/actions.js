function startAction(actionType) {
    resetActionState();
    gameState.currentAction = actionType;
    gameState.actionState.type = actionType;

    updateActionPrompt('');
    showConfirmButtons(); // Show cancel button

    // Re-render board to apply action-specific styles
    if (gameState.currentState) {
        renderGameBoard(gameState.currentState);
    }
}

function cancelAction() {
    resetActionState();
    updateActionPrompt('');
    // Re-render board to recalculate usable/unusable classes for the current phase
    if (gameState.currentState) {
        renderGameBoard(gameState.currentState);
    }
}

// Card selection
function handleCardClick(cardID, cardType, context, card = null) {
    console.log('Card clicked:', { cardID, cardType, context, isProtected: card?.protected_by?.id });

    if (!gameState.isYourTurn) return;

    const action = gameState.currentAction;
    const status = gameState.currentState;

    // Handle move_warrior: clicking a warrior in hand when move is available
    if (context === 'player-hand' && cardType === 'warrior' && status?.can_move_warrior && !action) {
        clearSelections();
        gameState.actionState.warriorId = cardID;
        highlightSelectedCard(cardID);

        const warriorCard = findCardById(cardID);
        const cardHtml = renderCardForModal(warriorCard);

        // In 2v2 mode, offer choice between own field and ally's field
        const allies = (gameState.currentState?.opponents || []).filter(o => o.is_ally && !o.is_eliminated);
        if (gameState.gameMode === '2v2' && allies.length > 0) {
            showMoveWarriorTargetModal(cardID, warriorCard, cardHtml, allies);
        } else {
            showActionConfirmModal({
                title: 'Move Warrior',
                cardsHtml: cardHtml,
                description: `${getCardName(warriorCard)} will move to your field`,
                onConfirm: () => {
                    sendAction('move_warrior', { warrior_id: cardID });
                    resetActionState();
                }
            });
        }
        return;
    }

    // Handle trade action
    if (action === 'trade' && context === 'player-hand') {
        if (!card?.can_be_traded) return;
        toggleCardSelection(cardID, 'player-hand');
        if (gameState.selectedCards.length === 3) {
            // Show trade confirmation popup with 3 selected cards + 1 card back
            showTradeConfirmModal();
        } else {
            updateActionPrompt(`Selected ${gameState.selectedCards.length}/3 cards for trade`);
        }
        return;
    }

    // Handle forge action
    if (action === 'forge' && context === 'player-hand') {
        handleForgeCardClick(cardID, card);
        return;
    }

    // Handle trade: clicking a tradeable non-usable card in hand starts trade mode
    if (context === 'player-hand' && card?.can_be_traded && !card?.can_be_used && !action && status?.can_trade) {
        startTradeFromCard(cardID);
        return;
    }

    // Handle attack phase card selection from hand
    if (status && status.current_action === 'attack' && context === 'player-hand') {
        handleAttackPhaseHandClick(cardID, card);
        return;
    }

    // Handle spy/steal phase card selection from hand
    if (status && status.current_action === 'spy/steal' && context === 'player-hand') {
        handleSpyStealPhaseHandClick(cardID, card);
        return;
    }

    // Handle buy phase card selection from hand
    if (status && status.current_action === 'buy' && context === 'player-hand') {
        handleBuyPhaseHandClick(cardID, card);
        return;
    }

    // Handle construct phase card selection from hand
    if (status && status.current_action === 'construct' && context === 'player-hand') {
        handleConstructPhaseHandClick(cardID, card);
        return;
    }

    // Handle target selection for attack phase (opponent field)
    if (gameState.actionState.weaponId && context.startsWith('opponent-field:')) {
        const opponentName = context.split(':')[1];
        const opponent = getOpponentByName(opponentName);
        const isAllyBoard = opponent?.is_ally || false;

        if (gameState.actionState.type === 'specialpower') {
            const userId = gameState.actionState.userId;
            if (userId) {
                const user = findCardById(userId);
                const userType = (user?.sub_type || '').toLowerCase();
                if (userType === 'archer') {
                    // Archer (Instant Kill) can only target NON-ally opponents
                    if (isAllyBoard) return;
                } else {
                    // Mage (Heal) and Knight (Protect) can only target ally boards
                    if (!isAllyBoard) return;
                }
            } else {
                // Must select a user warrior from your field first
                return;
            }
        } else {
            // Regular attacks cannot target allies
            if (isAllyBoard) return;
            // For Sword/Arrow/Poison: must select attacker warrior first
            if (gameState.actionState.type === 'attack' && !gameState.actionState.warriorId) {
                const weapon = findCardById(gameState.actionState.weaponId);
                if (WEAPON_ATTACKER_TYPES[(weapon?.sub_type || '').toLowerCase()]) return;
            }
        }

        gameState.actionState.targetPlayer = opponentName;
        handleAttackPhaseTargetClick(cardID, isAllyBoard ? 'ally' : 'enemy');
        return;
    }

    // Handle attacker warrior selection for regular-weapon attacks (player field)
    if (gameState.actionState.weaponId && gameState.actionState.type === 'attack' &&
        !gameState.actionState.warriorId && context === 'player-field') {
        handleAttackPhaseAttackerClick(cardID);
        return;
    }

    // Handle user warrior selection for special power (player field)
    if (gameState.actionState.weaponId && gameState.actionState.type === 'specialpower' &&
        !gameState.actionState.userId && context === 'player-field') {
        handleAttackPhaseUserClick(cardID);
        return;
    }

    // Handle target selection for special power on own field (Mage heal, Knight protect)
    if (gameState.actionState.weaponId && gameState.actionState.type === 'specialpower' &&
        gameState.actionState.userId && context === 'player-field') {
        // Re-click the selected user warrior = deselect, UNLESS it is also a valid target
        if (cardID === gameState.actionState.userId) {
            const weapon = findCardById(gameState.actionState.weaponId);
            const isValidTarget = weapon?.use_on?.includes(cardID);
            if (!isValidTarget) {
                document.querySelector(`[data-card-id="${cardID}"]`)?.classList.remove('selected-user');
                gameState.actionState.userId = null;
                clearSpecialPowerTargetSelection();
                updateActionPrompt(`✨ ${getCardName(weapon)} - Select a warrior from your field to use it`);
                return;
            }
            // Valid target (e.g. Knight self-protect, Mage self-heal) — fall through to target selection
        }
        const userId = gameState.actionState.userId;
        const user = findCardById(userId);
        const userType = (user?.sub_type || '').toLowerCase();
        // Only Mage and Knight can target own field
        if (userType === 'archer') {
            return; // Archer can only target enemies
        }
        handleAttackPhaseTargetClick(cardID, 'player');
        return;
    }
}

function highlightSelectedCard(cardID) {
    const card = document.querySelector(`[data-card-id="${cardID}"]`);
    if (card) {
        card.classList.add('selected');
    }
}

function clearSelections() {
    gameState.selectedCards = [];
    document.querySelectorAll('.card.selected').forEach(card => {
        card.classList.remove('selected');
    });
    document.querySelectorAll('.card.selected-user').forEach(card => {
        card.classList.remove('selected-user');
    });
    // Remove valid-target highlights and multiplier badges
    document.querySelectorAll('.card.valid-target').forEach(card => {
        card.classList.remove('valid-target');
    });
    // Remove selection mode classes from fields
    const playerField = document.getElementById('player-field');
    playerField?.classList.remove('selecting-ally');
    playerField?.classList.remove('targeting-mode');
    document.querySelectorAll('.opponent-field').forEach(f => {
        f.classList.remove('selecting-target');
        f.classList.remove('selecting-ally');
        f.classList.remove('targeting-mode');
    });
    document.querySelectorAll('.dmg-multiplier-badge').forEach(badge => {
        badge.remove();
    });
}

function toggleCardSelection(cardID, context) {
    const index = gameState.selectedCards.indexOf(cardID);
    if (index > -1) {
        gameState.selectedCards.splice(index, 1);
    } else {
        gameState.selectedCards.push(cardID);
    }

    // Update visual selection
    const container = document.getElementById('player-hand');

    const cardElement = container.querySelector(`[data-card-id="${cardID}"]`);
    if (cardElement) {
        cardElement.classList.toggle('selected');
    }
}

function resetActionState() {
    gameState.currentAction = null;
    gameState.selectedCards = [];
    gameState.actionState = {
        type: null,
        weaponId: null,
        userId: null,
        targetId: null,
        warriorId: null,
        targetPlayer: null
    };

    // Clear visual selections
    document.querySelectorAll('.card.selected, .card.valid-target, .card.selected-user').forEach(card => {
        card.classList.remove('selected', 'valid-target', 'selected-user');
    });

    // Remove selection mode classes from fields
    const playerField = document.getElementById('player-field');
    playerField?.classList.remove('selecting-ally', 'targeting-mode');
    document.querySelectorAll('.opponent-field').forEach(f => {
        f.classList.remove('selecting-target', 'selecting-ally', 'targeting-mode');
    });

    // Remove damage multiplier badges
    document.querySelectorAll('.dmg-multiplier-badge').forEach(badge => {
        badge.remove();
    });

    hideConfirmButtons();
}

function showConfirmButtons() {
    // Show the action prompt container (which includes cancel button)
    document.getElementById('action-prompt-container').classList.remove('hidden');
}

function hideConfirmButtons() {
    // Hide the action prompt container
    document.getElementById('action-prompt-container').classList.add('hidden');
}

// Attack phase handlers
function handleAttackPhaseHandClick(cardID, card) {
    // Clicking the already-selected weapon deselects it
    if (gameState.actionState.weaponId === cardID) {
        resetActionState();
        return;
    }

    // Check if card can be used
    if (card && card.can_be_used === false) {
        return; // Card cannot be used in this phase
    }

    // Determine weapon type from card data
    const cardType = card ? (card.type || '').toLowerCase() : '';

    // Clear previous selections but keep action active
    clearSelections();
    gameState.actionState.weaponId = cardID;

    if (cardType === 'specialpower') {
        gameState.actionState.type = 'specialpower';
        highlightSelectedCard(cardID);
        const powerName = getCardName(card);
        updateActionPrompt(`✨ ${powerName} - Select a warrior from your field to use it`);
        highlightValidUserWarriors(card);
        showConfirmButtons();
    } else if (cardType === 'catapult') {
        gameState.actionState.type = 'catapult';
        gameState.actionState.weaponId = cardID;
        highlightSelectedCard(cardID);
        const enemies = getEnemyOpponents().filter(e => e.castle?.constructed && (e.castle?.resource_cards > 0 || e.castle?.is_protected));
        if (enemies.length === 1) {
            gameState.actionState.targetPlayer = enemies[0].player_name;
            showCatapultModal();
        } else if (enemies.length > 1) {
            showTargetPlayerModal('Select a castle to attack', enemies, (playerName) => {
                gameState.actionState.weaponId = cardID;
                gameState.actionState.targetPlayer = playerName;
                showCatapultModal();
            });
        } else {
            updateActionPrompt('No enemy castles to attack!');
            resetActionState();
        }
    } else if (cardType === 'weapon') {
        gameState.actionState.type = 'attack';
        highlightSelectedCard(cardID);
        const weaponName = getCardName(card);
        const weaponDmg = card?.value || 0;
        const weaponSubType = (card?.sub_type || '').toLowerCase();
        if (WEAPON_ATTACKER_TYPES[weaponSubType]) {
            // Sword/Arrow/Poison: select attacker warrior first
            updateActionPrompt(`⚔️ ${weaponName} (${weaponDmg} DMG) - Select a warrior to attack with`);
            highlightValidAttackerWarriors(card);
        } else {
            // Other weapons: go directly to target selection
            updateActionPrompt(`⚔️ ${weaponName} (${weaponDmg} DMG) - Select a target`);
            highlightValidTargets(card);
        }
        showConfirmButtons();
    } else if (cardType === 'harpoon') {
        gameState.actionState.type = 'harpoon';
        highlightSelectedCard(cardID);
        updateActionPrompt(`🎯 Harpoon - Select a Dragon to kill`);
        highlightValidTargets(card);
        showConfirmButtons();
    } else if (cardType === 'bloodrain') {
        gameState.actionState.type = 'bloodrain';
        gameState.actionState.weaponId = cardID;
        highlightSelectedCard(cardID);
        const enemies = getEnemyOpponents().filter(e => (e.field?.length ?? 0) > 0 && !e.is_eliminated);
        if (enemies.length === 0) {
            updateActionPrompt('No enemies with warriors to target!');
            resetActionState();
        } else if (enemies.length === 1) {
            gameState.actionState.targetPlayer = enemies[0].player_name;
            showBloodRainConfirmModal(card, enemies[0]);
        } else {
            showTargetPlayerModal('Select a field to drench in Blood Rain', enemies, (playerName) => {
                const enemy = getOpponentByName(playerName);
                gameState.actionState.weaponId = cardID;
                gameState.actionState.targetPlayer = playerName;
                showBloodRainConfirmModal(card, enemy || { player_name: playerName, field: [] });
            }, (opp) => {
                const field = opp.field || [];
                const count = field.length;
                const hpList = field.map(w => `${w.value || '?'}HP`).join(', ');
                const hpSuffix = hpList ? ` — ${hpList}` : '';
                return `${count} warrior${count !== 1 ? 's' : ''} on field${hpSuffix}`;
            });
        }
    } else if (cardType === 'resurrection') {
        gameState.actionState.type = 'resurrection';
        gameState.actionState.weaponId = cardID;
        highlightSelectedCard(cardID);
        const allies = (gameState.currentState?.opponents || []).filter(o => o.is_ally && !o.is_eliminated);
        if (allies.length > 0) {
            showResurrectionTargetModal(cardID, card, allies);
        } else {
            showResurrectionConfirmModal(cardID, card, '');
        }
    } else if (cardType === 'ambush') {
        gameState.actionState.type = 'ambush';
        gameState.actionState.weaponId = cardID;
        highlightSelectedCard(cardID);
        const ambushAllies = (gameState.currentState?.opponents || []).filter(o => o.is_ally && !o.is_eliminated && !o.ambush_in_field);
        const ownHasAmbush = !!gameState.currentState?.current_player_ambush_in_field;
        if (ambushAllies.length > 0) {
            showAmbushTargetModal(card, cardID, ambushAllies, !ownHasAmbush);
        } else {
            showAmbushPlaceConfirmModal(card, cardID, '');
        }
    } else if (cardType === 'treason') {
        gameState.actionState.type = 'treason';
        gameState.actionState.weaponId = cardID;
        highlightSelectedCard(cardID);
        const enemies = getEnemyOpponents().filter(opp => (opp.field || []).some(w => w.value <= 5));
        if (enemies.length === 0) {
            updateActionPrompt('No enemies have warriors with 5 HP or less!');
            resetActionState();
            return;
        }
        if (enemies.length === 1) {
            gameState.actionState.targetPlayer = enemies[0].player_name;
            showTreasonModal();
        } else {
            showTargetPlayerModal('Select a player to steal warrior from', enemies,
                (playerName) => {
                    gameState.actionState.weaponId = cardID;
                    gameState.actionState.targetPlayer = playerName;
                    showTreasonModal();
                },
                (opp) => {
                    const weakCount = (opp.field || []).filter(w => w.value <= 5).length;
                    return weakCount > 0 ? `${weakCount} weak warrior(s) (≤5 HP)` : 'No warriors ≤5 HP';
                });
        }
    }
}

function handleAttackPhaseUserClick(cardID) {
    const warrior = findCardById(cardID);
    const warriorType = (warrior?.sub_type || '').toLowerCase();
    if (!SPECIAL_POWER_USER_TYPES.includes(warriorType)) return;

    // User selected a warrior to use the special power
    gameState.actionState.userId = cardID;
    highlightUserWarrior(cardID);

    // Remove attacker selection state from own field
    const playerField = document.getElementById('player-field');
    playerField?.classList.remove('selecting-ally', 'targeting-mode');
    playerField?.querySelectorAll('.card.valid-target').forEach(c => c.classList.remove('valid-target'));

    // Get the weapon and user cards for display
    const weapon = findCardById(gameState.actionState.weaponId);
    const user = findCardById(cardID);
    const userName = getCardName(user);

    // Get the effect and target info based on warrior type
    const userType = (user?.sub_type || '').toLowerCase();
    let effect = '';
    let emoji = '✨';
    let targetHint = 'Select a target';

    switch (userType) {
        case 'archer':
            effect = 'INSTANT KILL';
            emoji = '🎯';
            targetHint = 'Select an enemy to kill';
            break;
        case 'knight':
            effect = 'PROTECT';
            emoji = '🛡️';
            targetHint = 'Select an ally to protect';
            break;
        case 'mage':
            effect = 'HEAL';
            emoji = '💚';
            targetHint = 'Select an ally to heal';
            break;
        default:
            effect = getCardName(weapon);
    }

    updateActionPrompt(`${emoji} ${userName} will use ${effect} - ${targetHint}`);

    // Highlight valid targets for this specific warrior type, then enable field selection
    highlightSpecialPowerTargets(userType, weapon);
    enableSpecialPowerTargetSelection(userType);
}

function handleAttackPhaseAttackerClick(cardID) {
    const weapon = findCardById(gameState.actionState.weaponId);
    const warrior = findCardById(cardID);
    const weaponSubType = (weapon?.sub_type || '').toLowerCase();
    const validTypes = WEAPON_ATTACKER_TYPES[weaponSubType] || [];
    const warriorType = (warrior?.sub_type || '').toLowerCase();

    // Validate warrior-weapon compatibility
    if (validTypes.length > 0 && !validTypes.includes(warriorType)) return;

    gameState.actionState.warriorId = cardID;
    highlightUserWarrior(cardID);

    // Remove attacker selection state from own field
    const playerField = document.getElementById('player-field');
    playerField?.classList.remove('selecting-ally');
    playerField?.querySelectorAll('.card.valid-target').forEach(c => c.classList.remove('valid-target'));

    const warriorName = getCardName(warrior);
    const weaponName = getCardName(weapon);
    const weaponDmg = weapon?.value || 0;
    updateActionPrompt(`⚔️ ${warriorName} attacks with ${weaponName} (${weaponDmg} DMG) - Select a target`);

    // Now highlight valid enemy targets
    highlightValidTargets(weapon);
}

function handleAttackPhaseTargetClick(cardID, side) {
    // Check if this is a valid target
    const weapon = findCardById(gameState.actionState.weaponId);
    console.log('Attack target click:', { cardID, side, weaponUseOn: weapon?.use_on, isValidTarget: weapon?.use_on?.includes(cardID) });

    if (!weapon || !weapon.use_on || !weapon.use_on.includes(cardID)) {
        console.log('Target rejected - not in use_on list');
        return; // Not a valid target
    }

    gameState.actionState.targetId = cardID;
    highlightSelectedCard(cardID);

    const actionType = gameState.actionState.type;
    const target = findCardById(cardID);

    if (actionType === 'attack') {
        // Show attack confirmation popup
        showAttackConfirmModal(weapon, target);
    } else if (actionType === 'specialpower') {
        // Show special power confirmation popup
        const user = findCardById(gameState.actionState.userId);
        showSpecialPowerConfirmModal(weapon, user, target);
    } else if (actionType === 'harpoon') {
        showHarpoonConfirmModal(weapon, target);
    }
}

function showHarpoonConfirmModal(weapon, target) {
    const targetName = getCardName(target);

    let cardsHtml = renderCardForModal(weapon);
    cardsHtml += renderArrow();
    cardsHtml += renderCardForModal(target);

    showActionConfirmModal({
        title: 'Harpoon',
        cardsHtml: cardsHtml,
        description: `🎯 Harpoon → ${targetName} <span class="hp-preview hp-fatal">💀 INSTANT KILL</span>`,
        onConfirm: () => {
            sendAction('harpoon', {
                target_player: gameState.actionState.targetPlayer,
                weapon_id: gameState.actionState.weaponId,
                target_id: gameState.actionState.targetId
            });
            resetActionState();
        }
    });
}

function showBloodRainConfirmModal(weapon, targetOpponent) {
    const targetName = targetOpponent?.player_name || gameState.actionState.targetPlayer;
    const targetField = targetOpponent?.field || [];
    const dmg = weapon?.value || 4;

    let cardsHtml = renderCardForModal(weapon);
    cardsHtml += renderArrow();
    targetField.forEach(c => { cardsHtml += renderCardForModal(c); });
    if (targetField.length === 0) {
        cardsHtml += `<div class="card card-placeholder">🗡️ All Warriors</div>`;
    }

    const warriorCount = targetField.length;
    const warriorSummary = warriorCount === 1
        ? `1 warrior (${dmg} DMG)`
        : warriorCount > 1 ? `${warriorCount} warriors (${dmg} DMG each)` : `all warriors (${dmg} DMG each)`;
    const hpParts = targetField.map(w => `${getCardName(w)} <span class="hp-preview">${w.value || '?'}HP</span>`).join(' · ');
    const hpSuffix = hpParts ? ` — ${hpParts}` : '';

    showActionConfirmModal({
        title: 'Blood Rain',
        cardsHtml: cardsHtml,
        description: `🩸 Blood Rain hits all of ${targetName}'s warriors — ${warriorSummary}${hpSuffix}`,
        onConfirm: () => {
            sendAction('blood_rain', {
                target_player: gameState.actionState.targetPlayer,
                weapon_id: gameState.actionState.weaponId
            });
            resetActionState();
        }
    });
}

// Returns the flat Curse damage modifier for the given weapon, or 0 if not affected.
function getCurseWeaponModifier(weapon) {
    const gs = gameState.currentState;
    if (!gs || gs.current_event !== 'curse' || !gs.current_event_weapon_modifier) return 0;
    const subType = weapon?.sub_type || '';
    if (subType === gs.current_event_excluded_weapon) return 0;
    if (!['Sword', 'Arrow', 'Poison'].includes(subType)) return 0;
    return gs.current_event_weapon_modifier;
}

function showAttackConfirmModal(weapon, target) {
    const weaponName = getCardName(weapon);
    const weaponDmg = weapon?.value || 0;
    const targetName = getCardName(target);
    const targetHp = target?.value || 0;
    const targetId = target?.id;
    const multiplier = weapon?.dmg_mult?.[targetId] || 1;
    const hasDoubleDamage = multiplier > 1;

    const attacker = findCardById(gameState.actionState.warriorId);
    const killBonus = attacker?.kills || 0;

    const curseModifier = getCurseWeaponModifier(weapon);
    const effectiveWeaponDmg = Math.max(0, weaponDmg + curseModifier + killBonus);
    const effectiveDmg = effectiveWeaponDmg * multiplier;

    const isProtected = target?.protected_by && target.protected_by.id;
    const shieldHp = isProtected ? (target.protected_by.value || 0) : 0;

    let cardsHtml = '';
    if (attacker) {
        cardsHtml += renderCardForModal(attacker, { killBonus: killBonus });
        cardsHtml += renderArrow();
    }
    cardsHtml += renderCardForModal(weapon, { showDoubleDamage: hasDoubleDamage });
    cardsHtml += renderArrow();
    cardsHtml += renderCardForModal(target, { showShield: isProtected, shieldHp: shieldHp });

    let description;
    let dmgLabel;
    const curseSign = curseModifier > 0 ? '+' : '';
    const killSign = killBonus > 0 ? `+${killBonus}💀` : '';
    const hasModifiers = curseModifier !== 0 || killBonus > 0;
    if (hasDoubleDamage && hasModifiers) {
        const modStr = (curseModifier !== 0 ? `${curseSign}${curseModifier}` : '') + killSign;
        dmgLabel = `${weaponName} (${weaponDmg}${modStr}=${effectiveWeaponDmg} x${multiplier} = ${effectiveDmg} DMG)`;
    } else if (hasDoubleDamage) {
        dmgLabel = `${weaponName} (${weaponDmg} x${multiplier} = ${effectiveDmg} DMG)`;
    } else if (hasModifiers) {
        const modStr = (curseModifier !== 0 ? `${curseSign}${curseModifier}` : '') + killSign;
        dmgLabel = `${weaponName} (${weaponDmg}${modStr} = ${effectiveWeaponDmg} DMG)`;
    } else {
        dmgLabel = `${weaponName} (${weaponDmg} DMG)`;
    }

    if (isProtected) {
        const shieldAfter = Math.max(0, shieldHp - effectiveDmg);
        const shieldDestroyed = shieldAfter <= 0;
        const shieldPreview = shieldDestroyed
            ? `<span class="hp-preview hp-fatal">💥 DESTROYED</span>`
            : `<span class="hp-preview shield-hp">🛡️ ${shieldHp} → ${shieldAfter}</span>`;
        description = `${dmgLabel} → ${targetName}<br>` +
            `<span class="shield-info">🛡️ Shield absorbs damage — Warrior takes 0 DMG</span><br>` +
            `Shield: ${shieldPreview}`;
    } else {
        const resultingHp = Math.max(0, targetHp - effectiveDmg);
        const willDie = resultingHp <= 0;
        const hpPreview = willDie
            ? `<span class="hp-preview hp-fatal">💀 FATAL</span>`
            : `<span class="hp-preview">${targetHp} → ${resultingHp} HP</span>`;
        description = `${dmgLabel} → ${targetName} ${hpPreview}`;
    }

    showActionConfirmModal({
        title: 'Attack',
        cardsHtml: cardsHtml,
        description: description,
        onConfirm: () => {
            sendAction('attack', {
                warrior_id: gameState.actionState.warriorId,
                target_player: gameState.actionState.targetPlayer,
                weapon_id: gameState.actionState.weaponId,
                target_id: gameState.actionState.targetId
            });
            resetActionState();
        }
    });
}

function showSpecialPowerConfirmModal(specialPower, user, target) {
    const userName = getCardName(user);
    const targetName = getCardName(target);
    const targetHp = target?.value || 0;
    const userType = (user?.sub_type || '').toLowerCase();

    let title = 'Special Power';
    let description = '';
    let healedHp = 0;

    const isProtected = target?.protected_by && target.protected_by.id;
    const shieldHp = isProtected ? (target.protected_by.value || 0) : 0;

    switch (userType) {
        case 'archer':
            title = 'Instant Kill';
            if (isProtected) {
                description = `${userName} targets ${targetName}<br>` +
                    `<span class="shield-info">🛡️ Shield blocks the kill — Shield destroyed, warrior survives</span>`;
            } else {
                description = `${userName} will instantly kill ${targetName}`;
            }
            break;
        case 'knight':
            title = 'Protect';
            description = `${userName} will protect ${targetName} (${targetHp} HP)`;
            break;
        case 'mage':
            title = 'Heal';
            healedHp = (target?.sub_type || '').toLowerCase() === 'mercenary' ? 10 : 20;
            description = `${userName} will heal ${targetName} (${targetHp} → ${healedHp} HP)`;
            break;
        default:
            description = `${userName} will use ${getCardName(specialPower)} on ${targetName}`;
    }

    let cardsHtml = renderCardForModal(user);
    cardsHtml += renderArrow();
    cardsHtml += renderCardForModal(target, { showShield: isProtected, shieldHp: shieldHp, showHealResult: userType === 'mage', healedHp });

    showActionConfirmModal({
        title: title,
        cardsHtml: cardsHtml,
        description: description,
        onConfirm: () => {
            sendAction('special_power', {
                weapon_id: gameState.actionState.weaponId,
                user_id: gameState.actionState.userId,
                target_id: gameState.actionState.targetId
            });
            resetActionState();
        }
    });
}

// Trade confirmation modal
function showTradeConfirmModal() {
    // Get the 3 selected cards
    const selectedCardIds = gameState.selectedCards;
    let cardsHtml = '';

    // Render each selected card
    selectedCardIds.forEach((cardId, index) => {
        const card = findCardById(cardId);
        cardsHtml += renderCardForModal(card);
    });

    // Add arrow and 1 card back (the new card)
    cardsHtml += renderArrow();
    cardsHtml += renderCardBacks(1);

    showActionConfirmModal({
        title: 'Trade Cards',
        cardsHtml: cardsHtml,
        description: 'Trade 3 cards for 1 new card from the deck',
        onConfirm: () => {
            sendAction('trade', { card_ids: selectedCardIds });
            resetActionState();
        }
    });
}

// Spy/Steal phase handlers
function handleSpyStealPhaseHandClick(cardID, card) {
    if (card && card.can_be_used === false) {
        return;
    }

    const cardType = card ? (card.type || '').toLowerCase() : '';

    clearSelections();
    gameState.actionState.weaponId = cardID;
    highlightSelectedCard(cardID);

    if (cardType === 'spy') {
        gameState.actionState.type = 'spy';
        showSpyOptionsModal();
    } else if (cardType === 'thief') {
        gameState.actionState.type = 'thief';
        const enemies = getEnemyOpponents();
        if (enemies.length === 1) {
            // Only one enemy, skip player selection
            gameState.actionState.targetPlayer = enemies[0].player_name;
            showStealModal();
        } else {
            // Multiple enemies, show target player selection first
            showTargetPlayerModal('Select a player to steal from', enemies, (playerName) => {
                gameState.actionState.weaponId = cardID;
                gameState.actionState.targetPlayer = playerName;
                showStealModal();
            }, (opp) => `${opp.cards_in_hand} cards in hand`);
        }
    } else if (cardType === 'sabotage') {
        gameState.actionState.type = 'sabotage';
        gameState.pendingModalAction = 'sabotage';
        const enemies = getEnemyOpponents();
        if (enemies.length === 1) {
            sendAction('sabotage', { card_id: cardID, target_player: enemies[0].player_name });
        } else {
            showTargetPlayerModal('Select a player to sabotage', enemies,
                (playerName) => sendAction('sabotage', { card_id: cardID, target_player: playerName }),
                (opp) => `${opp.cards_in_hand} card(s) in hand`);
        }
    } else if (cardType === 'treason') {
        gameState.actionState.type = 'treason';
        const enemies = getEnemyOpponents().filter(opp => (opp.field || []).some(w => w.value <= 5));
        if (enemies.length === 0) {
            updateActionPrompt('No enemies have warriors with 5 HP or less!');
            resetActionState();
            return;
        }
        if (enemies.length === 1) {
            gameState.actionState.targetPlayer = enemies[0].player_name;
            showTreasonModal();
        } else {
            showTargetPlayerModal('Select a player to steal warrior from', enemies,
                (playerName) => {
                    gameState.actionState.weaponId = cardID;
                    gameState.actionState.targetPlayer = playerName;
                    showTreasonModal();
                },
                (opp) => {
                    const weakCount = (opp.field || []).filter(w => w.value <= 5).length;
                    return weakCount > 0 ? `${weakCount} weak warrior(s) (≤5 HP)` : 'No warriors ≤5 HP';
                });
        }
    }
}

// Buy phase handlers
function handleBuyPhaseHandClick(cardID, card) {
    if (card && card.can_be_used === false) {
        return;
    }

    // Ambush card: show confirmation modal before placing in field
    if (card && (card.type === 'Ambush' || card.sub_type === 'Ambush')) {
        showAmbushPlaceConfirmModal(card, cardID);
        return;
    }

    clearSelections();
    gameState.actionState.weaponId = cardID;
    gameState.actionState.type = 'buy';
    highlightSelectedCard(cardID);

    // Show buy confirmation popup with resource and card backs
    showBuyConfirmModal(card, cardID);
}

function showBuyConfirmModal(resource, cardID) {
    const resourceValue = resource?.value || 0;
    const cardsToReceive = Math.floor(resourceValue / 2);

    // When gold >= 8, offer a choice: deck cards or mercenary
    if (resourceValue >= 8) {
        const content = `
            <div class="target-player-options">
                <div class="target-player-option" onclick="window._buyChoiceCallback('deck')">
                    <span class="player-icon">🃏</span>
                    <div class="player-info">
                        <div class="player-name">Buy from Deck</div>
                        <div class="player-detail">Receive ${cardsToReceive} card${cardsToReceive !== 1 ? 's' : ''}</div>
                    </div>
                </div>
                <div class="target-player-option" onclick="window._buyChoiceCallback('mercenary')">
                    <span class="player-icon">⚔️</span>
                    <div class="player-info">
                        <div class="player-name">Hire Mercenary</div>
                        <div class="player-detail">10 HP · uses any weapon · no special powers</div>
                    </div>
                </div>
            </div>`;

        window._buyChoiceCallback = (choice) => {
            hideGameModal();
            delete window._buyChoiceCallback;
            if (choice === 'mercenary') {
                sendAction('buy_mercenary', { card_id: cardID });
                resetActionState();
            } else {
                showBuyDeckConfirmModal(resource, cardID, cardsToReceive);
            }
        };

        showGameModal('Use Gold', 'Choose how to spend your gold', content, true);
        return;
    }

    showBuyDeckConfirmModal(resource, cardID, cardsToReceive);
}

function showBuyDeckConfirmModal(resource, cardID, cardsToReceive) {
    const resourceValue = resource?.value || 0;
    let cardsHtml = renderCardForModal(resource);
    cardsHtml += renderArrow();
    cardsHtml += renderCardBacks(cardsToReceive);

    showActionConfirmModal({
        title: 'Buy Cards',
        cardsHtml: cardsHtml,
        description: `Trade ${resourceValue} coins for ${cardsToReceive} card${cardsToReceive !== 1 ? 's' : ''} from the deck`,
        onConfirm: () => {
            sendAction('buy', { card_id: cardID });
            resetActionState();
        }
    });
}

// Construct phase handlers
function handleConstructPhaseHandClick(cardID, card) {
    if (card && card.can_be_used === false) {
        return;
    }

    // Fortress card: fortify a castle instead of constructing
    if (card && card.type === 'Fortress') {
        handleFortressPhaseHandClick(cardID, card);
        return;
    }

    // Select one resource card for construct
    clearSelections();
    gameState.selectedCards = [cardID];
    gameState.actionState.type = 'construct';
    highlightSelectedCard(cardID);

    // In 2v2 mode, offer choice between own castle and ally's castle
    const allies = (gameState.currentState?.opponents || []).filter(o => o.is_ally && !o.is_eliminated);
    const alliesWithCastle = allies.filter(a => a.castle?.constructed);
    const ownCastle = gameState.currentState?.current_player_castle;
    const canConstructOwn = ownCastle?.constructed || (card.type === 'resource' && card.value === 1) || (card.type === 'weapon' && card.value === 1);

    if (gameState.gameMode === '2v2' && alliesWithCastle.length > 0) {
        showConstructTargetModal(card, cardID, canConstructOwn, alliesWithCastle);
    } else {
        showConstructConfirmModal(card, cardID, '');
    }
}

function handleFortressPhaseHandClick(cardID, card) {
    clearSelections();
    highlightSelectedCard(cardID);

    const allies = (gameState.currentState?.opponents || []).filter(o => o.is_ally && !o.is_eliminated);
    const alliesWithCastle = allies.filter(a => a.castle?.constructed);

    if (gameState.gameMode === '2v2' && alliesWithCastle.length > 0) {
        // In 2v2, show target choice: own castle or ally's
        showFortressTargetModal(card, alliesWithCastle);
    } else {
        showFortressConfirmModal(card, '');
    }
}

function showFortressTargetModal(card, allies) {
    let content = '<div class="target-player-options">';
    content += `
        <div class="target-player-option" onclick="window._fortressTargetCallback('')">
            <span class="player-icon">🏰</span>
            <div class="player-info">
                <div class="player-name">Your Castle</div>
                <div class="player-detail">Protect your own castle</div>
            </div>
        </div>
    `;
    allies.forEach(ally => {
        const name = ally.player_name;
        content += `
            <div class="target-player-option" onclick="window._fortressTargetCallback('${name}')">
                <span class="player-icon">🤝</span>
                <div class="player-info">
                    <div class="player-name">${name}'s Castle</div>
                    <div class="player-detail">Protect ally's castle</div>
                </div>
            </div>
        `;
    });
    content += '</div>';

    window._fortressTargetCallback = (targetPlayer) => {
        hideGameModal();
        delete window._fortressTargetCallback;
        showFortressConfirmModal(card, targetPlayer);
    };

    showGameModal('Fortress', 'Choose which castle to protect', content, true);
}

function showFortressConfirmModal(card, targetPlayer) {
    const castleLabel = targetPlayer ? `${targetPlayer}'s castle` : 'your castle';
    const cardsHtml = renderCardForModal(card) + renderArrow() + renderCastleIcon();

    showActionConfirmModal({
        title: 'Fortify Castle',
        cardsHtml: cardsHtml,
        description: `Place a fortress wall on ${castleLabel} to block the next catapult attack`,
        onConfirm: () => {
            const payload = { card_id: card.id };
            if (targetPlayer) payload.target_player = targetPlayer;
            sendAction('fortress', payload);
            resetActionState();
        }
    });
}

function showResurrectionTargetModal(cardID, card, allies) {
    let content = '<div class="target-player-options">';
    content += `
        <div class="target-player-option" onclick="window._resurrectionTargetCallback('')">
            <span class="player-icon">⚔️</span>
            <div class="player-info">
                <div class="player-name">Your Field</div>
                <div class="player-detail">Resurrect to your own field</div>
            </div>
        </div>
    `;
    allies.forEach(ally => {
        const name = ally.player_name;
        content += `
            <div class="target-player-option" onclick="window._resurrectionTargetCallback('${name}')">
                <span class="player-icon">🤝</span>
                <div class="player-info">
                    <div class="player-name">${name}'s Field</div>
                    <div class="player-detail">Resurrect to ally's field</div>
                </div>
            </div>
        `;
    });
    content += '</div>';

    window._resurrectionTargetCallback = (targetPlayer) => {
        hideGameModal();
        delete window._resurrectionTargetCallback;
        showResurrectionConfirmModal(cardID, card, targetPlayer);
    };

    showGameModal('Resurrection', 'Choose which field to resurrect a warrior to', content, true);
}

function showResurrectionConfirmModal(cardID, card, targetPlayer) {
    const fieldLabel = targetPlayer ? `${targetPlayer}'s field` : 'your field';
    const cemeteryIconHtml = `
        <div class="castle-modal">
            <div class="cemetery-modal-icon">☠️</div>
            <span class="castle-modal-label">Cemetery</span>
        </div>
    `;
    const cardsHtml = cemeteryIconHtml + renderArrow() + renderCardForModal(card);

    showActionConfirmModal({
        title: 'Resurrect Warrior',
        cardsHtml: cardsHtml,
        description: `Bring a random fallen warrior back from the cemetery to ${fieldLabel}`,
        onConfirm: () => {
            const payload = { card_id: cardID };
            if (targetPlayer) payload.target_player = targetPlayer;
            sendAction('resurrection', payload);
            resetActionState();
        }
    });
}

function showConstructTargetModal(resource, cardID, canConstructOwn, allies) {
    let content = '<div class="target-player-options">';

    // Option: own castle (if possible)
    if (canConstructOwn) {
        content += `
            <div class="target-player-option" onclick="window._constructTargetCallback('')">
                <span class="player-icon">🏰</span>
                <div class="player-info">
                    <div class="player-name">Your Castle</div>
                    <div class="player-detail">${gameState.currentState?.current_player_castle?.constructed ? 'Value: ' + (gameState.currentState.current_player_castle.value || 0) + '/' + (gameState.currentState.current_player_castle.resources_to_win || 25) : 'Start construction'}</div>
                </div>
            </div>
        `;
    }

    // Option: each ally's castle
    allies.forEach(ally => {
        const name = ally.player_name;
        const castleValue = ally.castle?.value || 0;
        content += `
            <div class="target-player-option" onclick="window._constructTargetCallback('${name}')">
                <span class="player-icon">🤝</span>
                <div class="player-info">
                    <div class="player-name">${name}'s Castle</div>
                    <div class="player-detail">Value: ${castleValue}/${ally.castle?.resources_to_win || 25}</div>
                </div>
            </div>
        `;
    });
    content += '</div>';

    window._constructTargetCallback = (targetPlayer) => {
        hideGameModal();
        delete window._constructTargetCallback;
        showConstructConfirmModal(resource, cardID, targetPlayer);
    };

    showGameModal('Construct', 'Choose which castle to build', content, true);
}

function showConstructConfirmModal(resource, cardID, targetPlayer) {
    const resourceName = getCardName(resource);
    const resourceValue = resource?.value || 0;

    let castle, targetName;
    if (targetPlayer) {
        const ally = getOpponentByName(targetPlayer);
        castle = ally?.castle;
        targetName = targetPlayer;
    } else {
        castle = gameState.currentState?.current_player_castle;
        targetName = '';
    }

    const currentValue = castle?.value || 0;

    // Apply Harvest event modifier when castle is already constructed
    const gs = gameState.currentState;
    const harvestMod = (gs?.current_event === 'harvest' && castle?.constructed)
        ? (gs.current_event_resource_modifier || 0)
        : 0;
    const effectiveValue = harvestMod !== 0
        ? Math.max(1, resourceValue + harvestMod)
        : resourceValue;
    const newValue = currentValue + effectiveValue;

    let cardsHtml = renderCardForModal(resource);
    cardsHtml += renderArrow();
    cardsHtml += renderCastleIcon();

    const castleLabel = targetName ? `${targetName}'s castle` : 'your castle';
    let description;
    if (!castle?.constructed) {
        description = `${resourceName} (${resourceValue} value) will be added to ${castleLabel}`;
    } else if (harvestMod !== 0) {
        const sign = harvestMod > 0 ? '+' : '';
        description = `${resourceName} (${resourceValue} ${sign}${harvestMod} Harvest = ${effectiveValue} gold) → ${castleLabel} value: ${currentValue} → ${newValue}/${castle?.resources_to_win || 25}`;
    } else {
        description = `${resourceName} (${resourceValue} gold) → ${castleLabel} value: ${currentValue} → ${newValue}/${castle?.resources_to_win || 25}`;
    }

    const payload = { card_id: cardID };
    if (targetPlayer) {
        payload.target_player = targetPlayer;
    }

    showActionConfirmModal({
        title: castle?.constructed ? 'Add Gold to Castle' : 'Construct Castle',
        cardsHtml: cardsHtml,
        description: description,
        onConfirm: () => {
            sendAction('construct', payload);
            resetActionState();
        }
    });
}

function highlightValidUserWarriors(weapon) {
    const playerField = document.getElementById('player-field');
    playerField.classList.add('selecting-ally');
    playerField.classList.add('targeting-mode');

    playerField.querySelectorAll('.card').forEach(cardEl => {
        const fieldCard = findCardById(cardEl.dataset.cardId);
        const warriorType = (fieldCard?.sub_type || '').toLowerCase();
        if (SPECIAL_POWER_USER_TYPES.includes(warriorType)) {
            cardEl.classList.add('valid-target');
        }
    });
}

function highlightValidAttackerWarriors(weapon) {
    const weaponSubType = (weapon?.sub_type || '').toLowerCase();
    const validTypes = WEAPON_ATTACKER_TYPES[weaponSubType] || [];

    const playerField = document.getElementById('player-field');
    playerField.classList.add('selecting-ally');
    playerField.classList.add('targeting-mode');

    playerField.querySelectorAll('.card').forEach(cardEl => {
        const fieldCard = findCardById(cardEl.dataset.cardId);
        const warriorType = (fieldCard?.sub_type || '').toLowerCase();
        if (validTypes.length === 0 || validTypes.includes(warriorType)) {
            cardEl.classList.add('valid-target');
        }
    });
}

function enableSpecialPowerTargetSelection(userType) {
    // Enable target selection on the appropriate field based on warrior type
    const playerField = document.getElementById('player-field');

    if (userType === 'archer') {
        // Archer (Instant Kill) targets enemies only — dim own field + ally fields too
        document.querySelectorAll('.opponent-board:not(.ally) .opponent-field').forEach(f => {
            f.classList.add('selecting-target');
            f.classList.add('targeting-mode');
        });
        playerField.classList.add('targeting-mode');
        document.querySelectorAll('.opponent-board.ally .opponent-field').forEach(f => {
            f.classList.add('targeting-mode');
        });
    } else {
        // Mage (Heal) and Knight (Protect) target own field + ally fields; dim enemies
        playerField.classList.add('selecting-ally');
        playerField.classList.add('targeting-mode');
        document.querySelectorAll('.opponent-board.ally .opponent-field').forEach(f => {
            f.classList.add('selecting-ally');
            f.classList.add('targeting-mode');
        });
        document.querySelectorAll('.opponent-board:not(.ally) .opponent-field').forEach(f => {
            f.classList.add('targeting-mode');
        });
    }
}

function highlightUserWarrior(cardID) {
    // Remove any existing selected-user highlight
    document.querySelectorAll('.card.selected-user').forEach(c => c.classList.remove('selected-user'));
    const card = document.querySelector(`[data-card-id="${cardID}"]`);
    if (card) {
        card.classList.add('selected-user');
        card.classList.remove('selected'); // use gold glow, not the default blue
    }
}

function highlightSpecialPowerTargets(userType, weapon) {
    if (!weapon?.use_on) return;
    const useOn = weapon.use_on;

    if (userType === 'archer') {
        // Valid targets are enemy (non-ally) field warriors
        document.querySelectorAll('.opponent-board:not(.ally) .opponent-field .card').forEach(card => {
            if (useOn.includes(card.dataset.cardId)) {
                card.classList.add('valid-target');
            }
        });
    } else {
        // Valid targets are own field + ally field warriors
        document.querySelectorAll('#player-field .card').forEach(card => {
            if (useOn.includes(card.dataset.cardId)) {
                card.classList.add('valid-target');
            }
        });
        document.querySelectorAll('.opponent-board.ally .opponent-field .card').forEach(card => {
            if (useOn.includes(card.dataset.cardId)) {
                card.classList.add('valid-target');
            }
        });
    }
}

function clearSpecialPowerTargetSelection() {
    // Remove valid-target from all field cards
    document.querySelectorAll('.opponent-field .card.valid-target, #player-field .card.valid-target').forEach(c => {
        c.classList.remove('valid-target');
    });
    // Remove targeting-mode and target selection classes from opponent fields
    document.querySelectorAll('.opponent-field').forEach(f => {
        f.classList.remove('selecting-target');
        f.classList.remove('selecting-ally');
        f.classList.remove('targeting-mode');
    });
    // Remove targeting-mode from player field (but keep selecting-ally for step 1)
    document.getElementById('player-field')?.classList.remove('targeting-mode');
}

function highlightValidTargets(weapon) {
    const dmgMult = weapon?.dmg_mult || {};

    // Highlight valid targets on all opponent fields
    document.querySelectorAll('.opponent-field .card').forEach(card => {
        const cardId = card.dataset.cardId;
        if (weapon && weapon.use_on && weapon.use_on.includes(cardId)) {
            card.classList.add('valid-target');
            // Show multiplier badge if > 1
            if (dmgMult[cardId] && dmgMult[cardId] > 1) {
                addMultiplierBadge(card, dmgMult[cardId]);
            }
        }
    });

    // Also highlight valid targets on player field (for Mage heal, Knight protect)
    const playerField = document.getElementById('player-field');
    playerField.querySelectorAll('.card').forEach(card => {
        const cardId = card.dataset.cardId;
        if (weapon && weapon.use_on && weapon.use_on.includes(cardId)) {
            card.classList.add('valid-target');
        }
    });
}

function addMultiplierBadge(cardElement, multiplier) {
    // Remove existing badge if any
    const existing = cardElement.querySelector('.dmg-multiplier-badge');
    if (existing) existing.remove();

    const badge = document.createElement('div');
    badge.className = 'dmg-multiplier-badge';
    badge.textContent = `x${multiplier}`;
    cardElement.appendChild(badge);
}

// Build attack summary for normal weapon attacks
function buildAttackSummary(weapon, target) {
    const weaponName = getCardName(weapon);
    const weaponDmg = weapon?.value || 0;
    const targetName = getCardName(target);
    const targetHp = target?.value || 0;
    const targetId = target?.id;
    const multiplier = weapon?.dmg_mult?.[targetId] || 1;
    const effectiveDmg = weaponDmg * multiplier;

    if (multiplier > 1) {
        return `⚔️ ${weaponName} (${weaponDmg} x${multiplier} = ${effectiveDmg} DMG) → ${targetName} (${targetHp} HP)`;
    }
    return `⚔️ ${weaponName} (${weaponDmg} DMG) → ${targetName} (${targetHp} HP)`;
}

// Build summary for special power attacks
function buildSpecialPowerSummary(specialPower, user, target) {
    const userName = getCardName(user);
    const targetName = getCardName(target);
    const targetHp = target?.value || 0;

    // Get the effect based on warrior type
    const userType = (user?.sub_type || '').toLowerCase();

    switch (userType) {
        case 'archer':
            return `🎯 ${userName} → INSTANT KILL → ${targetName}`;
        case 'knight':
            return `🛡️ ${userName} → PROTECT → ${targetName} (${targetHp} HP)`;
        case 'mage':
            return `💚 ${userName} → HEAL → ${targetName} (${targetHp} HP)`;
        default:
            return `✨ ${userName} → ${getCardName(specialPower)} → ${targetName}`;
    }
}

// Helper: get non-eliminated, non-ally opponents
function getEnemyOpponents() {
    const opponents = gameState.currentState?.opponents || [];
    return opponents.filter(o => !o.is_eliminated && !o.is_ally);
}

// Helper: get opponent data by name
function getOpponentByName(name) {
    const opponents = gameState.currentState?.opponents || [];
    return opponents.find(o => o.player_name === name) || null;
}

// Target Player Selection Modal
function showTargetPlayerModal(title, opponents, callback, detailFn) {
    // Skip modal when there's only one option
    if (opponents.length === 1) {
        callback(opponents[0].player_name);
        return;
    }

    const defaultDetail = (opp) => {
        const castle = opp.castle || {};
        return `Castle: ${castle.value || 0}/${castle.resources_to_win || 25} gold, ${castle.resource_cards || 0} resource cards`;
    };
    const getDetail = detailFn || defaultDetail;

    let content = '<div class="target-player-options">';
    opponents.forEach(opp => {
        const name = opp.player_name;
        const detail = getDetail(opp);
        content += `
            <div class="target-player-option" onclick="window._targetPlayerCallback('${name}')">
                <span class="player-icon">⚔</span>
                <div class="player-info">
                    <div class="player-name">${name}</div>
                    <div class="player-detail">${detail}</div>
                </div>
            </div>
        `;
    });
    content += '</div>';

    window._targetPlayerCallback = (playerName) => {
        hideGameModal();
        delete window._targetPlayerCallback;
        callback(playerName);
    };

    showGameModal(title, 'Choose a target player', content, true);
}

// Move Warrior Target Modal (2v2: own field vs ally's field)
function showMoveWarriorTargetModal(cardID, warriorCard, cardHtml, allies) {
    let content = '<div class="target-player-options">';
    // Option: own field
    content += `
        <div class="target-player-option" onclick="window._moveWarriorTargetCallback('')">
            <span class="player-icon">🛡</span>
            <div class="player-info">
                <div class="player-name">Your Field</div>
                <div class="player-detail">Move to your own field</div>
            </div>
        </div>
    `;
    // Option: each ally's field
    allies.forEach(ally => {
        const name = ally.player_name;
        const detail = `${(ally.field || []).length} warriors on field`;
        content += `
            <div class="target-player-option" onclick="window._moveWarriorTargetCallback('${name}')">
                <span class="player-icon">🤝</span>
                <div class="player-info">
                    <div class="player-name">${name}'s Field</div>
                    <div class="player-detail">${detail}</div>
                </div>
            </div>
        `;
    });
    content += '</div>';

    window._moveWarriorTargetCallback = (targetPlayer) => {
        hideGameModal();
        delete window._moveWarriorTargetCallback;

        const payload = { warrior_id: cardID };
        if (targetPlayer) {
            payload.target_player = targetPlayer;
        }

        showActionConfirmModal({
            title: 'Move Warrior',
            cardsHtml: cardHtml,
            description: targetPlayer
                ? `${getCardName(warriorCard)} will move to ${targetPlayer}'s field`
                : `${getCardName(warriorCard)} will move to your field`,
            onConfirm: () => {
                sendAction('move_warrior', payload);
                resetActionState();
            }
        });
    };

    showGameModal('Move Warrior', 'Choose where to deploy', content, true);
}

// Steal Modal
function showStealModal() {
    const targetName = gameState.actionState.targetPlayer;
    const opponent = getOpponentByName(targetName);
    const handCount = opponent?.cards_in_hand || 0;

    // Skip modal — pick a random card position
    if (handCount > 0) {
        selectStealPosition(Math.ceil(Math.random() * handCount));
        return;
    }

    let content = '';
    for (let i = 1; i <= handCount; i++) {
        content += `
            <div class="card-face-down" data-position="${i}" onclick="selectStealPosition(${i})">
                <span class="card-position">#${i}</span>
            </div>
        `;
    }

    showGameModal(`Steal from ${targetName}`, "Choose one of their cards", content, true);
}

function selectStealPosition(position) {
    gameState.pendingModalAction = 'steal';
    sendAction('steal', { card_id: gameState.actionState.weaponId, target_player: gameState.actionState.targetPlayer, card_position: position });
    hideGameModal();
}

// Treason Modal — shows eligible warriors (≤5 HP) from the target opponent's field
function showTreasonModal() {
    const targetName = gameState.actionState.targetPlayer;
    const opponent = getOpponentByName(targetName);
    const field = (opponent?.field || []).filter(w => w.value <= 5);

    if (field.length === 0) {
        updateActionPrompt(`${targetName} has no warriors with 5 HP or less!`);
        resetActionState();
        return;
    }

    // Skip modal when there's only one eligible warrior
    if (field.length === 1) {
        selectTreasonWarrior(field[0].id);
        return;
    }

    let content = '<div class="treason-warrior-grid">';
    field.forEach(warrior => {
        const subType = warrior.sub_type || warrior.type || 'Warrior';
        const imgKey = subType.toLowerCase();
        const imgSrc = `/static/img/cards/${imgKey}.webp`;
        content += `
            <div class="treason-warrior-option" onclick="selectTreasonWarrior('${warrior.id}')">
                <div class="card ${imgKey} has-image" style="border-color:${warrior.color};">
                    <div class="card-image"><img src="${imgSrc}" alt="${subType}" draggable="false" onerror="this.parentNode.style.display='none'"></div>
                    <div class="card-info">
                        <span class="card-name">${subType}</span>
                        <span class="card-stat-badge warrior">HP ${warrior.value}</span>
                    </div>
                </div>
            </div>
        `;
    });
    content += '</div>';

    showGameModal(`Treason — ${targetName}`, 'Choose a weakened warrior to convince (≤5 HP)', content, true);
}

function selectTreasonWarrior(warriorID) {
    sendAction('treason', { card_id: gameState.actionState.weaponId, target_player: gameState.actionState.targetPlayer, warrior_id: warriorID });
    hideGameModal();
}

// Spy Options Modal
function showSpyOptionsModal() {
    const enemies = getEnemyOpponents();
    const hasMultipleEnemies = enemies.length > 1;

    const content = `
        <div class="spy-option" onclick="selectSpyOption(1)">
            <div class="spy-option-title">Reveal Deck</div>
            <div class="spy-option-desc">See the top 5 cards from the deck</div>
        </div>
        <div class="spy-option" onclick="selectSpyOption(2)">
            <div class="spy-option-title">Reveal Enemy Hand</div>
            <div class="spy-option-desc">See all cards in your opponent's hand</div>
        </div>
    `;

    showGameModal('Choose Spy Action', 'Select what you want to spy on', content, true);
}

function selectSpyOption(option) {
    const cardId = gameState.actionState.weaponId;

    if (option === 1) {
        // Reveal deck - no target player needed
        gameState.pendingModalAction = 'spy_deck';
        const enemies = getEnemyOpponents();
        // Backend requires a target_player even for deck spy; use first enemy
        sendAction('spy', { card_id: cardId, target_player: enemies[0]?.player_name || '', option: option });
        hideGameModal();
    } else {
        // Reveal hand - need to select target player
        const enemies = getEnemyOpponents();
        if (enemies.length === 1) {
            gameState.pendingModalAction = 'spy_hand';
            sendAction('spy', { card_id: cardId, target_player: enemies[0].player_name, option: option });
            hideGameModal();
        } else {
            hideGameModal();
            showTargetPlayerModal('Whose hand do you want to reveal?', enemies, (playerName) => {
                gameState.pendingModalAction = 'spy_hand';
                sendAction('spy', { card_id: cardId, target_player: playerName, option: option });
            }, (opp) => `${opp.cards_in_hand} cards in hand`);
        }
    }
}

// Catapult Modal
function showCatapultModal() {
    const targetName = gameState.actionState.targetPlayer;
    const opponent = getOpponentByName(targetName);
    const isProtected = opponent?.castle?.is_protected || false;
    const resourceCount = opponent?.castle?.resource_cards || 0;

    if (isProtected) {
        confirmCatapultFortress(targetName);
        return;
    }

    if (resourceCount === 0) {
        updateActionPrompt('Castle has no resources to attack!');
        resetActionState();
        return;
    }

    // Pick a random resource position — castle resources are face-down to the attacker
    selectCatapultPosition(Math.ceil(Math.random() * resourceCount));
}

function selectCatapultPosition(position) {
    sendAction('catapult', { card_id: gameState.actionState.weaponId, target_player: gameState.actionState.targetPlayer, card_position: position });
    hideGameModal();
}

function confirmCatapultFortress(targetName) {
    sendAction('catapult', { card_id: gameState.actionState.weaponId, target_player: targetName, card_position: 1 });
    hideGameModal();
}

// Enter trade mode with a card pre-selected (called from card click or trade icon)
function startTradeFromCard(cardID) {
    gameState.currentAction = 'trade';
    gameState.actionState.type = 'trade';
    gameState.selectedCards = [cardID];
    updateActionPrompt('Selected 1/3 cards for trade');
    showConfirmButtons();
    renderGameBoard(gameState.currentState);
    document.querySelector(`[data-card-id="${cardID}"]`)?.classList.add('selected');
}

// Forge Mode — enter weapon selection for forging (called from badge click)
function startForgeFromCard(cardID, card) {
    resetActionState();
    gameState.currentAction = 'forge';
    gameState.selectedCards = [cardID];
    showConfirmButtons();
    renderGameBoard(gameState.currentState);
    document.querySelector(`[data-card-id="${cardID}"]`)?.classList.add('selected');
    updateActionPrompt(`Select another ${card.sub_type} to forge with (1/2)`);
}

// Handle card click during forge mode
function handleForgeCardClick(cardID, card) {
    const forgeableTypes = ['Sword', 'Arrow', 'Poison'];
    if (!forgeableTypes.includes(card.sub_type)) return;

    const selected = gameState.selectedCards;

    // Deselect if already selected
    if (selected.includes(cardID)) {
        selected.splice(selected.indexOf(cardID), 1);
        const el = document.querySelector(`[data-card-id="${cardID}"]`);
        if (el) el.classList.remove('selected');
        updateActionPrompt(`Select 2 weapons of the same type to forge (${selected.length}/2)`);
        return;
    }

    // Validate same sub_type when 1 already selected
    if (selected.length === 1) {
        const firstCard = findCardById(selected[0]);
        if (!firstCard || firstCard.sub_type !== card.sub_type) {
            updateActionPrompt('Both weapons must be the same type. Try again (1/2)');
            return;
        }
    }

    selected.push(cardID);
    const el = document.querySelector(`[data-card-id="${cardID}"]`);
    if (el) el.classList.add('selected');

    if (selected.length === 2) {
        showForgeConfirmModal(selected[0], selected[1]);
    } else {
        updateActionPrompt('Select 2 weapons of the same type to forge (1/2)');
    }
}

// Forge confirmation modal
function showForgeConfirmModal(id1, id2) {
    const card1 = findCardById(id1);
    const card2 = findCardById(id2);
    if (!card1 || !card2) return;

    const dmg1 = card1.value || 0;
    const dmg2 = card2.value || 0;
    const total = dmg1 + dmg2;

    let cardsHtml = renderCardForModal(card1);
    cardsHtml += '<div class="modal-plus">+</div>';
    cardsHtml += renderCardForModal(card2);
    cardsHtml += renderArrow();
    cardsHtml += renderCardForModal({...card1, value: total});

    showActionConfirmModal({
        title: 'Forge Weapons',
        cardsHtml: cardsHtml,
        description: `Combine into a ${card1.sub_type || card1.type} ${total}`,
        onConfirm: () => {
            gameState.pendingAction = 'forge';
            sendAction('forge', { card_id_1: id1, card_id_2: id2 });
            resetActionState();
        }
    });
}

// Ambush Target Selection Modal (2v2 — own field vs ally field)
function showAmbushTargetModal(card, cardID, allies, canPlaceOwn) {
    let content = '<div class="target-player-options">';

    if (canPlaceOwn === false) {
        // own field already has ambush — don't show own option
    } else {
        content += `
            <div class="target-player-option" onclick="window._ambushTargetCallback('')">
                <span class="player-icon">⚠</span>
                <div class="player-info">
                    <div class="player-name">Your Field</div>
                    <div class="player-detail">Place ambush on your own field</div>
                </div>
            </div>
        `;
    }

    allies.forEach(ally => {
        const name = ally.player_name;
        content += `
            <div class="target-player-option" onclick="window._ambushTargetCallback('${name}')">
                <span class="player-icon">🤝</span>
                <div class="player-info">
                    <div class="player-name">${name}'s Field</div>
                    <div class="player-detail">Place ambush on ally's field</div>
                </div>
            </div>
        `;
    });
    content += '</div>';

    window._ambushTargetCallback = (targetPlayer) => {
        hideGameModal();
        delete window._ambushTargetCallback;
        showAmbushPlaceConfirmModal(card, cardID, targetPlayer);
    };

    showGameModal('Place Ambush', 'Choose which field to protect', content, true);
}

// Ambush Place Confirmation Modal
function showAmbushPlaceConfirmModal(card, cardID, targetPlayer) {
    const effectRows = [
        { label: 'Reflect Damage', chance: '23%', desc: 'Attacker\'s weapon damage is reflected back at their own warrior.' },
        { label: 'Attack Cancelled', chance: '23%', desc: 'The attack is cancelled and the weapon is discarded.' },
        { label: 'Weapon Stolen', chance: '23%', desc: 'The attacking weapon is intercepted and added to your hand.' },
        { label: 'Drain Life', chance: '23%', desc: 'Your warrior takes no damage and heals HP equal to the weapon\'s value.' },
        { label: 'Instant Kill', chance: '8%', desc: 'One of the attacker\'s warriors is instantly killed.' },
    ];

    const rowsHtml = effectRows.map(e => `
        <tr>
            <td class="ambush-confirm-effect-label">${e.label}</td>
            <td class="ambush-confirm-chance">${e.chance}</td>
            <td class="ambush-confirm-desc">${e.desc}</td>
        </tr>`).join('');

    const description = `
        <div class="ambush-confirm-info">
            <p class="ambush-confirm-note">Triggers when an enemy uses a weapon attack against a warrior in your field. The effect is chosen randomly:</p>
            <table class="ambush-confirm-table">
                <tbody>${rowsHtml}</tbody>
            </table>
        </div>`;

    showActionConfirmModal({
        title: 'Place Ambush',
        cardsHtml: renderCardForModal(card),
        description,
        onConfirm: () => {
            sendMessage('place_ambush', { card_id: cardID, target_player: targetPlayer || '' });
            resetActionState();
        },
    });
}
