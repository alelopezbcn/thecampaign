// Game state
let ws = null;
let gameState = {
    playerName: '',
    gameID: '',
    isYourTurn: false,
    currentState: null,
    selectedCards: [],
    currentAction: null
};

// DOM Elements
const screens = {
    join: document.getElementById('join-screen'),
    waiting: document.getElementById('waiting-screen'),
    setup: document.getElementById('setup-screen'),
    game: document.getElementById('game-screen'),
    gameover: document.getElementById('gameover-screen')
};

// Initialize
document.addEventListener('DOMContentLoaded', () => {
    setupEventListeners();
});

function setupEventListeners() {
    // Join screen
    document.getElementById('join-btn').addEventListener('click', joinGame);
    document.getElementById('player-name').addEventListener('keypress', (e) => {
        if (e.key === 'Enter') joinGame();
    });
    document.getElementById('game-id').addEventListener('keypress', (e) => {
        if (e.key === 'Enter') joinGame();
    });

    // Setup screen
    document.getElementById('confirm-warriors-btn').addEventListener('click', confirmInitialWarriors);

    // Game screen actions
    document.getElementById('attack-btn').addEventListener('click', () => startAction('attack'));
    document.getElementById('move-warrior-btn').addEventListener('click', () => startAction('move_warrior'));
    document.getElementById('special-power-btn').addEventListener('click', () => startAction('special_power'));
    document.getElementById('trade-btn').addEventListener('click', () => startAction('trade'));
    document.getElementById('buy-btn').addEventListener('click', () => startAction('buy'));
    document.getElementById('construct-btn').addEventListener('click', () => startAction('construct'));
    document.getElementById('spy-btn').addEventListener('click', () => startAction('spy'));
    document.getElementById('steal-btn').addEventListener('click', () => startAction('steal'));
    document.getElementById('catapult-btn').addEventListener('click', () => startAction('catapult'));
    document.getElementById('end-turn-btn').addEventListener('click', () => sendAction('end_turn'));

    // Confirm/Cancel action buttons
    document.getElementById('confirm-attack-btn').addEventListener('click', confirmAttack);
    document.getElementById('cancel-action-btn').addEventListener('click', cancelAction);

    // Game over
    document.getElementById('new-game-btn').addEventListener('click', () => location.reload());
}

// WebSocket functions
function connectWebSocket() {
    const protocol = window.location.protocol === 'https:' ? 'wss:' : 'ws:';
    const wsUrl = `${protocol}//${window.location.host}/ws`;

    ws = new WebSocket(wsUrl);

    ws.onopen = () => {
        console.log('WebSocket connected');
        showStatus('connection-status', 'Connected to server', 'success');
    };

    ws.onmessage = (event) => {
        const message = JSON.parse(event.data);
        console.log('Received message:', message);
        handleMessage(message);
    };

    ws.onerror = (error) => {
        console.error('WebSocket error:', error);
        showStatus('connection-status', 'Connection error', 'error');
    };

    ws.onclose = () => {
        console.log('WebSocket closed');
        showStatus('connection-status', 'Disconnected from server', 'error');
    };
}

function sendMessage(type, payload = null) {
    if (ws && ws.readyState === WebSocket.OPEN) {
        const message = { type, payload };
        console.log('Sending message:', message);
        ws.send(JSON.stringify(message));
    } else {
        console.error('WebSocket not connected');
    }
}

// Message handlers
function handleMessage(message) {
    switch (message.type) {
        case 'error':
            handleError(message.payload);
            break;
        case 'player_joined':
            handlePlayerJoined(message.payload);
            break;
        case 'waiting_for_player':
            showWaitingScreen();
            break;
        case 'game_started':
            handleGameStarted(message.payload);
            break;
        case 'game_state':
            handleGameState(message.payload);
            break;
        case 'game_ended':
            handleGameEnded();
            break;
        case 'spy_result':
            handleSpyResult(message.payload);
            break;
        default:
            console.log('Unknown message type:', message.type);
    }
}

function handleError(payload) {
    console.error('Server error:', payload.message);
    showStatus('connection-status', payload.message, 'error');
    showStatus('setup-status', payload.message, 'error');
}

function handlePlayerJoined(payload) {
    console.log('Player joined:', payload.player_name);
}

function handleGameStarted(payload) {
    console.log('Game started:', payload);
    gameState.playerName = payload.your_name;
    gameState.gameID = payload.game_id;

    document.getElementById('current-game-id').textContent = payload.game_id;
    document.getElementById('player-name-display').textContent = payload.your_name;
    document.getElementById('game-id-display').textContent = `Game: ${payload.game_id}`;
}

function handleGameState(payload) {
    console.log('Game state updated:', payload);
    console.log('Newly drawn card from payload:', payload.newly_drawn_card);
    gameState.isYourTurn = payload.is_your_turn;
    gameState.currentState = payload.game_status;
    gameState.newlyDrawnCard = payload.newly_drawn_card || null;
    console.log('gameState.newlyDrawnCard set to:', gameState.newlyDrawnCard);

    // Reset action state when game state updates
    gameState.currentAction = null;
    gameState.selectedCards = [];
    hideConfirmButtons();
    updateActionPrompt('');

    // Determine which screen to show
    if (payload.game_status.current_player && !isSetupComplete(payload.game_status)) {
        showSetupScreen(payload.game_status);
    } else {
        showGameScreen(payload.game_status);
    }

    updateTurnIndicator();
}

function handleGameEnded() {
    showScreen('gameover');
    const winner = gameState.currentState ? gameState.currentState.current_player : 'Unknown';
    document.getElementById('gameover-title').textContent =
        winner === gameState.playerName ? 'Victory!' : 'Defeat';
    document.getElementById('gameover-message').textContent =
        `${winner} wins the game!`;
}

function handleSpyResult(payload) {
    alert('Spied cards: ' + JSON.stringify(payload.cards, null, 2));
}

// Screen management
function showScreen(screenName) {
    Object.values(screens).forEach(screen => screen.classList.add('hidden'));
    screens[screenName].classList.remove('hidden');
}

function showWaitingScreen() {
    document.getElementById('current-game-id').textContent = gameState.gameID;
    showScreen('waiting');
}

function showSetupScreen(status) {
    showScreen('setup');
    renderSetupHand(status);
    updateSetupTurnIndicator();
}

function showGameScreen(status) {
    showScreen('game');
    renderGameBoard(status);
    updateActionButtons();
}

// Game actions
function joinGame() {
    const playerName = document.getElementById('player-name').value.trim();
    const gameID = document.getElementById('game-id').value.trim();

    if (!playerName || !gameID) {
        showStatus('connection-status', 'Please enter both name and game ID', 'error');
        return;
    }

    gameState.playerName = playerName;
    gameState.gameID = gameID;

    connectWebSocket();

    // Send join message after connection
    setTimeout(() => {
        sendMessage('join_game', {
            player_name: playerName,
            game_id: gameID
        });
    }, 500);
}

function confirmInitialWarriors() {
    const selectedWarriors = gameState.selectedCards;

    if (selectedWarriors.length < 1 || selectedWarriors.length > 3) {
        showStatus('setup-status', 'Select 1-3 warriors', 'error');
        return;
    }

    sendMessage('set_initial_warriors', {
        warrior_ids: selectedWarriors
    });

    gameState.selectedCards = [];
}

function sendAction(actionType, payload = null) {
    sendMessage(actionType, payload);
    clearSelections();
    gameState.currentAction = null;
    updateActionPrompt('');
    hideConfirmButtons();
}

function confirmAttack() {
    if (gameState.selectedCards.length === 2) {
        sendAction('attack', {
            target_id: gameState.selectedCards[0],
            weapon_id: gameState.selectedCards[1]
        });
    }
}

function cancelAction() {
    clearSelections();
    gameState.currentAction = null;
    updateActionPrompt('');
    hideConfirmButtons();

    // Re-render hand to remove unusable styles
    if (gameState.currentState) {
        renderCards('player-hand', gameState.currentState.current_player_hand);
    }
}

function showConfirmButtons() {
    document.getElementById('confirm-action-buttons').classList.remove('hidden');
}

function hideConfirmButtons() {
    document.getElementById('confirm-action-buttons').classList.add('hidden');
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
}

function startAction(actionType) {
    clearSelections();
    hideConfirmButtons();
    gameState.currentAction = actionType;

    let prompt = '';
    switch (actionType) {
        case 'attack':
            prompt = 'Select an enemy target first';
            break;
        case 'move_warrior':
            prompt = 'Select a warrior from your hand';
            break;
        case 'special_power':
            prompt = 'Select: 1) Your warrior, 2) Target, 3) Special power card';
            break;
        case 'trade':
            prompt = 'Select 3 cards to trade';
            break;
        case 'buy':
            prompt = 'Select a resource card';
            break;
        case 'construct':
            prompt = 'Select a card to construct with';
            break;
        case 'spy':
            showSpyOptions();
            return;
        case 'steal':
            showStealOptions();
            return;
        case 'catapult':
            showCatapultOptions();
            return;
    }

    updateActionPrompt(prompt);

    // Re-render hand to apply unusable styles
    if (gameState.currentState) {
        renderCards('player-hand', gameState.currentState.current_player_hand);
    }
}

function showSpyOptions() {
    const option = confirm('Spy options:\nOK = Spy top 5 cards from deck\nCancel = Spy enemy hand');
    sendAction('spy', { option: option ? 1 : 2 });
}

function showStealOptions() {
    const enemyCardCount = gameState.currentState.cards_in_enemy_hand;
    const position = prompt(`Enter card position to steal (1-${enemyCardCount}):`);

    if (position && !isNaN(position)) {
        sendAction('steal', { card_position: parseInt(position) });
    }
}

function showCatapultOptions() {
    const resourceCount = gameState.currentState.resource_cards_in_enemy_castle;
    const position = prompt(`Enter resource position to attack (1-${resourceCount}):`);

    if (position && !isNaN(position)) {
        sendAction('catapult', { card_position: parseInt(position) });
    }
}

// Card selection
function handleCardClick(cardID, cardType, context) {
    if (!gameState.isYourTurn) return;

    const action = gameState.currentAction;

    if (context === 'setup') {
        toggleCardSelection(cardID, 'setup');
        document.getElementById('confirm-warriors-btn').disabled =
            gameState.selectedCards.length < 1 || gameState.selectedCards.length > 3;
        return;
    }

    if (!action) return;

    gameState.selectedCards.push(cardID);
    highlightSelectedCard(cardID);

    // Handle different actions
    switch (action) {
        case 'attack':
            if (gameState.selectedCards.length === 1) {
                updateActionPrompt('Now select a weapon from your hand');
            } else if (gameState.selectedCards.length === 2) {
                updateActionPrompt('Attack ready! Confirm or cancel.');
                showConfirmButtons();
            }
            break;
        case 'special_power':
            if (gameState.selectedCards.length === 3) {
                sendAction('special_power', {
                    user_id: gameState.selectedCards[0],
                    target_id: gameState.selectedCards[1],
                    weapon_id: gameState.selectedCards[2]
                });
            }
            break;
        case 'move_warrior':
            sendAction('move_warrior', { warrior_id: cardID });
            break;
        case 'trade':
            if (gameState.selectedCards.length === 3) {
                sendAction('trade', { card_ids: gameState.selectedCards });
            } else {
                updateActionPrompt(`Selected ${gameState.selectedCards.length}/3 cards for trade`);
            }
            break;
        case 'buy':
            sendAction('buy', { card_id: cardID });
            break;
        case 'construct':
            sendAction('construct', { card_id: cardID });
            break;
    }
}

function toggleCardSelection(cardID, context) {
    const index = gameState.selectedCards.indexOf(cardID);
    if (index > -1) {
        gameState.selectedCards.splice(index, 1);
    } else {
        gameState.selectedCards.push(cardID);
    }

    // Update visual selection
    const container = context === 'setup' ?
        document.getElementById('setup-hand') :
        document.getElementById('player-hand');

    const cardElement = container.querySelector(`[data-card-id="${cardID}"]`);
    if (cardElement) {
        cardElement.classList.toggle('selected');
    }
}

// Rendering functions
function renderSetupHand(status) {
    const container = document.getElementById('setup-hand');
    container.innerHTML = '';

    status.current_player_hand.forEach(card => {
        if (isWarrior(card)) {
            const cardElement = createCardElement(card, 'setup');
            container.appendChild(cardElement);
        }
    });
}

function renderGameBoard(status) {
    // Render enemy field
    renderCards('enemy-field', status.enemy_field);

    // Render player field
    renderCards('player-field', status.current_player_field);

    // Render player hand
    renderCards('player-hand', status.current_player_hand);

    // Render castles
    renderCastle('enemy-castle', status.enemy_castle);
    renderCastle('player-castle', status.current_player_castle);

    // Update enemy hand count
    document.getElementById('enemy-hand-count').textContent =
        `${status.cards_in_enemy_hand} cards in hand`;
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

    // Determine card type
    const cardType = getCardType(card);
    div.classList.add(cardType);

    // Apply card color from backend (with transparency for background)
    if (card.color) {
        const bgColor = hexToRgba(card.color, 0.3);
        div.style.setProperty('background', bgColor, 'important');
        div.style.setProperty('border-color', card.color, 'important');
    }

    // Check if card is unusable during attack phase
    const canBeUsedOnIDs = card.use_on || [];
    if (context === 'player-hand' && gameState.currentAction === 'attack') {
        const isUnusable = isCardUnusableDuringAttack(card, cardType, canBeUsedOnIDs);
        if (isUnusable) {
            div.classList.add('unusable');
        }
    }

    // Check if this is the newly drawn card and highlight it
    console.log('Creating card:', div.dataset.cardId, 'newlyDrawnCard:', gameState.newlyDrawnCard);
    if (gameState.newlyDrawnCard && div.dataset.cardId === gameState.newlyDrawnCard) {
        console.log('MATCH! Adding newly-drawn class to:', div.dataset.cardId);
        div.classList.add('newly-drawn');
        // Remove the highlight after animation completes
        setTimeout(() => {
            div.classList.remove('newly-drawn');
        }, 2000);
    }

    // Create card HTML
    div.innerHTML = `
        <div class="card-header">
            <span class="card-id">${div.dataset.cardId.substring(0, 6)}</span>
            <span class="card-type ${cardType}">${card.type || cardType}</span>
        </div>
        <div class="card-content">
            <div class="card-name">${getCardName(card)}</div>
            ${getCardStats(card, cardType)}
        </div>
    `;

    // Add click handler
    if (context === 'setup' || context === 'player-hand' || context === 'enemy-field' || context === 'player-field') {
        div.addEventListener('click', () => {
            handleCardClick(div.dataset.cardId, cardType, context);
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

    container.className = 'castle';
    if (isConstructed) {
        container.classList.add('constructed');
    }

    container.innerHTML = `
        <div class="castle-status">${isConstructed ? 'Constructed' : 'Not Constructed'}</div>
        <div class="castle-resources">
            ${isConstructed ? `<div>${resourceCount} resource cards</div>` : ''}
        </div>
    `;
}

// Helper functions
function hexToRgba(hex, alpha) {
    const r = parseInt(hex.slice(1, 3), 16);
    const g = parseInt(hex.slice(3, 5), 16);
    const b = parseInt(hex.slice(5, 7), 16);
    return `rgba(${r}, ${g}, ${b}, ${alpha})`;
}

function isCardUnusableDuringAttack(card, cardType, canBeUsedOnIDs) {
    const rawType = (card.type || '').toLowerCase();

    // Warriors, Resources, Spy, Thief are always unusable during attack
    if (cardType === 'warrior' || cardType === 'resource') {
        return true;
    }
    if (rawType === 'spy' || rawType === 'thief') {
        return true;
    }

    // Weapons and Special Powers need valid targets
    if (cardType === 'weapon' || rawType === 'specialpower') {
        return canBeUsedOnIDs.length === 0;
    }

    // Catapult needs enemy castle to be constructed
    if (rawType === 'catapult') {
        const enemyCastle = gameState.currentState?.enemy_castle;
        return !enemyCastle || !enemyCastle.constructed;
    }

    return true; // Unknown types are unusable
}

function getCardType(card) {
    // type is the category: Warrior, Weapon, Resource, SpecialPower, Spy, Thief, Catapult
    const typeName = card.type || '';
    const type = typeName.toLowerCase();

    if (type === 'warrior') return 'warrior';
    if (type === 'weapon') return 'weapon';
    if (type === 'resource') return 'resource';
    if (type === 'specialpower') return 'special';
    if (type === 'spy' || type === 'thief' || type === 'catapult') return 'special';
    return 'unknown';
}

function getCardName(card) {
    // sub_type is the specific type: Knight, Sword, etc.
    // Falls back to type if sub_type is empty
    if (card.sub_type) return card.sub_type;
    if (card.type) return card.type;
    return 'Unknown Card';
}

function getCardStats(card, cardType) {
    let stats = '<div class="card-stats">';

    if (cardType === 'warrior') {
        stats += `
            <div class="card-stat">
                <span class="card-stat-label">HP</span>
                <span class="card-stat-value">${card.value || 0}</span>
            </div>
        `;
    } else if (cardType === 'weapon') {
        stats += `
            <div class="card-stat">
                <span class="card-stat-label">DMG</span>
                <span class="card-stat-value">${card.value || 0}</span>
            </div>
        `;
    } else if (cardType === 'resource') {
        stats += `
            <div class="card-stat">
                <span class="card-stat-label">Value</span>
                <span class="card-stat-value">${card.value || 0}</span>
            </div>
        `;
    }

    stats += '</div>';
    return stats;
}

function isWarrior(card) {
    const type = getCardType(card);
    return type === 'warrior';
}

function isSetupComplete(status) {
    return status.current_player_field && status.current_player_field.length > 0;
}

function updateTurnIndicator() {
    const indicator = document.getElementById('turn-indicator');
    if (gameState.isYourTurn) {
        indicator.textContent = 'YOUR TURN';
        indicator.className = 'turn-indicator your-turn';
    } else {
        indicator.textContent = 'ENEMY TURN';
        indicator.className = 'turn-indicator enemy-turn';
    }
}

function updateSetupTurnIndicator() {
    const nameDisplay = document.getElementById('setup-player-name');
    const indicator = document.getElementById('setup-turn-indicator');

    nameDisplay.textContent = gameState.playerName;

    if (gameState.isYourTurn) {
        indicator.textContent = 'YOUR TURN - Select Warriors';
        indicator.className = 'turn-indicator your-turn';
        document.getElementById('setup-hand').style.opacity = '1';
        document.getElementById('confirm-warriors-btn').style.display = 'block';
    } else {
        indicator.textContent = 'WAITING - Opponent Selecting';
        indicator.className = 'turn-indicator enemy-turn';
        document.getElementById('setup-hand').style.opacity = '0.5';
        document.getElementById('confirm-warriors-btn').style.display = 'none';
    }
}

function updateActionButtons() {
    const isYourTurn = gameState.isYourTurn;
    const status = gameState.currentState;

    // Enable/disable all action buttons based on turn
    document.querySelectorAll('.btn-action, #end-turn-btn').forEach(btn => {
        btn.disabled = !isYourTurn;
    });

    if (isYourTurn && status) {
        // Customize based on available actions (using new boolean fields)
        document.getElementById('move-warrior-btn').disabled = !status.can_move_warrior;
        document.getElementById('attack-btn').disabled = !status.can_attack;
        document.getElementById('buy-btn').disabled = !status.can_buy;
        document.getElementById('construct-btn').disabled = !status.can_initiate_castle && !status.can_grow_castle;
        document.getElementById('spy-btn').disabled = !status.can_spy;
        document.getElementById('steal-btn').disabled = !status.can_steal;
        document.getElementById('catapult-btn').disabled = !status.can_catapult;
    }
}

function updateActionPrompt(text) {
    const prompt = document.getElementById('action-prompt');
    prompt.textContent = text;
    prompt.className = 'action-prompt' + (text ? ' active' : '');
}

function showStatus(elementId, message, type) {
    const element = document.getElementById(elementId);
    if (element) {
        element.textContent = message;
        element.className = `status-message ${type}`;
        setTimeout(() => {
            element.textContent = '';
            element.className = 'status-message';
        }, 5000);
    }
}

function generateCardID(card) {
    return `card_${Math.random().toString(36).substr(2, 9)}`;
}
