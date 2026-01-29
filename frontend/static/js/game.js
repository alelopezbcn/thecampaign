// Game state
let ws = null;
let gameState = {
    playerName: '',
    gameID: '',
    isYourTurn: false,
    currentState: null,
    selectedCards: [],
    currentAction: null,
    pendingAction: null, // Track last action sent to detect results (trade, buy, etc.)
    pendingModalAction: null, // Track spy/steal to show correct modal title
    executedPhases: [], // Track phases that were actually executed this turn
    lastTurnPlayer: null, // Track whose turn it was to detect turn changes
    // Action state for multi-step actions
    actionState: {
        type: null,       // 'move_warrior', 'trade', 'attack', 'specialpower', 'catapult'
        weaponId: null,
        userId: null,     // For special power - the warrior using the power
        targetId: null,   // Target enemy warrior
        warriorId: null   // For move warrior
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
    document.getElementById('confirm-action-btn').addEventListener('click', confirmAction);
    document.getElementById('cancel-action-btn').addEventListener('click', cancelAction);

    // Game modal close button
    document.getElementById('modal-close-btn').addEventListener('click', hideGameModal);

    // Game over
    document.getElementById('new-game-btn').addEventListener('click', () => location.reload());

    // Game over modal
    document.getElementById('gameover-modal-btn').addEventListener('click', () => location.reload());
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
        case 'steal_result':
            handleStealResult(message.payload);
            break;
        case 'initial_warriors':
            handleInitialWarriors(message.payload);
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

function handleInitialWarriors(payload) {
    console.log('Initial warriors received:', payload);
    gameState.isYourTurn = payload.is_your_turn;
    gameState.selectedCards = [];

    // Store warriors for rendering
    gameState.initialWarriors = payload.warriors;

    // Show setup screen and render warriors
    showScreen('setup');
    renderInitialWarriors(payload.warriors);
    updateSetupTurnIndicator();
}

function renderInitialWarriors(warriors) {
    const container = document.getElementById('setup-hand');
    container.innerHTML = '';

    warriors.forEach(warrior => {
        const cardElement = createCardElement(warrior, 'setup');
        container.appendChild(cardElement);
    });
}

function handleGameState(payload) {
    console.log('Game state updated:', payload);
    console.log('New cards from payload:', payload.game_status.new_cards);

    // Detect when your turn starts (transition from not your turn to your turn)
    const wasYourTurn = gameState.isYourTurn;
    const isNowYourTurn = payload.is_your_turn;

    if (!wasYourTurn && isNowYourTurn) {
        // Your turn just started, reset executed phases
        gameState.executedPhases = ['draw']; // Draw is always automatic
    }

    gameState.isYourTurn = isNowYourTurn;
    gameState.currentState = payload.game_status;
    // Use the first card from new_cards array for highlighting
    gameState.newlyDrawnCards = payload.game_status.new_cards || [];
    console.log('gameState.newlyDrawnCards set to:', gameState.newlyDrawnCards);

    // Reset action state when game state updates
    gameState.currentAction = null;
    gameState.selectedCards = [];
    resetActionState();
    updateActionPrompt('');

    // Determine which screen to show
    if (payload.game_status.current_player && !isSetupComplete(payload.game_status)) {
        showSetupScreen(payload.game_status);
    } else {
        showGameScreen(payload.game_status);
    }

    updateTurnIndicator();
    updatePhaseIndicator();

    // Check if we have new cards from a pending action (trade or buy)
    const newCards = payload.game_status.new_cards || [];
    if (newCards.length > 0 && payload.is_your_turn && gameState.pendingAction) {
        // Find the full card data for the new cards
        const acquiredCards = payload.game_status.current_player_hand.filter(
            card => newCards.includes(card.id)
        );
        if (acquiredCards.length > 0) {
            if (gameState.pendingAction === 'buy') {
                showBoughtCardsModal(acquiredCards);
            } else if (gameState.pendingAction === 'trade') {
                showTradedCardsModal(acquiredCards);
            }
        }
        gameState.pendingAction = null; // Clear after handling
    }

    // Check if we have modal cards from spy/steal action
    const modalCards = payload.game_status.modal_cards || [];
    console.log('Modal cards check:', {
        modalCards: modalCards,
        modalCardsLength: modalCards.length,
        isYourTurn: payload.is_your_turn,
        pendingModalAction: gameState.pendingModalAction
    });
    if (modalCards.length > 0 && payload.is_your_turn && gameState.pendingModalAction) {
        if (gameState.pendingModalAction === 'spy_deck') {
            showCardsModal(modalCards, 'Top Cards from Deck', 'First card (left) is on top of the deck', true);
        } else if (gameState.pendingModalAction === 'spy_hand') {
            showCardsModal(modalCards, 'Enemy Hand', "These are the cards in your opponent's hand");
        } else if (gameState.pendingModalAction === 'steal') {
            showCardsModal(modalCards, 'Card Stolen!', 'You stole this card from your opponent');
        }
        gameState.pendingModalAction = null; // Clear after handling
    }

    // Check for game over message
    const gameOverMsg = payload.game_status.game_over_msg;
    if (gameOverMsg && gameOverMsg.length > 0) {
        const isWinner = gameOverMsg.toLowerCase().includes(gameState.playerName.toLowerCase());
        showGameOverModal(isWinner, gameOverMsg);
    }

    // Check for error message
    const errorMsg = payload.game_status.error_msg;
    if (errorMsg && errorMsg.length > 0) {
        showErrorToast(errorMsg);
    }
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
    const cards = payload.cards || [];
    const source = payload.source; // 1 = deck, 2 = enemy hand
    showSpyResultModal(cards, source);
}

function handleStealResult(payload) {
    const card = payload.card;
    if (card) {
        showStealResultModal(card);
    }
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

    // After confirming, it's now the opponent's turn to select
    gameState.isYourTurn = false;
    updateSetupTurnIndicator();
}

function sendAction(actionType, payload = null) {
    // Track actions that will return new cards
    if (actionType === 'trade' || actionType === 'buy') {
        gameState.pendingAction = actionType;
    }

    // Track executed phases based on action type
    const actionToPhase = {
        'draw_card': 'draw',
        'attack': 'attack',
        'special_power': 'attack',
        'spy': 'spy/steal',
        'steal': 'spy/steal',
        'catapult': 'spy/steal',
        'buy': 'buy',
        'construct': 'construct'
    };
    const phase = actionToPhase[actionType];
    if (phase && !gameState.executedPhases.includes(phase)) {
        gameState.executedPhases.push(phase);
    }

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
    // Remove valid-target highlights and multiplier badges
    document.querySelectorAll('.card.valid-target').forEach(card => {
        card.classList.remove('valid-target');
    });
    document.querySelectorAll('.dmg-multiplier-badge').forEach(badge => {
        badge.remove();
    });
}

function startAction(actionType) {
    resetActionState();
    gameState.currentAction = actionType;
    gameState.actionState.type = actionType;

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
    showConfirmButtons(); // Show cancel button

    // Re-render board to apply action-specific styles
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

    if (context === 'setup') {
        toggleCardSelection(cardID, 'setup');
        document.getElementById('confirm-warriors-btn').disabled =
            gameState.selectedCards.length < 1 || gameState.selectedCards.length > 3;
        return;
    }

    // Handle move_warrior action
    if (action === 'move_warrior' && context === 'player-hand') {
        if (cardType !== 'warrior') return; // Only warriors can be moved
        clearSelections();
        gameState.actionState.warriorId = cardID;
        highlightSelectedCard(cardID);
        updateActionPrompt('Warrior selected - confirm to move to field');
        return;
    }

    // Handle trade action
    if (action === 'trade' && context === 'player-hand') {
        toggleCardSelection(cardID, 'player-hand');
        if (gameState.selectedCards.length === 3) {
            updateActionPrompt('3 cards selected - confirm to trade');
        } else {
            updateActionPrompt(`Selected ${gameState.selectedCards.length}/3 cards for trade`);
        }
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

    // Handle target selection for attack phase (enemy field)
    if (gameState.actionState.weaponId && context === 'enemy-field') {
        handleAttackPhaseTargetClick(cardID, 'enemy');
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
        handleAttackPhaseTargetClick(cardID, 'player');
        return;
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
        highlightSelectedCard(cardID);
        const catapultDmg = card?.value || 0;
        updateActionPrompt(`🏰 Catapult (${catapultDmg} DMG) → Enemy Castle - Confirm to fire!`);
        showConfirmButtons();
    } else if (cardType === 'weapon') {
        gameState.actionState.type = 'attack';
        highlightSelectedCard(cardID);
        const weaponName = getCardName(card);
        const weaponDmg = card?.value || 0;
        updateActionPrompt(`⚔️ ${weaponName} (${weaponDmg} DMG) - Select a target`);
        highlightValidTargets(card);
        showConfirmButtons();
    }
}

function handleAttackPhaseUserClick(cardID) {
    // User selected a warrior to use the special power
    gameState.actionState.userId = cardID;
    highlightSelectedCard(cardID);

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

    highlightValidTargets(weapon);
}

function handleAttackPhaseTargetClick(cardID, side) {
    // Check if this is a valid target
    const weapon = findCardById(gameState.actionState.weaponId);
    console.log('Attack target click:', { cardID, side, weaponUseOn: weapon?.use_on, isValidTarget: weapon?.use_on?.includes(cardID) });

    if (weapon && weapon.use_on && !weapon.use_on.includes(cardID)) {
        console.log('Target rejected - not in use_on list');
        return; // Not a valid target
    }

    gameState.actionState.targetId = cardID;
    highlightSelectedCard(cardID);

    const actionType = gameState.actionState.type;
    const target = findCardById(cardID);

    if (actionType === 'attack') {
        const summary = buildAttackSummary(weapon, target);
        updateActionPrompt(summary);
    } else if (actionType === 'specialpower') {
        const user = findCardById(gameState.actionState.userId);
        const summary = buildSpecialPowerSummary(weapon, user, target);
        updateActionPrompt(summary);
    }
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
        // Show spy options modal
        showSpyOptionsModal();
    } else if (cardType === 'thief') {
        gameState.actionState.type = 'thief';
        // Show steal modal immediately when thief is selected
        showStealModal();
    }
}

// Buy phase handlers
function handleBuyPhaseHandClick(cardID, card) {
    if (card && card.can_be_used === false) {
        return;
    }

    clearSelections();
    gameState.actionState.weaponId = cardID;
    gameState.actionState.type = 'buy';
    highlightSelectedCard(cardID);
    updateActionPrompt('Resource selected - confirm to buy a card');
    showConfirmButtons();
}

// Construct phase handlers
function handleConstructPhaseHandClick(cardID, card) {
    if (card && card.can_be_used === false) {
        return;
    }

    // Select one resource card for construct
    clearSelections();
    gameState.selectedCards = [cardID];
    gameState.actionState.type = 'construct';
    highlightSelectedCard(cardID);
    updateActionPrompt('Resource selected - confirm to construct castle');
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
    const dmgMult = weapon?.dmg_mult || {};

    // Highlight valid targets on enemy field
    const enemyField = document.getElementById('enemy-field');
    enemyField.querySelectorAll('.card').forEach(card => {
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

    // Search in enemy field
    for (const card of status.enemy_field || []) {
        if (card.id === cardId) return card;
    }

    return null;
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

function resetActionState() {
    gameState.currentAction = null;
    gameState.selectedCards = [];
    gameState.actionState = {
        type: null,
        weaponId: null,
        userId: null,
        targetId: null,
        warriorId: null
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

function confirmAction() {
    const actionState = gameState.actionState;
    const action = gameState.currentAction;

    // Handle move warrior
    if (action === 'move_warrior' && actionState.warriorId) {
        sendAction('move_warrior', { warrior_id: actionState.warriorId });
        resetActionState();
        return;
    }

    // Handle trade
    if (action === 'trade' && gameState.selectedCards.length === 3) {
        sendAction('trade', { card_ids: gameState.selectedCards });
        resetActionState();
        return;
    }

    // Handle catapult
    if (actionState.type === 'catapult' && actionState.weaponId) {
        sendAction('catapult', { card_position: 0 });
        resetActionState();
        return;
    }

    // Handle special power
    if (actionState.type === 'specialpower') {
        if (!actionState.weaponId || !actionState.userId || !actionState.targetId) return;
        sendAction('special_power', {
            weapon_id: actionState.weaponId,
            user_id: actionState.userId,
            target_id: actionState.targetId
        });
        resetActionState();
        return;
    }

    // Handle regular attack
    if (actionState.type === 'attack') {
        if (!actionState.weaponId || !actionState.targetId) return;
        sendAction('attack', {
            weapon_id: actionState.weaponId,
            target_id: actionState.targetId
        });
        resetActionState();
        return;
    }

    // Handle buy
    if (actionState.type === 'buy' && actionState.weaponId) {
        sendAction('buy', { card_id: actionState.weaponId });
        resetActionState();
        return;
    }

    // Handle construct (send one card at a time)
    if (actionState.type === 'construct' && gameState.selectedCards.length > 0) {
        sendAction('construct', { card_id: gameState.selectedCards[0] });
        resetActionState();
        return;
    }
}

function cancelAction() {
    resetActionState();
    updateActionPrompt('');

    // Re-render to clear highlights and restore normal state
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

    // Render cemetery
    renderCemetery(status.cemetery);

    // Render discard pile
    renderDiscardPile(status.discard_pile);

    // Render deck
    renderDeck(status.cards_in_deck);

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

    // Check if card can be used based on current action (only when it's your turn)
    const status = gameState.currentState;
    const currentAction = gameState.currentAction;

    if (context === 'player-hand' && gameState.isYourTurn) {
        // During endturn phase, all cards are disabled
        if (status && status.current_action === 'endturn') {
            div.classList.add('unusable');
        }
        // During trade action, all cards are usable
        else if (currentAction === 'trade') {
            // All cards enabled for trade
        }
        // During move_warrior action, only warriors are usable
        else if (currentAction === 'move_warrior') {
            if (cardType !== 'warrior') {
                div.classList.add('unusable');
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
        }, 2000);
    }

    // Check if card is protected (field cards only)
    const isProtected = card.protected_by && card.protected_by.id;
    const shieldHtml = isProtected ? `
        <div class="card-shield">
            <span class="shield-icon">🛡️</span>
            <span class="shield-value">${card.protected_by.value || '?'}</span>
        </div>
    ` : '';

    // Create card HTML
    div.innerHTML = `
        ${shieldHtml}
        <div class="card-header">
            <span class="card-id">${div.dataset.cardId.substring(0, 6)}</span>
            <span class="card-type ${cardType}">${card.type || cardType}</span>
        </div>
        <div class="card-content">
            <div class="card-name">${getCardName(card)}</div>
            ${getCardStats(card, cardType)}
        </div>
    `;

    // Add protected class for styling
    if (isProtected) {
        div.classList.add('protected');
    }

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
    const castleValue = castle.value || 0;

    container.className = 'castle';
    if (isConstructed) {
        container.classList.add('constructed');
    }

    container.innerHTML = `
        <div class="castle-status">${isConstructed ? 'Constructed' : 'Not Constructed'}</div>
        ${isConstructed ? `
            <div class="castle-info">
                <div class="castle-value">Value: ${castleValue}</div>
                <div class="castle-resources">${resourceCount} resource cards</div>
            </div>
        ` : ''}
    `;
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
        const card = cemetery.last_corp;
        const cardColor = card.color || '#666';
        lastCorpContainer.innerHTML = `
            <div class="cemetery-card" style="background: linear-gradient(135deg, ${cardColor}40, ${cardColor}20); border-color: ${cardColor};">
                <div class="card-type" style="color: ${cardColor};">${card.sub_type || card.type}</div>
            </div>
        `;
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
    // Update the phase tracker turn status
    updatePhaseTracker();
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

    // In endturn phase, only End Turn button is enabled
    if (status.current_action === 'endturn') {
        document.getElementById('end-turn-btn').disabled = false;
        return;
    }

    // Move Warrior - enabled if can_move_warrior is true
    document.getElementById('move-warrior-btn').disabled = !status.can_move_warrior;

    // Trade - enabled if can_trade is true (from backend)
    document.getElementById('trade-btn').disabled = !status.can_trade;

    // Skip Phase and End Turn - always enabled during your turn
    document.getElementById('skip-phase-btn').disabled = false;
    document.getElementById('end-turn-btn').disabled = false;
}

function updatePhaseTracker() {
    const status = gameState.currentState;
    const turnStatusElement = document.getElementById('phase-turn-status');
    const turnTextElement = turnStatusElement?.querySelector('.turn-text');

    // Update turn status
    if (turnStatusElement && turnTextElement) {
        turnStatusElement.classList.remove('your-turn', 'enemy-turn');
        if (gameState.isYourTurn) {
            turnStatusElement.classList.add('your-turn');
            turnTextElement.textContent = 'Your Turn';
        } else {
            turnStatusElement.classList.add('enemy-turn');
            turnTextElement.textContent = 'Enemy Turn';
        }
    }

    // Phase order
    const phaseOrder = ['draw', 'attack', 'spy/steal', 'buy', 'construct', 'endturn'];
    const currentAction = status?.current_action || '';

    // Find the index of the current action
    let currentIndex = phaseOrder.indexOf(currentAction);

    // Update each phase item
    phaseOrder.forEach((phase, index) => {
        const phaseItem = document.querySelector(`.phase-item[data-phase="${phase}"]`);
        if (!phaseItem) return;

        phaseItem.classList.remove('completed', 'current', 'skipped');

        if (currentIndex === -1) {
            // No current action, reset all
            return;
        }

        if (index < currentIndex) {
            // If it's enemy's turn, show all past phases as green
            // If it's your turn, check if phase was actually executed
            if (!gameState.isYourTurn || gameState.executedPhases.includes(phase)) {
                phaseItem.classList.add('completed');
            } else {
                phaseItem.classList.add('skipped');
            }
        } else if (index === currentIndex) {
            // This is the current phase
            phaseItem.classList.add('current');
        }
        // Phases after current index remain in default (pending) state
    });
}

// Keep old function name for compatibility
function updatePhaseIndicator() {
    updatePhaseTracker();
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
    resetActionState();
    updateActionPrompt('');
}

// Steal Modal
function showStealModal() {
    const enemyHandCount = gameState.currentState?.cards_in_enemy_hand || 0;

    let content = '';
    for (let i = 0; i < enemyHandCount; i++) {
        content += `
            <div class="card-face-down" data-position="${i}" onclick="selectStealPosition(${i})">
                <span class="card-position">#${i + 1}</span>
            </div>
        `;
    }

    showGameModal('Select a card to steal', "Choose one of your opponent's cards", content, true);
}

function selectStealPosition(position) {
    gameState.pendingModalAction = 'steal';
    sendAction('steal', { card_position: position });
    hideGameModal();
}

// Spy Options Modal
function showSpyOptionsModal() {
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
    hideGameModal();
    gameState.pendingModalAction = option === 1 ? 'spy_deck' : 'spy_hand';
    sendAction('spy', { option: option });
}

// Spy Result Modal
function showSpyResultModal(cards, source) {
    if (source === 1) {
        // Deck cards - show with position indicator
        showCardsModal(cards, 'Top Cards from Deck', 'First card (left) is on top of the deck', true);
    } else {
        // Enemy hand
        showCardsModal(cards, 'Enemy Hand', 'These are the cards in your opponent\'s hand');
    }
}

// Steal Result Modal
function showStealResultModal(card) {
    showCardsModal([card], 'Card Stolen!', 'You stole this card from your opponent');
}

// Bought Cards Modal
function showBoughtCardsModal(cards) {
    showCardsModal(cards, 'Cards Bought!', `You bought ${cards.length} card${cards.length > 1 ? 's' : ''}`);
}

// Traded Cards Modal
function showTradedCardsModal(cards) {
    showCardsModal(cards, 'Card Received!', 'You traded 3 cards for this');
}

// Normalize card from gamestatus.Card format to UI format
function normalizeCard(card) {
    // If card has card_type object (new gamestatus.Card format), normalize it
    if (card.card_type) {
        return {
            id: card.card_id,
            type: card.card_type.name,
            sub_type: card.card_type.sub_name,
            color: card.card_type.color,
            value: card.value
        };
    }
    // Already in UI format
    return card;
}

// Generic Cards Modal
function showCardsModal(cards, title, subtitle, showPositionIndicators = false) {
    let content = '';

    if (cards.length === 0) {
        content = '<p style="color: #b0b0b0;">No cards to show</p>';
    } else {
        cards.forEach((rawCard, index) => {
            const card = normalizeCard(rawCard);
            const cardType = getCardType(card);
            const cardName = getCardName(card);
            const bgColor = card.color ? hexToRgba(card.color, 0.3) : '';
            const borderColor = card.color || '';

            // Position indicator for deck cards
            let positionBadge = '';
            if (showPositionIndicators) {
                const positionLabel = index === 0 ? 'TOP' : `#${index + 1}`;
                const badgeStyle = index === 0
                    ? 'background: #4CAF50; color: white;'
                    : 'background: #666; color: #ccc;';
                positionBadge = `<span class="position-badge" style="${badgeStyle} padding: 2px 6px; border-radius: 4px; font-size: 10px; font-weight: bold; margin-left: 5px;">${positionLabel}</span>`;
            }

            content += `
                <div class="card ${cardType}" style="${bgColor ? `background: ${bgColor};` : ''} ${borderColor ? `border-color: ${borderColor};` : ''}">
                    <div class="card-header">
                        <span class="card-type ${cardType}">${card.type || cardType}</span>
                        ${positionBadge}
                    </div>
                    <div class="card-content">
                        <div class="card-name">${cardName}</div>
                        ${getCardStats(card, cardType)}
                    </div>
                </div>
            `;
        });
    }

    showGameModal(title, subtitle, content, true);
}

// Game Over Modal
function showGameOverModal(isWinner, message) {
    const modal = document.getElementById('gameover-modal');
    const iconElement = document.getElementById('gameover-modal-icon');
    const titleElement = document.getElementById('gameover-modal-title');
    const messageElement = document.getElementById('gameover-modal-message');

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
