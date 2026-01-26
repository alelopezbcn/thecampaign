// Game state
let ws = null;
let gameState = {
    playerName: '',
    gameID: '',
    isYourTurn: false,
    currentState: null,
    selectedCards: [],
    currentAction: null,
    // Attack phase state
    attackState: {
        weaponId: null,
        weaponType: null, // 'weapon', 'specialpower', 'catapult'
        userId: null,     // For special power - the warrior using the power
        targetId: null    // Target enemy warrior
    }
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

    // Game screen actions - only 4 buttons
    document.getElementById('move-warrior-btn').addEventListener('click', () => startAction('move_warrior'));
    document.getElementById('trade-btn').addEventListener('click', () => startAction('trade'));
    document.getElementById('skip-phase-btn').addEventListener('click', handleSkipPhase);
    document.getElementById('end-turn-btn').addEventListener('click', () => sendAction('end_turn'));

    // Confirm/Cancel action buttons
    document.getElementById('confirm-action-btn').addEventListener('click', confirmAttackAction);
    document.getElementById('cancel-action-btn').addEventListener('click', cancelAttackAction);

    // Game over
    document.getElementById('new-game-btn').addEventListener('click', () => location.reload());
}

function handleSkipPhase() {
    const status = gameState.currentState;
    // If we're in the last phase (endturn), end the turn instead
    if (status && status.current_action === 'endturn') {
        sendAction('end_turn');
    } else {
        sendAction('skip_phase');
    }
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
    console.log('New cards from payload:', payload.game_status.new_cards);
    gameState.isYourTurn = payload.is_your_turn;
    gameState.currentState = payload.game_status;
    // Use the first card from new_cards array for highlighting
    gameState.newlyDrawnCards = payload.game_status.new_cards || [];
    console.log('gameState.newlyDrawnCards set to:', gameState.newlyDrawnCards);

    // Reset action state when game state updates
    gameState.currentAction = null;
    gameState.selectedCards = [];
    resetAttackState();
    updateActionPrompt('');

    // Determine which screen to show
    if (payload.game_status.current_player && !isSetupComplete(payload.game_status)) {
        showSetupScreen(payload.game_status);
    } else {
        showGameScreen(payload.game_status);
    }

    updateTurnIndicator();
    updatePhaseIndicator();
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
    gameState.currentAction = actionType;

    let prompt = '';
    switch (actionType) {
        case 'move_warrior':
            prompt = 'Select a warrior from your hand';
            break;
        case 'trade':
            prompt = 'Select 3 cards to trade';
            break;
    }

    updateActionPrompt(prompt);

    // Re-render hand to apply styles
    if (gameState.currentState) {
        renderCards('player-hand', gameState.currentState.current_player_hand);
    }
}

// Card selection
function handleCardClick(cardID, cardType, context, card = null) {
    if (!gameState.isYourTurn) return;

    const action = gameState.currentAction;
    const status = gameState.currentState;

    if (context === 'setup') {
        toggleCardSelection(cardID, 'setup');
        document.getElementById('confirm-warriors-btn').disabled =
            gameState.selectedCards.length < 1 || gameState.selectedCards.length > 3;
        return;
    }

    // Handle attack phase card selection from hand
    if (status && status.current_action === 'attack' && context === 'player-hand') {
        handleAttackPhaseHandClick(cardID, card);
        return;
    }

    // Handle target selection for attack phase (enemy field)
    if (gameState.attackState.weaponId && context === 'enemy-field') {
        handleAttackPhaseTargetClick(cardID, 'enemy');
        return;
    }

    // Handle user warrior selection for special power (player field)
    if (gameState.attackState.weaponId && gameState.attackState.weaponType === 'specialpower' &&
        !gameState.attackState.userId && context === 'player-field') {
        handleAttackPhaseUserClick(cardID);
        return;
    }

    if (!action) return;

    gameState.selectedCards.push(cardID);
    highlightSelectedCard(cardID);

    // Handle different actions
    switch (action) {
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
    }
}

// Attack phase handlers
function handleAttackPhaseHandClick(cardID, card) {
    // Check if card can be used
    if (card && card.can_be_used === false) {
        return; // Card cannot be used in this phase
    }

    // Determine weapon type from card data
    const cardType = card ? (card.type || '').toLowerCase() : '';

    // Reset attack state
    resetAttackState();

    // Store weapon info
    gameState.attackState.weaponId = cardID;

    if (cardType === 'specialpower') {
        gameState.attackState.weaponType = 'specialpower';
        highlightSelectedCard(cardID);
        updateActionPrompt('Select a warrior from your field to use the special power');
        highlightValidUserWarriors(card);
    } else if (cardType === 'catapult') {
        gameState.attackState.weaponType = 'catapult';
        highlightSelectedCard(cardID);
        updateActionPrompt('Catapult selected - confirm to attack enemy castle');
        showConfirmButtons();
    } else if (cardType === 'weapon') {
        gameState.attackState.weaponType = 'weapon';
        highlightSelectedCard(cardID);
        updateActionPrompt('Select a target from enemy field');
        highlightValidTargets(card);
    }
}

function handleAttackPhaseUserClick(cardID) {
    // User selected a warrior to use the special power
    gameState.attackState.userId = cardID;
    highlightSelectedCard(cardID);

    updateActionPrompt('Now select a target from enemy field');

    // Get the weapon card to find valid targets
    const weapon = findCardById(gameState.attackState.weaponId);
    highlightValidTargets(weapon);
}

function handleAttackPhaseTargetClick(cardID, side) {
    // Check if this is a valid target
    const weapon = findCardById(gameState.attackState.weaponId);
    if (weapon && weapon.use_on && !weapon.use_on.includes(cardID)) {
        return; // Not a valid target
    }

    gameState.attackState.targetId = cardID;
    highlightSelectedCard(cardID);

    const weaponType = gameState.attackState.weaponType;
    if (weaponType === 'weapon') {
        updateActionPrompt('Attack ready - confirm to attack');
    } else if (weaponType === 'specialpower') {
        updateActionPrompt('Special power ready - confirm to use');
    }

    showConfirmButtons();
}

function highlightValidUserWarriors(weapon) {
    // For special powers, highlight warriors on player's field that can use the power
    const playerField = document.getElementById('player-field');
    playerField.querySelectorAll('.card').forEach(card => {
        const cardId = card.dataset.cardId;
        if (weapon && weapon.use_on && weapon.use_on.includes(cardId)) {
            card.classList.add('valid-target');
        }
    });
}

function highlightValidTargets(weapon) {
    // Highlight valid targets on enemy field
    const enemyField = document.getElementById('enemy-field');
    enemyField.querySelectorAll('.card').forEach(card => {
        const cardId = card.dataset.cardId;
        if (weapon && weapon.use_on && weapon.use_on.includes(cardId)) {
            card.classList.add('valid-target');
        }
    });
}

function findCardById(cardId) {
    const status = gameState.currentState;
    if (!status) return null;

    // Search in hand
    for (const card of status.current_player_hand || []) {
        if (card.id === cardId) return card;
    }

    // Search in player field
    for (const card of status.current_player_field || []) {
        if (card.id === cardId) return card;
    }

    return null;
}

function resetAttackState() {
    gameState.attackState = {
        weaponId: null,
        weaponType: null,
        userId: null,
        targetId: null
    };

    // Clear visual selections
    document.querySelectorAll('.card.selected, .card.valid-target').forEach(card => {
        card.classList.remove('selected', 'valid-target');
    });

    hideConfirmButtons();
}

function showConfirmButtons() {
    document.getElementById('confirm-action-buttons').classList.remove('hidden');
}

function hideConfirmButtons() {
    document.getElementById('confirm-action-buttons').classList.add('hidden');
}

function confirmAttackAction() {
    const attackState = gameState.attackState;

    if (!attackState.weaponId) return;

    if (attackState.weaponType === 'catapult') {
        // Catapult attack on castle
        sendAction('catapult', { card_position: 0 }); // Backend determines position
    } else if (attackState.weaponType === 'specialpower') {
        if (!attackState.userId || !attackState.targetId) return;
        sendAction('special_power', {
            weapon_id: attackState.weaponId,
            user_id: attackState.userId,
            target_id: attackState.targetId
        });
    } else if (attackState.weaponType === 'weapon') {
        if (!attackState.targetId) return;
        sendAction('attack', {
            weapon_id: attackState.weaponId,
            target_id: attackState.targetId
        });
    }

    resetAttackState();
    updateActionPrompt('');
}

function cancelAttackAction() {
    resetAttackState();
    updateActionPrompt('');

    // Re-render to clear highlights
    if (gameState.currentState) {
        renderGameBoard(gameState.currentState);
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

    // Store card data for attack phase logic
    div.dataset.cardType = card.type || '';
    div.dataset.cardSubType = card.sub_type || '';
    if (card.use_on) {
        div.dataset.useOn = JSON.stringify(card.use_on);
    }

    // Determine card type
    const cardType = getCardType(card);
    div.classList.add(cardType);

    // Apply card color from backend (with transparency for background)
    if (card.color) {
        const bgColor = hexToRgba(card.color, 0.3);
        div.style.setProperty('background', bgColor, 'important');
        div.style.setProperty('border-color', card.color, 'important');
    }

    // Check if card can be used during attack phase (only for hand cards)
    const status = gameState.currentState;
    if (context === 'player-hand' && status && status.current_action === 'attack') {
        if (card.can_be_used === false) {
            div.classList.add('unusable');
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

    // Disable all action buttons first
    document.querySelectorAll('.btn-action, #skip-phase-btn, #end-turn-btn').forEach(btn => {
        btn.disabled = true;
    });

    if (!isYourTurn || !status) return;

    // Move Warrior - enabled if can_move_warrior is true
    document.getElementById('move-warrior-btn').disabled = !status.can_move_warrior;

    // Trade - enabled if can_trade is true (from backend)
    document.getElementById('trade-btn').disabled = !status.can_trade;

    // Skip Phase and End Turn - always enabled during your turn
    document.getElementById('skip-phase-btn').disabled = false;
    document.getElementById('end-turn-btn').disabled = false;
}

function updatePhaseIndicator() {
    const phaseElement = document.getElementById('current-phase');
    const status = gameState.currentState;

    if (!status || !status.current_action) {
        phaseElement.textContent = '';
        return;
    }

    const phaseNames = {
        'draw': 'Draw Card',
        'attack': 'Attack Phase',
        'spy/steal': 'Spy/Steal Phase',
        'buy': 'Buy Phase',
        'construct': 'Construct Phase',
        'endturn': 'End Turn'
    };

    const phaseName = phaseNames[status.current_action] || status.current_action;
    phaseElement.textContent = phaseName;
    phaseElement.className = `phase-badge phase-${status.current_action.replace('/', '-')}`;
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
