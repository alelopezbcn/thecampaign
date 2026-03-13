// Game Modal Functions
function showGameModal(title, subtitle, content, showCloseOnly = true) {
    const modal = document.getElementById('game-modal');
    document.getElementById('modal-title').textContent = title;
    document.getElementById('modal-subtitle').textContent = subtitle;
    document.getElementById('modal-cards-container').innerHTML = content;

    // Show/hide close button based on context
    const closeBtn = document.getElementById('modal-close-btn');
    closeBtn.style.display = showCloseOnly ? 'block' : 'none';

    modal.classList.remove('hidden');
}

function hideGameModal() {
    const modal = document.getElementById('game-modal');
    modal.classList.add('hidden');
    if (cardsModalTimer) { clearTimeout(cardsModalTimer); cardsModalTimer = null; }
    resetActionState();
    updateActionPrompt('');
    if (gameState.currentState) {
        renderGameBoard(gameState.currentState);
    }

    // Play deferred animations that were waiting for modal close
    if (pendingAnimationsCallback) {
        const callback = pendingAnimationsCallback;
        pendingAnimationsCallback = null;
        callback();
    }
}

function showActionConfirmModal(config) {
    const modal = document.getElementById('action-confirm-modal');
    document.getElementById('action-confirm-title').textContent = config.title || 'Confirm Action';
    document.getElementById('action-confirm-cards').innerHTML = config.cardsHtml || '';
    document.getElementById('action-confirm-description').innerHTML = config.description || '';

    actionConfirmCallback = config.onConfirm || null;

    // Hide the bottom confirm buttons and clear the action prompt
    hideConfirmButtons();
    updateActionPrompt('');

    modal.classList.remove('hidden');
}

function hideActionConfirmModal() {
    const modal = document.getElementById('action-confirm-modal');
    modal.classList.add('hidden');
    actionConfirmCallback = null;
}

function onActionConfirmYes() {
    if (actionConfirmCallback) {
        actionConfirmCallback();
    }
    hideActionConfirmModal();
}

function onActionConfirmNo() {
    hideActionConfirmModal();
    resetActionState();
    updateActionPrompt('');
    if (gameState.currentState) {
        renderGameBoard(gameState.currentState);
    }
}

// Render a card for the action confirm modal
function renderCardForModal(card, options = {}) {
    if (!card) return '';

    const cardType = getCardType(card);
    const cardName = getCardName(card);
    const bgColor = card.color ? hexToRgba(card.color, 0.3) : '';
    const borderColor = card.color || '';
    const imageUrl = getCardImageUrl(card);

    let wrapperClass = 'card-wrapper';
    let badgeHtml = '';

    if (options.showDoubleDamage) {
        badgeHtml = '<div class="double-damage-badge">x2 DMG</div>';
    }
    if (options.showShield) {
        badgeHtml += `<div class="shield-badge">🛡️ ${options.shieldHp} HP</div>`;
    }
    if (options.showHealResult) {
        badgeHtml += `<div class="heal-result-badge">✨ ${options.healedHp} HP</div>`;
    }
    if (options.killBonus > 0) {
        badgeHtml += `<div class="kill-bonus-badge">💀 +${options.killBonus} DMG</div>`;
    }

    let cardHtml;
    if (imageUrl) {
        cardHtml = `
            <div class="card ${cardType} has-image" style="${borderColor ? `border-color: ${borderColor};` : ''}">
                <div class="card-image">
                    <img src="${imageUrl}" alt="${cardName}" draggable="false">
                </div>
                <div class="card-info">
                    <span class="card-name">${cardName}</span>
                    ${getCardStatBadge(card, cardType)}
                </div>
            </div>
        `;
    } else {
        cardHtml = `
            <div class="card ${cardType}" style="${bgColor ? `background: ${bgColor};` : ''} ${borderColor ? `border-color: ${borderColor};` : ''}">
                <div class="card-header">
                    <span class="card-type ${cardType}">${card.type || cardType}</span>
                </div>
                <div class="card-content">
                    <div class="card-name">${cardName}</div>
                    ${getCardStats(card, cardType)}
                </div>
            </div>
        `;
    }

    if (badgeHtml) {
        return `<div class="${wrapperClass}">${badgeHtml}${cardHtml}</div>`;
    }
    return cardHtml;
}

// Render card backs for the modal
function renderCardBacks(count) {
    if (count <= 0) return '';

    let html = '<div class="cards-group">';
    for (let i = 0; i < count; i++) {
        html += `
            <div class="card-back-modal">
                <div class="card-back-modal-inner">
                    <span class="card-back-modal-emblem">?</span>
                </div>
            </div>
        `;
    }
    html += '</div>';
    return html;
}

// Render castle for construct action (same style as board castle)
function renderCastleIcon() {
    return `
        <div class="castle-modal">
            <div class="castle-icon"></div>
            <span class="castle-modal-label">Castle</span>
        </div>
    `;
}

// Arrow element for modal
function renderArrow() {
    return '<span class="action-confirm-arrow">→</span>';
}

// Turn Transition Toast
function showTurnTransitionModal(playerName, duration = 3000, overrideText = null) {
    if (!gameState.currentState) return;
    const container = document.getElementById('error-toast-container');
    if (!container) return;

    const text = overrideText || (playerName === gameState.playerName ? 'Your Turn!' : `${playerName}'s Turn`);
    const isYou = text.startsWith('Your Turn');

    const toast = document.createElement('div');
    toast.className = 'error-toast turn-transition-toast' + (isYou ? ' turn-transition-toast-you' : '');
    toast.innerHTML =
        '<span class="error-toast-icon">&#9876;</span>' +
        '<div class="error-toast-content">' +
            '<div class="error-toast-title">' + text + '</div>' +
        '</div>' +
        '<button class="error-toast-close" onclick="this.closest(\'.error-toast\').remove()">&#x2715;</button>';
    container.appendChild(toast);

    setTimeout(() => {
        toast.classList.add('hiding');
        setTimeout(() => toast.remove(), 300);
    }, duration);
}

function hideTurnTransitionModal() {}

function showStolenCardModal(card, action = 'stolen') {
    const modal = document.getElementById('stolen-card-modal');
    const container = document.getElementById('stolen-card-container');
    const text = document.getElementById('stolen-card-text');

    if (!modal || !container || !text) return;

    // Render the stolen card
    const cardName = card.sub_type || card.type;
    container.innerHTML = renderCardForModal(card);

    const title = modal.querySelector('.stolen-card-title');
    const icon = modal.querySelector('.stolen-card-icon');
    if (action === 'sabotaged') {
        if (title) title.textContent = 'You were sabotaged!';
        if (icon) icon.textContent = '💣';
        text.textContent = `${cardName} was sabotaged from you!`;
    } else {
        if (title) title.textContent = 'You were robbed!';
        if (icon) icon.innerHTML = '&#128163;';
        text.textContent = `${cardName} was stolen from you!`;
    }

    modal.classList.remove('hidden');

    if (stolenCardTimer) clearTimeout(stolenCardTimer);
    stolenCardTimer = setTimeout(() => {
        hideStolenCardModal();
    }, 5000);
}

function hideStolenCardModal() {
    const modal = document.getElementById('stolen-card-modal');
    modal.classList.add('hidden');
    if (stolenCardTimer) {
        clearTimeout(stolenCardTimer);
        stolenCardTimer = null;
    }
}

function showSpyNotificationModal(message) {
    const modal = document.getElementById('spy-notification-modal');
    const textEl = document.getElementById('spy-notification-text');

    if (!modal || !textEl) return;

    textEl.textContent = message;
    modal.classList.remove('hidden');

    if (spyNotificationTimer) clearTimeout(spyNotificationTimer);
    spyNotificationTimer = setTimeout(() => {
        hideSpyNotificationModal();
    }, 5000);
}

function hideSpyNotificationModal() {
    const modal = document.getElementById('spy-notification-modal');
    if (modal) modal.classList.add('hidden');
    if (spyNotificationTimer) {
        clearTimeout(spyNotificationTimer);
        spyNotificationTimer = null;
    }
}

let catapultNotificationTimer = null;

function showCatapultNotificationModal(notification) {
    const isTarget = notification.target_player === gameState.playerName;

    if (isTarget) {
        const modal = document.getElementById('catapult-notification-modal');
        const text = document.getElementById('catapult-notification-text');
        if (!modal || !text) return;

        if (notification.blocked) {
            text.textContent = `${notification.attacker_name} launched a catapult at your castle, but your Fortress wall absorbed the hit and was destroyed!`;
        } else {
            text.textContent = `${notification.attacker_name} launched a catapult and removed ${notification.gold_stolen} gold from your castle!`;
        }

        modal.classList.remove('hidden');

        if (catapultNotificationTimer) clearTimeout(catapultNotificationTimer);
        catapultNotificationTimer = setTimeout(() => hideCatapultNotificationModal(), 6000);
        return;
    }

    // Spectators and the attacker: toast
    const container = document.getElementById('error-toast-container');
    if (!container) return;

    let msg;
    if (notification.blocked) {
        msg = `${notification.attacker_name} catapulted ${notification.target_player}'s castle — Fortress wall absorbed the hit!`;
    } else {
        msg = `${notification.attacker_name} catapulted ${notification.target_player}'s castle and removed ${notification.gold_stolen} gold!`;
    }

    const toast = document.createElement('div');
    toast.className = 'error-toast catapult-toast';
    toast.innerHTML =
        '<span class="error-toast-icon">&#128163;</span>' +
        '<div class="error-toast-content">' +
            '<div class="error-toast-title">Catapult Attack!</div>' +
            '<div class="error-toast-message">' + msg + '</div>' +
        '</div>' +
        '<button class="error-toast-close" onclick="this.closest(\'.error-toast\').remove()">&#x2715;</button>';
    container.appendChild(toast);

    setTimeout(() => {
        toast.classList.add('hiding');
        setTimeout(() => toast.remove(), 300);
    }, 7000);
}

function hideCatapultNotificationModal() {
    const modal = document.getElementById('catapult-notification-modal');
    if (modal) modal.classList.add('hidden');
    if (catapultNotificationTimer) {
        clearTimeout(catapultNotificationTimer);
        catapultNotificationTimer = null;
    }
}

function showTreasonNotificationModal(notification) {
    const modal = document.getElementById('treason-notification-modal');
    const container = document.getElementById('treason-warrior-container');
    const text = document.getElementById('treason-notification-text');
    if (!modal || !container || !text) return;

    const warrior = notification.warrior_card;
    const stolenBy = notification.stolen_by;
    container.innerHTML = renderCardForModal(warrior);
    const warriorName = warrior.sub_type || warrior.type || 'Warrior';
    text.textContent = `${warriorName} (${warrior.value} HP) moved to ${stolenBy} rank!`;

    modal.classList.remove('hidden');

    if (treasonNotificationTimer) clearTimeout(treasonNotificationTimer);
    treasonNotificationTimer = setTimeout(() => hideTreasonNotificationModal(), 5000);
}

function hideTreasonNotificationModal() {
    const modal = document.getElementById('treason-notification-modal');
    if (modal) modal.classList.add('hidden');
    if (treasonNotificationTimer) {
        clearTimeout(treasonNotificationTimer);
        treasonNotificationTimer = null;
    }
}

function showChampionsBountyModal(notification) {
    const container = document.getElementById('error-toast-container');
    if (!container) return;

    const isYou = notification.earned_by === gameState.playerName;
    const cardWord = notification.cards === 1 ? 'card' : 'cards';
    const msg = isYou
        ? `You slew the champion's warrior and drew ${notification.cards} ${cardWord}!`
        : `${notification.earned_by} slew the champion's warrior and drew ${notification.cards} ${cardWord}!`;

    const toast = document.createElement('div');
    toast.className = 'error-toast champions-bounty-toast';
    toast.innerHTML =
        '<span class="error-toast-icon">&#127942;</span>' +
        '<div class="error-toast-content">' +
            '<div class="error-toast-title">Champion\'s Bounty!</div>' +
            '<div class="error-toast-message">' + msg + '</div>' +
        '</div>' +
        '<button class="error-toast-close" onclick="this.closest(\'.error-toast\').remove()">&#x2715;</button>';
    container.appendChild(toast);

    setTimeout(() => {
        toast.classList.add('hiding');
        setTimeout(() => toast.remove(), 300);
    }, 7000);
}

function hideChampionsBountyModal() {}

let resurrectionNotificationTimer = null;

function showResurrectionModal(notification) {
    const isYou = notification.player_name === gameState.playerName;
    const warriorName = notification.warrior_card?.sub_type || 'a warrior';
    const target = notification.target_player;
    const targetIsYou = target === gameState.playerName;

    if (isYou) {
        // Active player: show card reveal popup
        const modal = document.getElementById('resurrection-notification-modal');
        const container = document.getElementById('resurrection-warrior-container');
        const text = document.getElementById('resurrection-notification-text');
        if (!modal || !container || !text) return;

        container.innerHTML = renderCardForModal(notification.warrior_card);
        text.textContent = targetIsYou || !target || target === notification.player_name
            ? `${warriorName} returned from the cemetery to your field!`
            : `${warriorName} returned from the cemetery to ${target}'s field!`;

        modal.classList.remove('hidden');

        if (resurrectionNotificationTimer) clearTimeout(resurrectionNotificationTimer);
        resurrectionNotificationTimer = setTimeout(() => hideResurrectionModal(), 6000);
        return;
    }

    // Other players: toast notification
    const toastContainer = document.getElementById('error-toast-container');
    if (!toastContainer) return;

    let msg;
    if (targetIsYou) {
        msg = `${notification.player_name} resurrected ${warriorName} to your field!`;
    } else if (target && target !== notification.player_name) {
        msg = `${notification.player_name} resurrected ${warriorName} to ${target}'s field!`;
    } else {
        msg = `${notification.player_name} resurrected ${warriorName} from the cemetery!`;
    }

    const toast = document.createElement('div');
    toast.className = 'error-toast resurrection-toast';
    toast.innerHTML =
        '<span class="error-toast-icon">&#129503;</span>' +
        '<div class="error-toast-content">' +
            '<div class="error-toast-title">Resurrection!</div>' +
            '<div class="error-toast-message">' + msg + '</div>' +
        '</div>' +
        '<button class="error-toast-close" onclick="this.closest(\'.error-toast\').remove()">&#x2715;</button>';
    toastContainer.appendChild(toast);

    setTimeout(() => {
        toast.classList.add('hiding');
        setTimeout(() => toast.remove(), 300);
    }, 7000);
}

function hideResurrectionModal() {
    const modal = document.getElementById('resurrection-notification-modal');
    if (modal) modal.classList.add('hidden');
    if (resurrectionNotificationTimer) {
        clearTimeout(resurrectionNotificationTimer);
        resurrectionNotificationTimer = null;
    }
}

let ambushNotificationTimer = null;

// Returns HTML detail rows for the ambush modal, per effect type.
function buildAmbushDetailRows(a) {
    const effect = a.effect_display;
    const weaponStr = a.weapon_type ? `${a.weapon_type} (${a.damage_amount} dmg)` : '';

    function hpChange(before, after, died) {
        if (died) return `<span class="ambush-detail-value died">${before} HP &#8594; DEAD</span>`;
        if (after > before) return `<span class="ambush-detail-value healed">${before} HP &#8594; ${after} HP (+${after - before})</span>`;
        if (after < before) return `<span class="ambush-detail-value">${before} HP &#8594; ${after} HP</span>`;
        return `<span class="ambush-detail-value">${before} HP (unchanged)</span>`;
    }

    function row(label, valueHtml) {
        return `<div class="ambush-detail-row"><span class="ambush-detail-label">${label}</span>${valueHtml}</div>`;
    }

    let rows = '';

    if (effect === 'Reflect Damage') {
        if (weaponStr)               rows += row('Weapon', `<span class="ambush-detail-value">${weaponStr}</span>`);
        if (a.attacker_warrior_type) rows += row(`${a.attacker_warrior_type} (attacker)`, hpChange(a.attacker_hp_before, a.attacker_hp_after, a.attacker_died));
        if (a.target_warrior_type)   rows += row(`${a.target_warrior_type} (target)`, `<span class="ambush-detail-value">${a.target_hp_before} HP (unchanged)</span>`);

    } else if (effect === 'Attack Cancelled') {
        if (weaponStr)               rows += row('Weapon discarded', `<span class="ambush-detail-value">${weaponStr}</span>`);
        if (a.attacker_warrior_type) rows += row('Attacker', `<span class="ambush-detail-value">${a.attacker_warrior_type} (${a.attacker_hp_before} HP)</span>`);
        if (a.target_warrior_type)   rows += row('Target',   `<span class="ambush-detail-value">${a.target_warrior_type} (${a.target_hp_before} HP)</span>`);

    } else if (effect === 'Weapon Stolen') {
        if (weaponStr)               rows += row('Weapon stolen', `<span class="ambush-detail-value">${weaponStr}</span>`);
        rows += row('Now in hand of', `<span class="ambush-detail-value">${a.defender_name}</span>`);
        if (a.attacker_warrior_type) rows += row('Attacker', `<span class="ambush-detail-value">${a.attacker_warrior_type} (${a.attacker_hp_before} HP)</span>`);

    } else if (effect === 'Drain Life') {
        if (weaponStr)             rows += row('Attack absorbed', `<span class="ambush-detail-value">${weaponStr}</span>`);
        if (a.target_warrior_type) rows += row(`${a.target_warrior_type} (target)`, hpChange(a.target_hp_before, a.target_hp_after, false));

    } else if (effect === 'Instant Kill') {
        if (a.attacker_warrior_type) rows += row('Warrior killed', `<span class="ambush-detail-value died">${a.attacker_warrior_type} &#8594; DEAD</span>`);
        if (weaponStr)               rows += row('Weapon discarded', `<span class="ambush-detail-value">${weaponStr}</span>`);
    }

    return rows ? `<div class="ambush-notification-details">${rows}</div>` : '';
}

// Returns a one-line summary for spectator toasts.
function buildAmbushSpectatorMessage(a) {
    const effect = a.effect_display;
    const attWarrior = a.attacker_warrior_type || 'warrior';
    const defWarrior = a.target_warrior_type   || 'warrior';
    const dmg = a.damage_amount;

    if (effect === 'Reflect Damage')   return a.attacker_died
        ? `${a.attacker_name}'s ${attWarrior} was killed by the reflection`
        : `${a.attacker_name}'s ${attWarrior} took ${dmg} reflected damage`;
    if (effect === 'Attack Cancelled') return `${a.attacker_name}'s ${attWarrior} attack on ${a.defender_name}'s ${defWarrior} was cancelled`;
    if (effect === 'Weapon Stolen')    return `${a.attacker_name}'s weapon was stolen and added to ${a.defender_name}'s hand`;
    if (effect === 'Drain Life')       return `${a.defender_name}'s ${defWarrior} absorbed ${dmg} damage and healed`;
    if (effect === 'Instant Kill')     return `${a.attacker_name}'s ${attWarrior} was instantly killed`;
    return ambushEffectDescriptions[effect] || '';
}

function showAmbushTriggeredModal(ambushTriggered) {
    const effectDisplay = ambushTriggered.effect_display;
    const colorClass = ambushEffectColorClass[effectDisplay] || '';
    const isInvolved = gameState.playerName === ambushTriggered.attacker_name ||
                       gameState.playerName === ambushTriggered.defender_name;

    if (isInvolved) {
        const modal = document.getElementById('ambush-notification-modal');
        const body  = document.getElementById('ambush-notification-body');
        if (!modal || !body) return;

        const isAttacker = gameState.playerName === ambushTriggered.attacker_name;
        const roleText = isAttacker
            ? `${ambushTriggered.defender_name}'s ambush triggered on your attack!`
            : `Your ambush triggered against ${ambushTriggered.attacker_name}!`;

        body.innerHTML =
            `<p class="ambush-notification-text">${roleText}</p>` +
            `<p class="ambush-notification-effect ambush-toast-effect ${colorClass}">${effectDisplay}</p>` +
            buildAmbushDetailRows(ambushTriggered);

        modal.classList.remove('hidden');

        if (ambushNotificationTimer) clearTimeout(ambushNotificationTimer);
        ambushNotificationTimer = setTimeout(() => hideAmbushNotificationModal(), 8000);
        return;
    }

    // Spectators: enriched toast
    const container = document.getElementById('error-toast-container');
    if (!container) return;

    const toast = document.createElement('div');
    toast.className = 'error-toast ambush-toast';
    toast.innerHTML =
        '<span class="error-toast-icon">&#9888;</span>' +
        '<div class="error-toast-content">' +
            `<div class="error-toast-title">Ambush! &#8212; <span class="ambush-toast-effect ${colorClass}">${effectDisplay}</span></div>` +
            `<div class="error-toast-message">${buildAmbushSpectatorMessage(ambushTriggered)}</div>` +
        '</div>' +
        '<button class="error-toast-close" onclick="this.closest(\'.error-toast\').remove()">&#x2715;</button>';
    container.appendChild(toast);

    setTimeout(() => {
        toast.classList.add('hiding');
        setTimeout(() => toast.remove(), 300);
    }, 7000);
}

function hideAmbushNotificationModal() {
    const modal = document.getElementById('ambush-notification-modal');
    if (modal) modal.classList.add('hidden');
    if (ambushNotificationTimer) {
        clearTimeout(ambushNotificationTimer);
        ambushNotificationTimer = null;
    }
}

// Generic Cards Modal
function showCardsModal(cards, title, subtitle, showPositionIndicators = false) {
    let content = '';

    if (cards.length === 0) {
        content = '<p style="color: #b0b0b0;">No cards to show</p>';
    } else {
        cards.forEach((rawCard, index) => {
            const card = normalizeCard(rawCard);

            // Position indicator for deck cards
            let positionBadge = '';
            if (showPositionIndicators) {
                const positionLabel = index === 0 ? 'TOP' : `#${index + 1}`;
                const badgeStyle = index === 0
                    ? 'background: #4CAF50; color: white;'
                    : 'background: #666; color: #ccc;';
                positionBadge = `<span class="position-badge" style="${badgeStyle} padding: 2px 6px; border-radius: 4px; font-size: 10px; font-weight: bold; margin-left: 5px;">${positionLabel}</span>`;
            }

            content += renderCardForModal(card);
            if (positionBadge) {
                content += positionBadge;
            }
        });
    }

    showGameModal(title, subtitle, content, true);

    if (cardsModalTimer) clearTimeout(cardsModalTimer);
    cardsModalTimer = setTimeout(() => hideGameModal(), 5000);
}

// Bought Cards Modal
function showBoughtCardsModal(cards) {
    showCardsModal(cards, 'Cards Bought!', `You bought ${cards.length} card${cards.length > 1 ? 's' : ''}`);
}

// Traded Cards Modal
function showTradedCardsModal(cards) {
    showCardsModal(cards, 'Card Received!', 'You traded 3 cards for this');
}

// Forge Result Modal
function showForgeResultModal(cards) {
    showCardsModal(cards, 'Weapons Forged!', 'Your weapons have been combined');
}

// Game Over Modal
function showGameOverModal(isWinner, message, playerStats, gameStartedAt) {
    const modal = document.getElementById('gameover-modal');
    const iconElement = document.getElementById('gameover-modal-icon');
    const titleElement = document.getElementById('gameover-modal-title');
    const messageElement = document.getElementById('gameover-modal-message');
    const statsEl = document.getElementById('gameover-stats');

    if (isWinner) {
        iconElement.textContent = '🏆';
        titleElement.textContent = 'Victory!';
        titleElement.className = 'gameover-modal-title victory';
    } else {
        iconElement.textContent = '💀';
        titleElement.textContent = 'Defeat';
        titleElement.className = 'gameover-modal-title defeat';
    }

    messageElement.textContent = message;

    if (playerStats && playerStats.length > 0) {
        const rows = playerStats.map(s => {
            const crown = s.is_winner ? ' 👑' : '';
            const mvp = s.is_mvp ? ' ⭐' : '';
            return `<tr class="${s.is_winner ? 'stat-winner' : ''}">
                <td>${s.name}${crown}${mvp}</td>
                <td>${s.kills}</td>
                <td>${s.damage}</td>
                <td>${s.castle_value}</td>
            </tr>`;
        }).join('');
        let durationRow = '';
        if (gameStartedAt) {
            const elapsed = Math.floor((Date.now() - new Date(gameStartedAt)) / 1000);
            durationRow = `<tr class="stat-duration"><td colspan="4">⏱ Game duration: ${formatTime(elapsed)}</td></tr>`;
        }
        statsEl.innerHTML = `<table class="gameover-stats-table">
            <thead><tr><th>Player</th><th>Kills</th><th>Damage</th><th>Castle</th></tr></thead>
            <tbody>${rows}${durationRow}</tbody>
        </table>`;
        statsEl.classList.remove('hidden');
    } else {
        statsEl.classList.add('hidden');
    }

    modal.classList.remove('hidden');
}

function showEliminationModal(playerName, isSelf) {
    const modal = document.getElementById('game-modal');
    const title = document.getElementById('modal-title');
    const subtitle = document.getElementById('modal-subtitle');
    const cardsContainer = document.getElementById('modal-cards-container');
    const closeBtn = document.getElementById('modal-close-btn');

    if (isSelf) {
        title.textContent = 'You have been eliminated!';
        subtitle.textContent = 'All your warriors have fallen. You are out of the battle.';
        cardsContainer.innerHTML = '<div style="font-size: 4em; text-align: center; padding: 20px;">💀</div>';
    } else {
        title.textContent = `${playerName} eliminated!`;
        subtitle.textContent = `${playerName} has lost all warriors and is out of the battle.`;
        cardsContainer.innerHTML = '<div style="font-size: 4em; text-align: center; padding: 20px;">⚔️</div>';
    }

    closeBtn.textContent = 'Continue';
    closeBtn.onclick = () => modal.classList.add('hidden');
    modal.classList.remove('hidden');
}

// Error Toast
function showErrorToast(message) {
    const container = document.getElementById('error-toast-container');

    const toast = document.createElement('div');
    toast.className = 'error-toast';
    toast.innerHTML = `
        <span class="error-toast-icon">⚠️</span>
        <div class="error-toast-content">
            <div class="error-toast-title">Error</div>
            <div class="error-toast-message">${message}</div>
        </div>
        <button class="error-toast-close" onclick="this.parentElement.remove()">×</button>
    `;

    container.appendChild(toast);

    // Auto-remove after 7 seconds
    setTimeout(() => {
        toast.classList.add('hiding');
        setTimeout(() => toast.remove(), 300);
    }, 7000);
}

// ===================== Event Banner + Notifications =====================

// Wraps modifier numbers in the event description with colored spans.
// Positive values → green, negative values → red.
function formatEventDesc(desc, eventType) {
    // Escape HTML entities first to prevent XSS
    let html = desc
        .replace(/&/g, '&amp;')
        .replace(/</g, '&lt;')
        .replace(/>/g, '&gt;');

    if (eventType === 'curse' || eventType === 'harvest') {
        // These descriptions use explicit +/- signs (e.g. "-2" or "+3")
        html = html.replace(/([+-]\d+)/g, (match) => {
            const cls = match[0] === '+' ? 'event-num-positive' : 'event-num-negative';
            return '<span class="' + cls + '">' + match + '</span>';
        });
    } else if (eventType === 'plague') {
        // Format: "lose N HP" → red, "gain N HP" → green
        html = html.replace(/\b(gain|lose)\s+(\d+)/g, (match, verb, num) => {
            const cls = verb === 'gain' ? 'event-num-positive' : 'event-num-negative';
            return verb + ' <span class="' + cls + '">' + num + '</span>';
        });
    } else if (eventType === 'abundance' || eventType === 'bloodlust') {
        // Bare digit → always positive/green
        html = html.replace(/\b(\d+)\b/g, (match) => '<span class="event-num-positive">' + match + '</span>');
    }

    return html;
}

function renderEventBanner(gs) {
    const banner = document.getElementById('event-banner');
    if (!banner) return;
    const eventType = gs.current_event || '';
    banner.dataset.event = eventType;
    banner.querySelector('.event-banner-name').textContent = gs.current_event_display || 'Calm';
    const desc = gs.current_event_description || 'No special effects this round';
    banner.querySelector('.event-banner-desc').innerHTML = formatEventDesc(desc, eventType);
}

function showEventTurnModal(gs) {
    const modal = document.getElementById('event-turn-modal');
    if (!modal) return;
    const content = document.getElementById('event-turn-modal-content');
    content.dataset.event = gs.current_event || '';
    const eventType = gs.current_event || '';
    const eventIcons = { curse: '⚔️', harvest: '🌾', plague: '☠️', abundance: '💰', bloodlust: '🩸', calm: '🌿' };
    const iconEl = document.getElementById('event-turn-modal-icon');
    if (iconEl) iconEl.textContent = eventIcons[eventType] || '🌐';
    document.getElementById('event-turn-modal-name').textContent = gs.current_event_display || '';
    const modalDesc = gs.current_event_description || '';
    document.getElementById('event-turn-modal-desc').innerHTML = formatEventDesc(modalDesc, eventType);
    modal.classList.remove('hidden');
    if (eventTurnModalTimer) clearTimeout(eventTurnModalTimer);
    eventTurnModalTimer = setTimeout(() => hideEventTurnModal(), 5000);
}

function hideEventTurnModal() {
    const modal = document.getElementById('event-turn-modal');
    if (modal) modal.classList.add('hidden');
    if (eventTurnModalTimer) { clearTimeout(eventTurnModalTimer); eventTurnModalTimer = null; }
}

// Event Change Toast (shown to non-active players when the round event changes)
function showEventChangeToast(gs) {
    const container = document.getElementById('error-toast-container');
    if (!container) return;
    const name = gs.current_event_display || 'Calm';
    const desc = gs.current_event_description || 'No special effects this round';
    const toast = document.createElement('div');
    toast.className = 'error-toast event-change-toast';
    toast.dataset.event = gs.current_event || '';
    toast.innerHTML =
        '<span class="error-toast-icon">&#9670;</span>' +
        '<div class="error-toast-content">' +
            '<div class="error-toast-title">New Round &#8212; ' + name + '</div>' +
            '<div class="error-toast-message">' + desc + '</div>' +
        '</div>' +
        '<button class="error-toast-close" onclick="this.closest(\'.error-toast\').remove()">&#x2715;</button>';
    container.appendChild(toast);
    setTimeout(() => {
        toast.classList.add('hiding');
        setTimeout(() => toast.remove(), 300);
    }, 7000);
}

// Waiting for Reconnect Modal
let waitingReconnectCountdownId = null;

function showWaitingForReconnectModal(disconnectedPlayers, secsUntilGameEnds) {
    const modal = document.getElementById('waiting-reconnect-modal');
    const playersEl = document.getElementById('waiting-reconnect-players');
    const countdownEl = document.getElementById('waiting-reconnect-countdown');
    if (!modal || !playersEl || !countdownEl) return;

    playersEl.innerHTML = disconnectedPlayers
        .map(name => `<div class="waiting-reconnect-player">${name} disconnected&hellip;</div>`)
        .join('');

    let secsLeft = secsUntilGameEnds;
    const fmt = s => s > 60 ? `${Math.floor(s / 60)}m ${s % 60}s` : `${s}s`;
    countdownEl.textContent = fmt(secsLeft);

    clearInterval(waitingReconnectCountdownId);
    waitingReconnectCountdownId = setInterval(() => {
        secsLeft = Math.max(0, secsLeft - 1);
        countdownEl.textContent = fmt(secsLeft);
        if (secsLeft === 0) clearInterval(waitingReconnectCountdownId);
    }, 1000);

    modal.classList.remove('hidden');
}

function hideWaitingForReconnectModal() {
    clearInterval(waitingReconnectCountdownId);
    waitingReconnectCountdownId = null;
    const modal = document.getElementById('waiting-reconnect-modal');
    if (modal) modal.classList.add('hidden');
}
