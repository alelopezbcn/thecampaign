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
    // resetActionState already clears visual selections
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

// Turn Transition Modal
function showTurnTransitionModal(playerName, duration = 3000, overrideText = null) {
    // Don't show on the very first state (game just started)
    if (!gameState.currentState) return;

    const modal = document.getElementById('turn-transition-modal');
    const playerEl = document.getElementById('turn-transition-player');
    const bar = document.getElementById('turn-transition-bar');

    if (!modal || !playerEl || !bar) return;

    if (overrideText) {
        playerEl.textContent = overrideText;
    } else {
        const isYou = playerName === gameState.playerName;
        playerEl.textContent = isYou ? 'Your Turn!' : `${playerName}'s Turn`;
    }

    // Reset and start countdown bar animation
    bar.style.animation = 'none';
    bar.offsetHeight; // force reflow
    bar.style.animation = `countdown ${duration / 1000}s linear forwards`;

    modal.classList.remove('hidden');

    // Clear any existing timer
    if (turnTransitionTimer) clearTimeout(turnTransitionTimer);
    turnTransitionTimer = setTimeout(() => {
        hideTurnTransitionModal();
    }, duration);
}

function hideTurnTransitionModal() {
    const modal = document.getElementById('turn-transition-modal');
    modal.classList.add('hidden');
    if (turnTransitionTimer) {
        clearTimeout(turnTransitionTimer);
        turnTransitionTimer = null;
    }
}

function showStolenCardModal(card, action = 'stolen') {
    const modal = document.getElementById('stolen-card-modal');
    const container = document.getElementById('stolen-card-container');
    const text = document.getElementById('stolen-card-text');

    if (!modal || !container || !text) return;

    // Render the stolen card
    const cardName = card.sub_type || card.type;
    container.innerHTML = renderCardForModal(card);
    text.textContent = action === 'sabotaged'
        ? `${cardName} was sabotaged from you!`
        : `${cardName} was stolen from you!`;

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
    const modal = document.getElementById('champions-bounty-modal');
    const textEl = document.getElementById('champions-bounty-text');
    const fill = document.getElementById('champions-bounty-timer-fill');
    if (!modal || !textEl) return;

    const isYou = notification.earned_by === gameState.playerName;
    const cardWord = notification.cards === 1 ? 'card' : 'cards';
    textEl.textContent = isYou
        ? `You slew the champion's warrior and drew ${notification.cards} ${cardWord}!`
        : `${notification.earned_by} slew the champion's warrior and drew ${notification.cards} ${cardWord}!`;

    // Restart the timer bar animation
    if (fill) {
        fill.style.animation = 'none';
        fill.offsetHeight; // reflow
        fill.style.animation = 'bountyTimerShrink 5s linear forwards';
    }

    modal.classList.remove('hidden');

    if (championsBountyTimer) clearTimeout(championsBountyTimer);
    championsBountyTimer = setTimeout(() => hideChampionsBountyModal(), 5000);
}

function hideChampionsBountyModal() {
    const modal = document.getElementById('champions-bounty-modal');
    if (modal) modal.classList.add('hidden');
    if (championsBountyTimer) {
        clearTimeout(championsBountyTimer);
        championsBountyTimer = null;
    }
}

function showResurrectionModal(notification) {
    const modal = document.getElementById('resurrection-modal');
    const textEl = document.getElementById('resurrection-text');
    const fill = document.getElementById('resurrection-timer-fill');
    if (!modal || !textEl) return;

    const isYou = notification.player_name === gameState.playerName;
    const warriorName = notification.warrior_card?.sub_type || 'a warrior';
    const target = notification.target_player;
    const targetIsYou = target === gameState.playerName;

    const warriorImgEl = document.getElementById('resurrection-warrior-img');
    if (warriorImgEl) {
        const imageUrl = notification.warrior_card ? getCardImageUrl(notification.warrior_card) : null;
        if (imageUrl) {
            warriorImgEl.innerHTML = `<img src="${imageUrl}" alt="${warriorName}" class="resurrection-warrior-card-img">`;
        } else {
            warriorImgEl.innerHTML = '';
        }
    }

    let msg;
    if (isYou) {
        msg = targetIsYou || !target || target === notification.player_name
            ? `You resurrected ${warriorName} from the cemetery!`
            : `You resurrected ${warriorName} to ${target}'s field!`;
    } else {
        msg = targetIsYou
            ? `${notification.player_name} resurrected ${warriorName} to your field!`
            : target && target !== notification.player_name
                ? `${notification.player_name} resurrected ${warriorName} to ${target}'s field!`
                : `${notification.player_name} resurrected ${warriorName} from the cemetery!`;
    }
    textEl.textContent = msg;

    if (fill) {
        fill.style.animation = 'none';
        fill.offsetHeight;
        fill.style.animation = 'resurrectionTimerShrink 5s linear forwards';
    }

    modal.classList.remove('hidden');

    if (resurrectionTimer) clearTimeout(resurrectionTimer);
    resurrectionTimer = setTimeout(() => hideResurrectionModal(), 5000);
}

function hideResurrectionModal() {
    const modal = document.getElementById('resurrection-modal');
    if (modal) modal.classList.add('hidden');
    if (resurrectionTimer) {
        clearTimeout(resurrectionTimer);
        resurrectionTimer = null;
    }
}

function showAmbushTriggeredModal(effectDisplay) {
    const modal = document.getElementById('ambush-triggered-modal');
    const effectEl = document.getElementById('ambush-triggered-effect');
    const textEl = document.getElementById('ambush-triggered-text');
    if (!modal) return;
    if (effectEl) {
        effectEl.textContent = effectDisplay;
        effectEl.className = 'ambush-triggered-effect-label ' + (ambushEffectColorClass[effectDisplay] || '');
    }
    if (textEl) textEl.textContent = ambushEffectDescriptions[effectDisplay] || '';
    modal.classList.remove('hidden');

    if (ambushTriggeredTimer) clearTimeout(ambushTriggeredTimer);
    ambushTriggeredTimer = setTimeout(() => hideAmbushTriggeredModal(), 5000);
}

function hideAmbushTriggeredModal() {
    const modal = document.getElementById('ambush-triggered-modal');
    if (modal) modal.classList.add('hidden');
    if (ambushTriggeredTimer) {
        clearTimeout(ambushTriggeredTimer);
        ambushTriggeredTimer = null;
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
function showGameOverModal(isWinner, message, playerStats) {
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
        statsEl.innerHTML = `<table class="gameover-stats-table">
            <thead><tr><th>Player</th><th>Kills</th><th>Damage</th><th>Castle</th></tr></thead>
            <tbody>${rows}</tbody>
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

    // Auto-remove after 5 seconds
    setTimeout(() => {
        toast.classList.add('hiding');
        setTimeout(() => toast.remove(), 300);
    }, 5000);
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
    }, 4000);
}
