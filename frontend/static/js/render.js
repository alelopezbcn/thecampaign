function renderGameBoard(status) {
    // Render all opponent boards
    renderOpponents(status.opponents || []);

    // Render own board header (name + field HP)
    const playerHeader = document.getElementById('player-board-header');
    if (playerHeader) {
        const fieldHP = status.current_player_field_hp || 0;
        const hpBadge = fieldHP > 0 ? `<span class="field-hp-badge">${fieldHP} HP</span>` : '';
        playerHeader.innerHTML = `<span class="opponent-name">${status.current_player || ''}</span>${hpBadge}`;
    }

    // Active player board glow
    const playerBoard = document.querySelector('.player-board');
    if (playerBoard) {
        if (status.turn_player === status.current_player) {
            playerBoard.classList.add('active-turn');
        } else {
            playerBoard.classList.remove('active-turn');
        }
    }

    // Render player field
    renderCards('player-field', status.current_player_field);
    if (status.current_player_ambush_in_field) {
        prependFaceDownAmbushCard(document.getElementById('player-field'));
    }

    // Render player hand
    renderCards('player-hand', status.current_player_hand);

    // Render player castle
    renderCastle('player-castle', status.current_player_castle);

    // Render cemetery
    renderCemetery(status.cemetery);

    // Render discard pile
    renderDiscardPile(status.discard_pile);

    // Render deck
    renderDeck(status.cards_in_deck);

    // Render history
    renderHistory(status.history);

    // Show/hide player eliminated/disconnected overlay
    updatePlayerEliminatedState(status.is_eliminated || status.is_disconnected);
}

function updatePlayerEliminatedState(isEliminated) {
    const playerBoard = document.querySelector('.player-board');
    const actionPanel = document.querySelector('.action-panel');

    if (!isEliminated) {
        playerBoard?.classList.remove('eliminated');
        // Remove overlay if exists
        document.getElementById('player-eliminated-overlay')?.remove();
        return;
    }

    playerBoard?.classList.add('eliminated');
    actionPanel?.classList.add('eliminated');

    // Add overlay if not already present
    if (!document.getElementById('player-eliminated-overlay')) {
        const overlay = document.createElement('div');
        overlay.id = 'player-eliminated-overlay';
        overlay.className = 'player-eliminated-overlay';
        overlay.innerHTML = `
            <div class="eliminated-overlay-content">
                <div class="eliminated-overlay-icon">💀</div>
                <div class="eliminated-overlay-text">ELIMINATED</div>
            </div>
        `;
        playerBoard?.appendChild(overlay);
    }

    // Disable all action buttons
    document.querySelectorAll('.action-buttons .btn').forEach(btn => {
        btn.disabled = true;
    });
}

// Tracks which opponent boards are collapsed (persists across re-renders)
const collapsedOpponents = new Set();

// Auto-collapse inactive boards setting (persisted to localStorage)
let autoCollapseInactive = localStorage.getItem('autoCollapseInactive') === 'true';

function initAutoCollapseCheckbox() {
    const checkbox = document.getElementById('auto-collapse-inactive');
    if (!checkbox) return;
    checkbox.checked = autoCollapseInactive;
    checkbox.addEventListener('change', () => {
        autoCollapseInactive = checkbox.checked;
        localStorage.setItem('autoCollapseInactive', autoCollapseInactive);
        if (!autoCollapseInactive) {
            collapsedOpponents.clear();
        }
        if (gameState.currentState) {
            renderOpponents(gameState.currentState.opponents || []);
        }
    });
}

function renderOpponents(opponents) {
    const container = document.getElementById('opponents-container');
    container.innerHTML = '';
    container.setAttribute('data-count', opponents.length);

    // Auto-collapse inactive boards if the setting is on
    if (autoCollapseInactive) {
        const turnPlayer = gameState.currentState && gameState.currentState.turn_player;
        opponents.forEach(o => {
            if (o.player_name !== turnPlayer) {
                collapsedOpponents.add(o.player_name);
            } else {
                collapsedOpponents.delete(o.player_name);
            }
        });
    }

    // In 2v2, place ally in center with enemies on left and right
    if (gameState.gameMode === '2v2' && opponents.length === 3) {
        const enemies = opponents.filter(o => !o.is_ally);
        const ally = opponents.find(o => o.is_ally);
        if (ally) {
            opponents = [enemies[0], ally, enemies[1]];
        }
    }

    opponents.forEach(opponent => {
        const board = document.createElement('div');
        board.className = 'opponent-board';
        board.dataset.opponentName = opponent.player_name;
        if (opponent.is_ally) board.classList.add('ally');
        if (opponent.is_eliminated || opponent.is_disconnected) board.classList.add('eliminated');
        if (gameState.currentState && opponent.player_name === gameState.currentState.turn_player) {
            board.classList.add('active-turn');
        }
        if (collapsedOpponents.has(opponent.player_name)) {
            board.classList.add('collapsed');
        }

        // Header
        const header = document.createElement('div');
        header.className = 'opponent-header';

        // Collapse toggle — header collapses; clicking anywhere on a collapsed board expands it
        const isCollapsed = collapsedOpponents.has(opponent.player_name);
        const collapseBtn = document.createElement('span');
        collapseBtn.className = 'opponent-collapse-btn';
        collapseBtn.textContent = isCollapsed ? '▶' : '◀';

        const toggleCollapse = (name) => {
            if (collapsedOpponents.has(name)) {
                collapsedOpponents.delete(name);
                board.classList.remove('collapsed');
                collapseBtn.textContent = '◀';
            } else {
                collapsedOpponents.add(name);
                board.classList.add('collapsed');
                collapseBtn.textContent = '▶';
            }
            container.classList.toggle('has-collapsed', container.querySelector('.opponent-board.collapsed') !== null);
        };

        // Header click: collapse (stop propagation so board handler doesn't also fire)
        header.addEventListener('click', (e) => {
            e.stopPropagation();
            toggleCollapse(opponent.player_name);
        });

        // Board click: expand only (ignored when already expanded)
        board.addEventListener('click', () => {
            if (collapsedOpponents.has(opponent.player_name)) {
                toggleCollapse(opponent.player_name);
            }
        });
        header.appendChild(collapseBtn);

        const disconnectEntry = disconnectCountdowns[opponent.player_name];
        const badgeHtml = opponent.is_eliminated
            ? '<span class="opponent-badge eliminated-badge">Eliminated</span>'
            : opponent.is_disconnected
                ? `<span class="opponent-badge eliminated-badge disconnect-countdown" data-player="${opponent.player_name}">Disconnected${disconnectEntry ? ` (${disconnectEntry.secondsLeft}s)` : ''}</span>`
                : '';
        const fieldHP = opponent.field_hp || 0;
        const hpBadge = fieldHP > 0 ? `<span class="field-hp-badge">${fieldHP} HP</span>` : '';
        const headerContent = document.createElement('div');
        headerContent.className = 'opponent-header-content';
        headerContent.innerHTML = `
            <span class="opponent-name">${opponent.player_name}</span>
            ${hpBadge}
            ${opponent.is_ally ? '<span class="opponent-badge ally-badge">Ally</span>' : ''}
            ${badgeHtml}
        `;
        header.appendChild(headerContent);
        board.appendChild(header);

        // Internal layout
        const area = document.createElement('div');
        area.className = 'opponent-area';

        // Castle
        const castleArea = document.createElement('div');
        castleArea.className = 'opponent-castle-area';
        castleArea.innerHTML = '<h4>Castle</h4>';
        const castleDiv = document.createElement('div');
        castleDiv.className = 'castle';
        castleArea.appendChild(castleDiv);
        renderCastleInto(castleDiv, opponent.castle);
        area.appendChild(castleArea);

        // Field
        const fieldArea = document.createElement('div');
        fieldArea.className = 'opponent-field-area';
        fieldArea.innerHTML = '<h4>Field</h4>';
        const fieldDiv = document.createElement('div');
        fieldDiv.className = 'opponent-field field card-container';
        fieldDiv.dataset.opponentName = opponent.player_name;
        fieldArea.appendChild(fieldDiv);
        area.appendChild(fieldArea);

        // Render field cards
        const fieldCards = opponent.field || [];
        if (fieldCards.length === 0 && !opponent.ambush_in_field) {
            fieldDiv.innerHTML = '<div style="color: #666; padding: 10px; font-size: 0.85em;">No warriors</div>';
        } else {
            if (opponent.ambush_in_field) {
                prependFaceDownAmbushCard(fieldDiv);
            }
            fieldCards.forEach(card => {
                const cardElement = createCardElement(card, `opponent-field:${opponent.player_name}`);
                fieldDiv.appendChild(cardElement);
            });
        }

        board.appendChild(area);

        // Hand
        const handArea = document.createElement('div');
        handArea.className = 'opponent-hand-area';
        handArea.innerHTML = `<h4>Hand (${opponent.cards_in_hand || 0})</h4>`;
        const handDiv = document.createElement('div');
        handDiv.className = 'opponent-hand';
        renderOpponentHandInto(handDiv, opponent.cards_in_hand || 0);
        handArea.appendChild(handDiv);
        board.appendChild(handArea);

        // Collapsed summary — shown when board is collapsed
        const castleGoal = (opponent.castle && opponent.castle.resources_to_win) || (gameState.gameMode === '2v2' ? 30 : 25);
        const castleStr = (opponent.castle && opponent.castle.constructed)
            ? `${opponent.castle.value}/${castleGoal}`
            : '—';
        const warriorStr = fieldCards.length > 0
            ? `${fieldCards.length} (${fieldHP}HP)`
            : '0';
        const collapsedSummary = document.createElement('div');
        collapsedSummary.className = 'opponent-collapsed-summary';
        collapsedSummary.innerHTML = `
            <div class="collapsed-player-name">${opponent.player_name}</div>
            <div class="collapsed-stat" title="Castle value">🏰 ${castleStr}</div>
            <div class="collapsed-stat" title="Warriors / total HP">⚔ ${warriorStr}</div>
            <div class="collapsed-stat" title="Cards in hand">🃏 ${opponent.cards_in_hand || 0}</div>
        `;
        board.appendChild(collapsedSummary);

        container.appendChild(board);
    });

    // Sync container flex layout when any board is collapsed
    container.classList.toggle('has-collapsed', container.querySelector('.opponent-board.collapsed') !== null);
}

function renderCastleInto(container, castle) {
    if (!castle) {
        container.innerHTML = '<div class="castle-status">Not Constructed</div>';
        return;
    }

    const isConstructed = castle.constructed || false;
    const resourceCount = castle.resource_cards || 0;
    const castleValue = castle.value || 0;
    const isProtected = castle.is_protected || false;

    container.className = 'castle';
    if (isConstructed) container.classList.add('constructed');
    if (isProtected) container.classList.add('fortified');

    if (isConstructed) {
        const castleGoal = castle.resources_to_win || (gameState.gameMode === '2v2' ? 30 : 25);
        const progressPct = Math.min(100, (castleValue / castleGoal) * 100);
        const fortressIndicator = isProtected
            ? `<div class="castle-fortress-indicator" title="Protected by a Fortress wall"><svg class="fortress-shield-icon" viewBox="0 0 80 96" xmlns="http://www.w3.org/2000/svg"><defs><linearGradient id="shieldGrad" x1="30%" y1="0%" x2="70%" y2="100%"><stop offset="0%" stop-color="#3b82f6"/><stop offset="100%" stop-color="#1e3a8a"/></linearGradient></defs><path d="M40 4 L76 18 L76 52 Q76 80 40 92 Q4 80 4 52 L4 18 Z" fill="url(#shieldGrad)" stroke="#d4a825" stroke-width="4.5" stroke-linejoin="round"/><line x1="40" y1="24" x2="40" y2="70" stroke="rgba(255,255,255,0.4)" stroke-width="3" stroke-linecap="round"/><line x1="21" y1="47" x2="59" y2="47" stroke="rgba(255,255,255,0.4)" stroke-width="3" stroke-linecap="round"/></svg><div class="fortress-badge">Fortified</div></div>`
            : '';
        container.innerHTML = `
            <div class="castle-image">
                <img src="/static/img/cards/castle.webp" alt="Castle" draggable="false">
            </div>
            ${fortressIndicator}
            <div class="castle-progress">
                <div class="castle-progress-bar">
                    <div class="castle-progress-fill" style="width: ${progressPct}%"></div>
                </div>
                <div class="castle-progress-label">${castleValue}/${castleGoal}</div>
            </div>
        `;
    } else {
        container.innerHTML = `
            <div class="castle-icon"></div>
            <div class="castle-status">Not Constructed</div>
        `;
    }
}

function renderOpponentHandInto(container, cardCount) {
    container.innerHTML = '';
    if (!cardCount || cardCount === 0) {
        container.innerHTML = '<div style="color: #666; font-size: 0.8em;">No cards</div>';
        return;
    }
    for (let i = 0; i < cardCount; i++) {
        const cardBack = document.createElement('div');
        cardBack.className = 'opponent-hand-card';
        cardBack.innerHTML = `
            <div class="card-back-mini">
                <div class="card-back-mini-inner">
                    <span class="card-back-mini-emblem">⚔</span>
                </div>
            </div>
        `;
        container.appendChild(cardBack);
    }
}

function renderCards(containerId, cards) {
    const container = document.getElementById(containerId);
    container.innerHTML = '';

    if (!cards || cards.length === 0) {
        container.innerHTML = '<div style="color: #666; padding: 20px;">No cards</div>';
        return;
    }

    cards.forEach(card => {
        const cardElement = createCardElement(card, containerId);
        container.appendChild(cardElement);
    });
}


function createCardElement(card, context) {
    const div = document.createElement('div');
    div.className = 'card animating';
    div.dataset.cardId = card.id || generateCardID(card);

    // Store card data for attack phase logic
    div.dataset.cardType = card.type || '';
    div.dataset.cardSubType = card.sub_type || '';
    if (card.use_on) {
        div.dataset.useOn = JSON.stringify(card.use_on);
    }

    // Determine card type
    const cardType = getCardType(card);
    div.classList.add(cardType);

    // Check if we have an image for this card
    const imageUrl = getCardImageUrl(card);
    if (imageUrl) {
        div.classList.add('has-image');
    }

    // Apply card color from backend
    if (card.color) {
        if (!imageUrl) {
            const bgColor = hexToRgba(card.color, 0.3);
            div.style.setProperty('background', bgColor, 'important');
        }
        div.style.setProperty('border-color', card.color, 'important');
    }

    // Check if card can be used based on current action (only when it's your turn)
    const status = gameState.currentState;
    const currentAction = gameState.currentAction;

    if (context === 'player-hand' && gameState.isYourTurn) {
        // During endturn phase, all cards are disabled
        if (status && status.current_action === 'endturn') {
            div.classList.add('unusable');
        }
        // During trade action, only cards that can be traded are usable
        else if (currentAction === 'trade') {
            if (card.can_be_traded === false) {
                div.classList.add('unusable');
            } else {
                div.classList.add('usable');
            }
        }
        // During move_warrior action, only warriors are usable
        else if (currentAction === 'move_warrior') {
            if (cardType !== 'warrior') {
                div.classList.add('unusable');
            } else {
                div.classList.add('usable');
            }
        }
        // During forge action, only forgeable weapons are usable
        else if (currentAction === 'forge') {
            const forgeableTypes = ['Sword', 'Arrow', 'Poison'];
            if (cardType !== 'weapon' || !forgeableTypes.includes(card.sub_type)) {
                div.classList.add('unusable');
            } else {
                div.classList.add('usable');
            }
        }
        // Warriors are only usable during move_warrior action
        else if (cardType === 'warrior') {
            div.classList.add('unusable');
        }
        // During action phases (attack, spy/steal, buy, construct), use backend can_be_used flag
        else if (status && ['attack', 'spy/steal', 'buy', 'construct'].includes(status.current_action)) {
            if (card.can_be_used === false) {
                div.classList.add('unusable');
            } else {
                div.classList.add('usable');
            }
        }
    }

    // Check if this is a newly drawn card and highlight it
    const newCards = gameState.newlyDrawnCards || [];
    console.log('Creating card:', div.dataset.cardId, 'newlyDrawnCards:', newCards);
    if (newCards.includes(div.dataset.cardId)) {
        console.log('MATCH! Adding newly-drawn class to:', div.dataset.cardId);
        div.classList.add('newly-drawn');
        // Remove the highlight after animation completes
        setTimeout(() => {
            div.classList.remove('newly-drawn');
        }, 5000);
    }

    // Check if card is protected (field cards only)
    const isProtected = card.protected_by && card.protected_by.id;
    const shieldHtml = isProtected ? `
        <div class="card-shield">
            <span class="shield-icon">🛡️</span>
            <span class="shield-value">${card.protected_by.value || '?'}</span>
        </div>
    ` : '';

    // Kill counter badge — shown on warriors with at least one kill
    const cardKills = card.kills || 0;
    const killsBadgeHtml = (cardType === 'warrior' && cardKills > 0)
        ? `<div class="kills-badge" title="${cardKills} kill${cardKills !== 1 ? 's' : ''} — each kill adds +1 damage to this warrior's attacks">💀 ${cardKills}</div>`
        : '';

    // Create card HTML
    if (imageUrl) {
        div.innerHTML = `
            ${shieldHtml}
            ${killsBadgeHtml}
            <div class="card-image">
                <img src="${imageUrl}" alt="${getCardName(card)}" draggable="false">
            </div>
            <div class="card-info">
                <span class="card-name">${getCardName(card)}</span>
                ${getCardStatBadge(card, cardType)}
            </div>
        `;
    } else {
        div.innerHTML = `
            ${shieldHtml}
            ${killsBadgeHtml}
            <div class="card-header">
                <span class="card-id">${div.dataset.cardId.substring(0, 6)}</span>
                <span class="card-type ${cardType}">${card.type || cardType}</span>
            </div>
            <div class="card-content">
                <div class="card-name">${getCardName(card)}</div>
                ${getCardStats(card, cardType)}
            </div>
        `;
    }

    // Add protected class for styling
    if (isProtected) {
        div.classList.add('protected');
    }

    // Add tooltip for new players
    const cardKey = (card.sub_type || card.type || '').toLowerCase();
    const description = cardConfig[cardKey]?.description;
    if (description) {
        div.dataset.tooltip = description;
    }

    // Add click handler
    if (context === 'player-hand' || context.startsWith('opponent-field:') || context === 'player-field') {
        div.addEventListener('click', () => {
            handleCardClick(div.dataset.cardId, cardType, context, card);
        });
    }

    return div;
}

function renderCastle(containerId, castle) {
    const container = document.getElementById(containerId);

    if (!castle) {
        container.innerHTML = '<div class="castle-status">Not Constructed</div>';
        return;
    }

    const isConstructed = castle.constructed || false;
    const resourceCount = castle.resource_cards || 0;
    const castleValue = castle.value || 0;
    const isProtected = castle.is_protected || false;

    container.className = 'castle';
    if (isConstructed) container.classList.add('constructed');
    if (isProtected) container.classList.add('fortified');

    if (isConstructed) {
        const castleGoal = castle.resources_to_win || (gameState.gameMode === '2v2' ? 30 : 25);
        const progressPct = Math.min(100, (castleValue / castleGoal) * 100);
        const fortressIndicator = isProtected
            ? `<div class="castle-fortress-indicator" title="Protected by a Fortress wall"><svg class="fortress-shield-icon" viewBox="0 0 80 96" xmlns="http://www.w3.org/2000/svg"><defs><linearGradient id="shieldGrad" x1="30%" y1="0%" x2="70%" y2="100%"><stop offset="0%" stop-color="#3b82f6"/><stop offset="100%" stop-color="#1e3a8a"/></linearGradient></defs><path d="M40 4 L76 18 L76 52 Q76 80 40 92 Q4 80 4 52 L4 18 Z" fill="url(#shieldGrad)" stroke="#d4a825" stroke-width="4.5" stroke-linejoin="round"/><line x1="40" y1="24" x2="40" y2="70" stroke="rgba(255,255,255,0.4)" stroke-width="3" stroke-linecap="round"/><line x1="21" y1="47" x2="59" y2="47" stroke="rgba(255,255,255,0.4)" stroke-width="3" stroke-linecap="round"/></svg><div class="fortress-badge">Fortified</div></div>`
            : '';
        container.innerHTML = `
            <div class="castle-image">
                <img src="/static/img/cards/castle.webp" alt="Castle" draggable="false">
            </div>
            <div class="castle-progress">
                <div class="castle-progress-bar">
                    <div class="castle-progress-fill" style="width: ${progressPct}%"></div>
                </div>
                <div class="castle-progress-label">${castleValue}/${castleGoal}</div>
            </div>
            ${fortressIndicator}
        `;
    } else {
        container.innerHTML = `
            <div class="castle-icon"></div>
            <div class="castle-status">Not Constructed</div>
        `;
    }
}

function renderCemetery(cemetery) {
    const countElement = document.getElementById('cemetery-corps-count');
    const lastCorpContainer = document.getElementById('cemetery-last-corp');

    if (!cemetery) {
        countElement.textContent = '0';
        lastCorpContainer.innerHTML = '';
        return;
    }

    countElement.textContent = cemetery.corps || 0;

    if (cemetery.last_corp) {
        lastCorpContainer.innerHTML = '';
        const cardElement = createCardElement(cemetery.last_corp, 'cemetery');
        lastCorpContainer.appendChild(cardElement);
    } else {
        lastCorpContainer.innerHTML = '';
    }
}

function renderDiscardPile(discardPile) {
    const countElement = document.getElementById('discard-pile-cards-count');
    const lastCardContainer = document.getElementById('discard-pile-last-card');

    if (!discardPile) {
        countElement.textContent = '0';
        lastCardContainer.innerHTML = '';
        return;
    }

    countElement.textContent = discardPile.cards || 0;

    if (discardPile.last_card) {
        lastCardContainer.innerHTML = '';
        const cardElement = createCardElement(discardPile.last_card, 'discard-pile');
        lastCardContainer.appendChild(cardElement);
    } else {
        lastCardContainer.innerHTML = '';
    }
}

function renderDeck(cardsInDeck) {
    const countElement = document.getElementById('deck-cards-count');
    const deckElement = document.getElementById('deck');

    if (countElement) {
        countElement.textContent = cardsInDeck || 0;
    }

    // Add visual effect for low deck count
    if (deckElement) {
        deckElement.classList.remove('deck-low', 'deck-empty');
        if (cardsInDeck === 0) {
            deckElement.classList.add('deck-empty');
        } else if (cardsInDeck <= 5) {
            deckElement.classList.add('deck-low');
        }
    }
}

function addErrorToHistory(message) {
    const container = document.getElementById('history-list');
    if (!container) return;

    // Remove "No events yet" placeholder if present
    const empty = container.querySelector('.history-empty');
    if (empty) empty.remove();

    gameState.historyMessages.push({ text: message, isError: true });

    const item = document.createElement('div');
    item.className = 'history-item history-error';
    item.textContent = message;
    container.appendChild(item);

    container.scrollTop = container.scrollHeight;
}

function renderHistory(newMessages) {
    const container = document.getElementById('history-list');
    if (!container) return;

    // Accumulate new messages (backend sends {msg, color} objects)
    if (newMessages && newMessages.length > 0) {
        gameState.historyMessages = gameState.historyMessages.concat(newMessages);
    }

    container.innerHTML = '';

    if (gameState.historyMessages.length === 0) {
        container.innerHTML = '<div class="history-empty">No events yet</div>';
        return;
    }

    gameState.historyMessages.forEach(message => {
        const item = document.createElement('div');
        const text = message.msg || message.text || message;
        const color = message.color;
        item.className = 'history-item' + (message.isError ? ' history-error' : '');
        item.textContent = text;
        if (color && !message.isError) {
            item.style.borderLeftColor = color;
            item.style.color = color;
        }
        container.appendChild(item);
    });

    // Scroll to bottom to show latest message
    container.scrollTop = container.scrollHeight;
}

// Ambush card in field — shows the card image (effect hidden until triggered)
function prependFaceDownAmbushCard(container) {
    const div = document.createElement('div');
    div.className = 'card ambush-face-down has-image';
    div.dataset.tooltip = 'Ambush — Triggers on weapon attacks only (not Special Powers). Possible effects: Reflect Damage (23%), Cancel Attack (23%), Steal Weapon (23%), Drain Life (23%), Instant Kill (8%)';
    div.style.backgroundImage = 'url(/static/img/cards/ambush.webp)';
    div.style.backgroundSize = 'cover';
    div.style.backgroundPosition = 'center';
    container.prepend(div);
}
