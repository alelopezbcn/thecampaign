// Game state
let ws = null;
let reconnectAttempts = 0;
let reconnectTimer = null;
let timerInterval = null;
let pendingAnimationsCallback = null; // Deferred animations waiting for modal close
let endTurnCountdownTimer = null;
const END_TURN_COUNTDOWN_SECS = 3;
const MAX_RECONNECT_ATTEMPTS = 20;
let gameState = {
    playerName: '',
    gameID: '',
    gameMode: '1v1',
    isYourTurn: false,
    currentState: null,
    selectedCards: [],
    currentAction: null,
    pendingAction: null, // Track last action sent to detect results (trade, buy, etc.)
    pendingModalAction: null, // Track spy/steal to show correct modal title
    executedPhases: [], // Track phases that were actually executed this turn
    lastTurnPlayer: null, // Track whose turn it was to detect turn changes
    historyMessages: [], // Accumulated history messages
    waitingPlayers: [], // Track players who have joined the waiting room
    maxPlayers: 2, // Max players for current game mode
    teamAssignments: {}, // playerName -> teamNumber (1 or 2), 2v2 only
    isCreator: false, // Whether this player created the room
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
    game: document.getElementById('game-screen'),
    gameover: document.getElementById('gameover-screen')
};

// Initialize
document.addEventListener('DOMContentLoaded', () => {
    setupEventListeners();
    checkUrlParams();
    fetch('/api/version')
        .then(r => r.json())
        .then(data => {
            const el = document.getElementById('game-version');
            if (el) el.textContent = data.version;
        })
        .catch(() => {});
    fetch('/api/card-config')
        .then(r => r.json())
        .then(data => { cardConfig = data; })
        .catch(() => {});
});

function checkUrlParams() {
    const params = new URLSearchParams(window.location.search);
    const gameID = params.get('game');
    if (gameID) {
        gameState.gameID = gameID.toUpperCase();
        // Show the URL-join section, hide others
        document.getElementById('create-game-section').classList.add('hidden');
        document.getElementById('join-url-section').classList.remove('hidden');
        document.getElementById('join-game-code').textContent = gameState.gameID;
    }
}

function setupEventListeners() {
    // Create game
    document.getElementById('create-btn').addEventListener('click', createGame);
    document.getElementById('player-name').addEventListener('keypress', (e) => {
        if (e.key === 'Enter') createGame();
    });

    // Join by code
    document.getElementById('join-code-btn').addEventListener('click', joinGameByCode);
    document.getElementById('game-id').addEventListener('keypress', (e) => {
        if (e.key === 'Enter') joinGameByCode();
    });
    document.getElementById('join-player-name').addEventListener('keypress', (e) => {
        if (e.key === 'Enter') joinGameByCode();
    });

    // Join by URL
    document.getElementById('join-url-btn').addEventListener('click', joinGameByUrl);
    document.getElementById('url-player-name').addEventListener('keypress', (e) => {
        if (e.key === 'Enter') joinGameByUrl();
    });

    // Toggle between create/join sections
    document.getElementById('show-join-link').addEventListener('click', (e) => {
        e.preventDefault();
        document.getElementById('create-game-section').classList.add('hidden');
        document.getElementById('join-code-section').classList.remove('hidden');
    });
    document.getElementById('show-create-link').addEventListener('click', (e) => {
        e.preventDefault();
        document.getElementById('join-code-section').classList.add('hidden');
        document.getElementById('create-game-section').classList.remove('hidden');
    });

    // Copy link button
    document.getElementById('copy-link-btn').addEventListener('click', () => {
        const shareUrl = document.getElementById('share-url');
        const btn = document.getElementById('copy-link-btn');
        const text = shareUrl.value;
        if (navigator.clipboard && window.isSecureContext) {
            navigator.clipboard.writeText(text);
        } else {
            shareUrl.select();
            document.execCommand('copy');
        }
        btn.textContent = 'Copied!';
        setTimeout(() => btn.textContent = 'Copy Link', 2000);
    });

    // Game mode selector
    document.querySelectorAll('.game-mode-option').forEach(option => {
        option.addEventListener('click', () => {
            document.querySelectorAll('.game-mode-option').forEach(o => o.classList.remove('selected'));
            option.classList.add('selected');
            gameState.gameMode = option.dataset.mode;
        });
    });

    // Game screen actions - only 4 buttons
    document.getElementById('move-warrior-btn').addEventListener('click', () => startAction('move_warrior'));
    document.getElementById('trade-btn').addEventListener('click', () => startAction('trade'));
    document.getElementById('skip-phase-btn').addEventListener('click', handleSkipPhase);
    document.getElementById('end-turn-btn').addEventListener('click', () => sendAction('end_turn'));
    document.getElementById('endturn-popup-btn').addEventListener('click', () => {
        clearEndTurnCountdown();
        sendAction('end_turn');
    });

    // Cancel action button
    document.getElementById('cancel-action-btn').addEventListener('click', cancelAction);

    // Game modal close button
    document.getElementById('modal-close-btn').addEventListener('click', hideGameModal);

    // Action confirm modal buttons
    document.getElementById('action-confirm-yes').addEventListener('click', onActionConfirmYes);
    document.getElementById('action-confirm-no').addEventListener('click', onActionConfirmNo);

    // Start game button
    document.getElementById('start-game-btn').addEventListener('click', () => {
        sendAction('start_game');
        document.getElementById('start-game-btn').disabled = true;
        document.getElementById('start-game-btn').textContent = 'Starting...';
    });

    // Turn transition modal close
    document.getElementById('turn-transition-close').addEventListener('click', hideTurnTransitionModal);

    // Stolen card modal close
    document.getElementById('stolen-card-close').addEventListener('click', hideStolenCardModal);

    // Spy notification modal close
    document.getElementById('spy-notification-close').addEventListener('click', hideSpyNotificationModal);

    // Desertion notification modal close
    document.getElementById('desertion-notification-close').addEventListener('click', hideDesertionNotificationModal);

    // Ambush triggered modal close
    document.getElementById('ambush-triggered-close').addEventListener('click', hideAmbushTriggeredModal);
    document.getElementById('ambush-triggered-modal').addEventListener('click', (e) => {
        if (e.target === e.currentTarget) hideAmbushTriggeredModal();
    });

    // Game over
    document.getElementById('new-game-btn').addEventListener('click', () => location.reload());

    // Game over modal
    document.getElementById('gameover-modal-btn').addEventListener('click', () => sendMessage('restart_game'));

    // Global keyboard shortcuts
    document.addEventListener('keydown', handleGlobalKeyboard);

    // Close modals when clicking outside content
    const modalOverlays = [
        { id: 'game-modal', hide: hideGameModal },
        { id: 'action-confirm-modal', hide: onActionConfirmNo },
        { id: 'stolen-card-modal', hide: hideStolenCardModal },
        { id: 'spy-notification-modal', hide: hideSpyNotificationModal },
        { id: 'desertion-notification-modal', hide: hideDesertionNotificationModal },
        { id: 'gameover-modal', hide: () => location.reload() },
    ];
    modalOverlays.forEach(({ id, hide }) => {
        document.getElementById(id).addEventListener('click', (e) => {
            if (e.target === e.currentTarget) hide();
        });
    });
}

function handleGlobalKeyboard(e) {
    // Don't intercept when typing in input fields
    if (e.target.tagName === 'INPUT' || e.target.tagName === 'TEXTAREA') return;

    const actionConfirmModal = document.getElementById('action-confirm-modal');
    const gameModal = document.getElementById('game-modal');
    const endturnPopup = document.getElementById('endturn-popup');
    const actionPrompt = document.getElementById('action-prompt-container');

    const isActionConfirmOpen = actionConfirmModal && !actionConfirmModal.classList.contains('hidden');
    const isGameModalOpen = gameModal && !gameModal.classList.contains('hidden');
    const isEndturnPopupOpen = endturnPopup && !endturnPopup.classList.contains('hidden');
    const isActionPromptOpen = actionPrompt && !actionPrompt.classList.contains('hidden');

    const turnTransitionModal = document.getElementById('turn-transition-modal');
    const stolenCardModal = document.getElementById('stolen-card-modal');
    const spyNotificationModal = document.getElementById('spy-notification-modal');
    const isTurnTransitionOpen = turnTransitionModal && !turnTransitionModal.classList.contains('hidden');
    const isStolenCardOpen = stolenCardModal && !stolenCardModal.classList.contains('hidden');
    const isSpyNotificationOpen = spyNotificationModal && !spyNotificationModal.classList.contains('hidden');

    if (e.key === 'Escape') {
        if (isTurnTransitionOpen) {
            hideTurnTransitionModal();
        } else if (isStolenCardOpen) {
            hideStolenCardModal();
        } else if (isSpyNotificationOpen) {
            hideSpyNotificationModal();
        } else if (isActionConfirmOpen) {
            onActionConfirmNo();
        } else if (isGameModalOpen) {
            hideGameModal();
        } else if (isActionPromptOpen) {
            cancelAction();
        }
    } else if (e.key === 'Enter') {
        if (isActionConfirmOpen) {
            onActionConfirmYes();
        } else if (isGameModalOpen) {
            hideGameModal();
        } else if (isEndturnPopupOpen) {
            clearEndTurnCountdown();
            sendAction('end_turn');
        }
    }
}

function handleSkipPhase() {
    const status = gameState.currentState;
    // If we're in the last phase (endturn), end the turn instead
    if (status && status.current_action === 'endturn') {
        clearEndTurnCountdown();
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
        reconnectAttempts = 0;
        if (reconnectTimer) {
            clearTimeout(reconnectTimer);
            reconnectTimer = null;
        }
        showStatus('connection-status', 'Connected to server', 'success');

        // Send join_game for both create (empty gameID) and join/reconnect flows
        if (gameState.playerName) {
            sendMessage('join_game', {
                player_name: gameState.playerName,
                game_id: gameState.gameID,
                game_mode: gameState.gameMode
            });
        }
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
        ws = null;

        // Clear any pending reconnect timer
        if (reconnectTimer) {
            clearTimeout(reconnectTimer);
            reconnectTimer = null;
        }

        // Only auto-reconnect if we were in a game and under the retry limit
        if (gameState.playerName && gameState.gameID && reconnectAttempts < MAX_RECONNECT_ATTEMPTS) {
            const delay = Math.min(1000 * Math.pow(2, reconnectAttempts), 10000);
            reconnectAttempts++;
            showStatus('connection-status', `Reconnecting in ${delay / 1000}s... (attempt ${reconnectAttempts})`, 'error');
            reconnectTimer = setTimeout(() => {
                console.log(`Reconnect attempt ${reconnectAttempts}`);
                connectWebSocket();
            }, delay);
        } else if (reconnectAttempts >= MAX_RECONNECT_ATTEMPTS) {
            showStatus('connection-status', 'Could not reconnect. Please refresh the page.', 'error');
        } else {
            showStatus('connection-status', 'Disconnected from server', 'error');
        }
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
        default:
            console.log('Unknown message type:', message.type);
    }
}

function handleError(payload) {
    console.error('Server error:', payload.message);
    showStatus('connection-status', payload.message, 'error');
    addErrorToHistory(payload.message);
}

function handlePlayerJoined(payload) {
    console.log('Player joined:', payload.player_name);

    // Capture server-generated game ID (important for create flow)
    if (payload.game_id) {
        gameState.gameID = payload.game_id;
        // Update URL without reload so the shareable link works
        const newUrl = `${window.location.pathname}?game=${payload.game_id}`;
        window.history.replaceState({}, '', newUrl);
    }

    if (payload.game_mode) {
        gameState.gameMode = payload.game_mode;
    }
    if (payload.max_players) {
        gameState.maxPlayers = payload.max_players;
    }
    // Use the full players list from the server (authoritative)
    if (payload.players && payload.players.length > 0) {
        gameState.waitingPlayers = payload.players;
    } else if (payload.player_name && !gameState.waitingPlayers.includes(payload.player_name)) {
        gameState.waitingPlayers.push(payload.player_name);
    }

    if (payload.teams) {
        gameState.teamAssignments = payload.teams;
    }

    showWaitingScreen();
    updateWaitingScreen();
}

function handleGameStarted(payload) {
    console.log('Game started:', payload);
    gameState.playerName = payload.your_name;
    gameState.gameID = payload.game_id;
    gameState.currentState = null;
    gameState.isYourTurn = false;
    gameState.executedPhases = [];

    document.getElementById('gameover-modal').classList.add('hidden');
    document.getElementById('current-game-id').textContent = payload.game_id;
}

function handleGameState(payload) {
    console.log('Game state updated:', payload);
    console.log('New cards from payload:', payload.game_status.new_cards);

    // Save previous state for damage detection
    const previousState = gameState.currentState;

    // Detect when your turn starts (transition from not your turn to your turn)
    const wasYourTurn = gameState.isYourTurn;
    const isNowYourTurn = payload.is_your_turn;

    if (!wasYourTurn && isNowYourTurn) {
        // Your turn just started, reset executed phases
        gameState.executedPhases = ['draw']; // Draw is always automatic
    }

    gameState.isYourTurn = isNowYourTurn;
    gameState.currentState = payload.game_status;

    if (payload.game_status.game_mode) {
        gameState.gameMode = payload.game_status.game_mode;
    }

    // Detect newly eliminated players
    if (previousState) {
        checkForEliminations(previousState, payload.game_status);
    }

    // Detect killed warriors and clone their DOM elements before re-render
    let killedWarriors = [];
    if (previousState) {
        killedWarriors = prepareDeathAnimations(previousState, payload.game_status);
    }

    // Detect hand cards that will vanish before re-render
    let vanishedCards = [];
    if (previousState) {
        vanishedCards = prepareHandCardVanish(previousState, payload.game_status);
    }

    // Detect steal and clone opponent card-back before re-render
    let stealData = null;
    if (previousState) {
        stealData = prepareStealAnimation(previousState, payload.game_status);
    }

    // Detect sabotage and clone opponent card-back before re-render
    let sabotageData = null;
    if (previousState) {
        sabotageData = prepareSabotageAnimation(previousState, payload.game_status);
    }

    // Detect desertion (warrior flying from enemy field to own field)
    let desertionData = null;
    if (previousState) {
        desertionData = prepareDesertionAnimation(previousState, payload.game_status);
    }

    // Detect warrior move for animation (before re-render)
    let warriorMoveData = null;
    if (previousState) {
        warriorMoveData = prepareWarriorMoveAnimation(previousState, payload.game_status);
    }

    // Detect weapon attack for animation (before re-render)
    let attackAnimData = null;
    if (previousState) {
        attackAnimData = prepareAttackAnimation(previousState, payload.game_status);
    }

    // Detect blood rain for animation (before re-render)
    let bloodRainAnimData = null;
    if (previousState) {
        bloodRainAnimData = prepareBloodRainAnimation(previousState, payload.game_status);
    }

    // Detect deck draw for animation
    let deckDrawInfo = null;
    if (previousState) {
        deckDrawInfo = detectDeckDraw(previousState, payload.game_status);
    }

    // Determine if a result modal will be shown (check before resetActionState)
    const _newCards = payload.game_status.new_cards || [];
    const _modalCards = payload.game_status.modal_cards || [];
    const willShowBuyTradeModal = (
        _newCards.length > 0 && payload.is_your_turn && gameState.pendingAction &&
        ['buy', 'trade'].includes(gameState.pendingAction)
    );
    const willShowSpyStealModal = (
        _modalCards.length > 0 && payload.is_your_turn && gameState.pendingModalAction
    );

    // Clear any stale deferred animations from a previous state update
    pendingAnimationsCallback = null;

    // Use the first card from new_cards array for highlighting
    gameState.newlyDrawnCards = _newCards;
    console.log('gameState.newlyDrawnCards set to:', gameState.newlyDrawnCards);

    // Reset action state when game state updates
    gameState.currentAction = null;
    gameState.selectedCards = [];
    resetActionState();
    updateActionPrompt('');

    showGameScreen(payload.game_status);

    // Bundle all post-render animations
    const playAllAnimations = () => {
        if (previousState) {
            setTimeout(() => showDamageFeedback(previousState, payload.game_status), 50);
        }
        if (killedWarriors.length > 0) {
            playDeathAnimations(killedWarriors);
        }
        if (vanishedCards.length > 0) {
            playCardVanishAnimations(vanishedCards);
        }
        if (previousState) {
            const newlyProtected = detectNewProtections(previousState, payload.game_status);
            if (newlyProtected.length > 0) {
                setTimeout(() => showProtectionAnimations(newlyProtected), 50);
            }
            const brokenShields = detectBrokenShields(previousState, payload.game_status);
            if (brokenShields.length > 0) {
                setTimeout(() => showShieldBreakAnimations(brokenShields), 50);
            }
        }
        if (previousState) {
            setTimeout(() => {
                const changes = detectCastleChanges(previousState, payload.game_status);
                changes.constructions.forEach(c => showCastleConstructionAnimation(c));
                changes.goldAdded.forEach(c => showCastleGoldAnimation(c));
                changes.goldRemoved.forEach(c => showCastleAttackAnimation(c));
                // Fortress destroyed by catapult
                if (payload.game_status.last_action === 'catapult_blocked') {
                    changes.fortressDestroyed.forEach(c => showFortressDestroyedAnimation(c));
                }
            }, 50);
        }
        if (warriorMoveData) {
            playWarriorMoveAnimation(warriorMoveData, payload.game_status);
        }
        if (attackAnimData) {
            playAttackAnimation(attackAnimData, payload.game_status);
        }
        if (bloodRainAnimData) {
            playBloodRainAnimation(bloodRainAnimData, payload.game_status);
        }
        if (stealData) {
            playStealAnimation(stealData);
        }
        if (sabotageData) {
            playSabotageAnimation(sabotageData);
        }
        if (desertionData) {
            playDesertionAnimation(desertionData);
        }
        if (deckDrawInfo) {
            setTimeout(() => playDeckDrawAnimation(deckDrawInfo), 50);
        }
        if (previousState) {
            setTimeout(() => {
                const pileChanges = detectPileAndCemeteryChanges(previousState, payload.game_status);
                if (pileChanges.pileAdded) showPileAnimation();
                if (pileChanges.cemeteryAdded) showCemeteryAnimation();
                if (payload.game_status.last_action === 'resurrection') showResurrectionAnimation();
                if (payload.game_status.last_action === 'buy_mercenary') showMercenaryHiredAnimation();
                if (payload.game_status.last_action === 'place_ambush') showAmbushPlacedAnimation(payload.game_status);
            }, 50);
        }
    };

    // For buy/trade: defer animations until modal close (show modal first)
    // For spy/steal: play animations first, then show result modal after animations finish
    if (willShowBuyTradeModal) {
        pendingAnimationsCallback = playAllAnimations;
    } else {
        playAllAnimations();
    }

    updateTurnIndicator();
    updatePhaseIndicator();
    updatePlayerListPanel();
    startTimers(payload.game_status);

    // Spectator: show who's next when the active player enters the endturn phase
    const prevAction = previousState?.current_action;
    const nowAction = payload.game_status.current_action;
    if (!isNowYourTurn && nowAction === 'endturn' && prevAction !== 'endturn') {
        const nextPlayer = payload.game_status.next_turn_player;
        if (nextPlayer) {
            const isNextYou = nextPlayer === gameState.playerName;
            const label = isNextYou ? 'Your Turn Next!' : `Next: ${nextPlayer}`;
            showTurnTransitionModal(nextPlayer, END_TURN_COUNTDOWN_SECS * 1000, label);
        }
    }

    // Show turn transition modal when end_turn is actually processed
    if (payload.game_status.last_action === 'end_turn') {
        const turnPlayer = payload.game_status.turn_player;
        const label = isNowYourTurn ? 'Your Turn!' : `${turnPlayer}'s Turn`;
        showTurnTransitionModal(turnPlayer, END_TURN_COUNTDOWN_SECS * 1000, label);
    }

    // Detect if a card was stolen from us
    const stolenCards = payload.game_status.stolen_from_you_card;
    if (stolenCards && stolenCards.length > 0) {
        showStolenCardModal(stolenCards[0]);
    }

    // Detect if a card was sabotaged from us
    const sabotagedCards = payload.game_status.sabotaged_from_you_card;
    if (sabotagedCards && sabotagedCards.length > 0) {
        showStolenCardModal(sabotagedCards[0]);
    }

    // Detect spy notification
    const spyNotification = payload.game_status.spy_notification;
    if (spyNotification) {
        showSpyNotificationModal(spyNotification);
    }

    // Detect ambush triggered
    const ambushTriggered = payload.game_status.ambush_triggered;
    if (ambushTriggered) {
        showAmbushTriggeredModal(ambushTriggered.effect_display);
    }

    // Detect desertion notification (victim only)
    const desertionNotification = payload.game_status.desertion_notification;
    if (desertionNotification) {
        showDesertionNotificationModal(desertionNotification);
    }

    // Check if we have new cards from a pending action (trade or buy)
    const newCards = _newCards;
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
    // Show after animations finish so player sees the steal/spy animation first
    const modalCards = _modalCards;
    if (modalCards.length > 0 && payload.is_your_turn && gameState.pendingModalAction) {
        const action = gameState.pendingModalAction;
        gameState.pendingModalAction = null;
        const animDelay = 1400; // Wait for animations to finish
        setTimeout(() => {
            if (action === 'spy_deck') {
                showCardsModal(modalCards, 'Top Cards from Deck', 'First card (left) is on top of the deck', true);
            } else if (action === 'spy_hand') {
                showCardsModal(modalCards, 'Enemy Hand', "These are the cards in your opponent's hand");
            } else if (action === 'steal') {
                showCardsModal(modalCards, 'Card Stolen!', 'You stole this card from your opponent');
            } else if (action === 'sabotage') {
                showCardsModal(modalCards, 'Card Destroyed!', "You destroyed this card from your opponent's hand");
            }
        }, animDelay);
    }

    // Check for game over message
    const gameOverMsg = payload.game_status.game_over_msg;
    if (gameOverMsg && gameOverMsg.length > 0) {
        const isWinner = checkIsWinner(gameOverMsg, payload.game_status);
        showGameOverModal(isWinner, gameOverMsg);
    }

    // Check for error message
    const errorMsg = payload.game_status.error_msg;
    if (errorMsg && errorMsg.length > 0) {
        showErrorToast(errorMsg);
    }
}

function checkIsWinner(gameOverMsg, status) {
    return !!status.is_winner;
}

function handleGameEnded() {
    const gameOverMsg = gameState.currentState?.game_over_msg || 'Game Over!';
    const isWinner = checkIsWinner(gameOverMsg, gameState.currentState || {});
    showGameOverModal(isWinner, gameOverMsg);
}

// Screen management
function showScreen(screenName) {
    Object.values(screens).forEach(screen => screen.classList.add('hidden'));
    screens[screenName].classList.remove('hidden');
}

function showWaitingScreen() {
    document.getElementById('current-game-id').textContent = gameState.gameID;

    // Populate shareable URL
    const shareUrl = `${window.location.origin}${window.location.pathname}?game=${gameState.gameID}`;
    const shareInput = document.getElementById('share-url');
    if (shareInput) {
        shareInput.value = shareUrl;
    }

    // Ensure current player is in the list
    if (gameState.playerName && !gameState.waitingPlayers.includes(gameState.playerName)) {
        gameState.waitingPlayers.push(gameState.playerName);
    }

    updateWaitingScreen();
    showScreen('waiting');
}

function updateWaitingScreen() {
    const modeBadge = document.getElementById('waiting-mode-badge');
    const countEl = document.getElementById('waiting-player-count');
    const listEl = document.getElementById('waiting-players-list');

    if (modeBadge) {
        modeBadge.textContent = gameState.gameMode.toUpperCase();
    }

    if (countEl) {
        countEl.textContent = `${gameState.waitingPlayers.length}/${gameState.maxPlayers} players`;
    }

    // Show start button only for room creator, enable when all joined
    const startBtn = document.getElementById('start-game-btn');
    if (startBtn) {
        if (!gameState.isCreator) {
            startBtn.style.display = 'none';
        } else {
            startBtn.style.display = '';
            const allJoined = gameState.waitingPlayers.length >= gameState.maxPlayers;
            startBtn.disabled = !allJoined;
            if (allJoined) {
                startBtn.textContent = 'Start Game';
            }
        }
    }

    if (!listEl) return;
    listEl.innerHTML = '';

    // 2v2 mode: show two team columns
    if (gameState.gameMode === '2v2') {
        listEl.classList.add('team-layout');
        listEl.innerHTML = renderTeamWaitingLayout();
        return;
    }

    // Non-2v2: existing flat layout
    listEl.classList.remove('team-layout');

    for (const name of gameState.waitingPlayers) {
        const isSelf = name === gameState.playerName;
        const slot = document.createElement('div');
        slot.className = `player-slot ${isSelf ? 'self' : 'filled'}`;
        slot.innerHTML = `
            <div class="player-slot-icon">${isSelf ? '⚔' : '🛡'}</div>
            <span class="player-slot-name">${name}</span>
            ${isSelf ? '<span class="player-slot-you">YOU</span>' : ''}
        `;
        listEl.appendChild(slot);
    }

    const emptySlots = gameState.maxPlayers - gameState.waitingPlayers.length;
    for (let i = 0; i < emptySlots; i++) {
        const slot = document.createElement('div');
        slot.className = 'player-slot empty';
        slot.innerHTML = `
            <div class="player-slot-icon">?</div>
            <span class="player-slot-name">Waiting...</span>
        `;
        listEl.appendChild(slot);
    }
}

function renderTeamWaitingLayout() {
    const teams = gameState.teamAssignments;
    const team1Players = gameState.waitingPlayers.filter(n => teams[n] === 1);
    const team2Players = gameState.waitingPlayers.filter(n => teams[n] === 2);

    function renderTeamColumn(teamNum, players) {
        let html = `<div class="team-column">`;
        html += `<div class="team-header">Team ${teamNum}</div>`;

        for (const name of players) {
            const isSelf = name === gameState.playerName;
            html += `
                <div class="player-slot ${isSelf ? 'self' : 'filled'}">
                    <div class="player-slot-icon">${isSelf ? '\u2694' : '\uD83D\uDEE1'}</div>
                    <span class="player-slot-name">${name}</span>
                    ${isSelf ? '<span class="player-slot-you">YOU</span>' : ''}
                </div>
            `;
        }

        for (let i = players.length; i < 2; i++) {
            html += `
                <div class="player-slot empty">
                    <div class="player-slot-icon">?</div>
                    <span class="player-slot-name">Waiting...</span>
                </div>
            `;
        }

        html += `</div>`;
        return html;
    }

    let layout = `<div class="teams-container">`;
    layout += renderTeamColumn(1, team1Players);
    layout += `<div class="team-vs">VS</div>`;
    layout += renderTeamColumn(2, team2Players);
    layout += `</div>`;

    layout += `<button class="btn btn-secondary swap-team-btn" onclick="sendSwapTeam()">Swap Team</button>`;

    return layout;
}

function sendSwapTeam() {
    sendMessage('swap_team');
}

function showGameScreen(status) {
    showScreen('game');
    renderGameBoard(status);
    updateActionButtons();
}

// Game actions
function createGame() {
    const playerName = document.getElementById('player-name').value.trim();
    if (!playerName) {
        showStatus('connection-status', 'Please enter your name', 'error');
        return;
    }
    gameState.playerName = playerName;
    gameState.gameID = ''; // empty = server will generate
    gameState.isCreator = true;
    gameState.waitingPlayers = [];
    gameState.teamAssignments = {};
    gameState.maxPlayers = { '1v1': 2, '2v2': 4, 'ffa3': 3, 'ffa5': 5 }[gameState.gameMode] || 2;
    connectWebSocket();
}

function joinGameByCode() {
    const playerName = document.getElementById('join-player-name').value.trim();
    const gameID = document.getElementById('game-id').value.trim().toUpperCase();
    if (!playerName || !gameID) {
        showStatus('connection-status', 'Please enter both name and game code', 'error');
        return;
    }
    gameState.playerName = playerName;
    gameState.gameID = gameID;
    gameState.waitingPlayers = [];
    gameState.teamAssignments = {};
    connectWebSocket();
}

function joinGameByUrl() {
    const playerName = document.getElementById('url-player-name').value.trim();
    if (!playerName) {
        showStatus('connection-status', 'Please enter your name', 'error');
        return;
    }
    gameState.playerName = playerName;
    // gameState.gameID already set from URL param
    gameState.waitingPlayers = [];
    gameState.teamAssignments = {};
    connectWebSocket();
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
        'harpoon': 'attack',
        'blood_rain': 'attack',
        'spy': 'spy/steal',
        'steal': 'spy/steal',
        'sabotage': 'spy/steal',
        'desertion': 'spy/steal',
        'catapult': 'spy/steal',
        'buy': 'buy',
        'construct': 'construct',
        'fortress': 'construct'
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
    // Remove selection mode classes from fields
    document.getElementById('player-field')?.classList.remove('selecting-ally');
    document.querySelectorAll('.opponent-field').forEach(f => {
        f.classList.remove('selecting-target');
        f.classList.remove('selecting-ally');
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

    // Handle move_warrior action
    if (action === 'move_warrior' && context === 'player-hand') {
        if (cardType !== 'warrior') return; // Only warriors can be moved
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
        toggleCardSelection(cardID, 'player-hand');
        if (gameState.selectedCards.length === 3) {
            // Show trade confirmation popup with 3 selected cards + 1 card back
            showTradeConfirmModal();
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
            }
        } else {
            // Regular attacks cannot target allies
            if (isAllyBoard) return;
        }

        gameState.actionState.targetPlayer = opponentName;
        handleAttackPhaseTargetClick(cardID, isAllyBoard ? 'ally' : 'enemy');
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
        gameState.actionState.weaponId = cardID;
        highlightSelectedCard(cardID);
        const enemies = getEnemyOpponents().filter(e => e.castle?.constructed && (e.castle?.resource_cards > 0 || e.castle?.is_protected));
        if (enemies.length === 1) {
            gameState.actionState.targetPlayer = enemies[0].player_name;
            showCatapultModal();
        } else if (enemies.length > 1) {
            showTargetPlayerModal('Select a castle to attack', enemies, (playerName) => {
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
        updateActionPrompt(`⚔️ ${weaponName} (${weaponDmg} DMG) - Select a target`);
        highlightValidTargets(card);
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
                gameState.actionState.targetPlayer = playerName;
                showBloodRainConfirmModal(card, enemy || { player_name: playerName, field: [] });
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

    // Enable target selection on the appropriate field based on warrior type
    // Don't pre-highlight targets - only the selected one will be highlighted
    enableSpecialPowerTargetSelection(userType);
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

    showActionConfirmModal({
        title: 'Blood Rain',
        cardsHtml: cardsHtml,
        description: `🩸 Blood Rain hits all of ${targetName}'s warriors — ${warriorSummary}`,
        onConfirm: () => {
            sendAction('blood_rain', {
                target_player: gameState.actionState.targetPlayer,
                weapon_id: gameState.actionState.weaponId
            });
            resetActionState();
        }
    });
}

function showAttackConfirmModal(weapon, target) {
    const weaponName = getCardName(weapon);
    const weaponDmg = weapon?.value || 0;
    const targetName = getCardName(target);
    const targetHp = target?.value || 0;
    const targetId = target?.id;
    const multiplier = weapon?.dmg_mult?.[targetId] || 1;
    const effectiveDmg = weaponDmg * multiplier;
    const hasDoubleDamage = multiplier > 1;

    const isProtected = target?.protected_by && target.protected_by.id;
    const shieldHp = isProtected ? (target.protected_by.value || 0) : 0;

    let cardsHtml = renderCardForModal(weapon, { showDoubleDamage: hasDoubleDamage });
    cardsHtml += renderArrow();
    cardsHtml += renderCardForModal(target, { showShield: isProtected, shieldHp: shieldHp });

    let description;
    let dmgLabel;
    if (hasDoubleDamage) {
        dmgLabel = `${weaponName} (${weaponDmg} x${multiplier} = ${effectiveDmg} DMG)`;
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
            description = `${userName} will heal ${targetName} (${targetHp} HP)`;
            break;
        default:
            description = `${userName} will use ${getCardName(specialPower)} on ${targetName}`;
    }

    let cardsHtml = renderCardForModal(user);
    cardsHtml += renderArrow();
    cardsHtml += renderCardForModal(target, { showShield: isProtected, shieldHp: shieldHp });

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
                gameState.actionState.targetPlayer = playerName;
                showStealModal();
            }, (opp) => `${opp.cards_in_hand} cards in hand`);
        }
    } else if (cardType === 'sabotage') {
        gameState.actionState.type = 'sabotage';
        gameState.pendingModalAction = 'sabotage';
        const enemies = getEnemyOpponents();
        if (enemies.length === 1) {
            sendAction('sabotage', { target_player: enemies[0].player_name });
        } else {
            showTargetPlayerModal('Select a player to sabotage', enemies,
                (playerName) => sendAction('sabotage', { target_player: playerName }),
                (opp) => `${opp.cards_in_hand} card(s) in hand`);
        }
    } else if (cardType === 'desertion') {
        gameState.actionState.type = 'desertion';
        const enemies = getEnemyOpponents();
        if (enemies.length === 1) {
            gameState.actionState.targetPlayer = enemies[0].player_name;
            showDesertionModal();
        } else {
            showTargetPlayerModal('Select a player to steal warrior from', enemies,
                (playerName) => {
                    gameState.actionState.targetPlayer = playerName;
                    showDesertionModal();
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

    // When gold >= 6, offer a choice: deck cards or mercenary
    if (resourceValue >= 6) {
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
                        <div class="player-detail">15 HP · uses any weapon · no special powers</div>
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
            const payload = {};
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
            const payload = {};
            if (targetPlayer) payload.target_player = targetPlayer;
            sendAction('resurrection', payload);
            resetActionState();
        }
    });
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

function showConstructTargetModal(resource, cardID, canConstructOwn, allies) {
    let content = '<div class="target-player-options">';

    // Option: own castle (if possible)
    if (canConstructOwn) {
        content += `
            <div class="target-player-option" onclick="window._constructTargetCallback('')">
                <span class="player-icon">🏰</span>
                <div class="player-info">
                    <div class="player-name">Your Castle</div>
                    <div class="player-detail">${gameState.currentState?.current_player_castle?.constructed ? 'Value: ' + (gameState.currentState.current_player_castle.value || 0) + '/25' : 'Start construction'}</div>
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
                    <div class="player-detail">Value: ${castleValue}/25</div>
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
    const newValue = currentValue + resourceValue;

    let cardsHtml = renderCardForModal(resource);
    cardsHtml += renderArrow();
    cardsHtml += renderCastleIcon();

    const castleLabel = targetName ? `${targetName}'s castle` : 'your castle';
    const description = castle?.constructed
        ? `${resourceName} (${resourceValue} gold) → ${castleLabel} value: ${currentValue} → ${newValue}/25`
        : `${resourceName} (${resourceValue} value) will be added to ${castleLabel}`;

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
    // For special powers, enable selection on player's field
    // Don't highlight all warriors - only the selected one will be highlighted
    const playerField = document.getElementById('player-field');
    // Enable hover/selection on player field for SpecialPower
    playerField.classList.add('selecting-ally');
    // Note: We don't add valid-target here anymore
    // The user will click to select, and that will highlight with 'selected' class
}

function enableSpecialPowerTargetSelection(userType) {
    // Enable target selection on the appropriate field based on warrior type
    const playerField = document.getElementById('player-field');

    if (userType === 'archer') {
        // Archer (Instant Kill) targets enemies only (not allies)
        document.querySelectorAll('.opponent-board:not(.ally) .opponent-field').forEach(f => f.classList.add('selecting-target'));
    } else {
        // Mage (Heal) and Knight (Protect) target own field + ally fields
        playerField.classList.add('selecting-ally');
        document.querySelectorAll('.opponent-board.ally .opponent-field').forEach(f => f.classList.add('selecting-ally'));
    }
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

    // Search in all opponent fields
    for (const opponent of status.opponents || []) {
        for (const card of opponent.field || []) {
            if (card.id === cardId) return card;
        }
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
        warriorId: null,
        targetPlayer: null
    };

    // Clear visual selections
    document.querySelectorAll('.card.selected, .card.valid-target').forEach(card => {
        card.classList.remove('selected', 'valid-target');
    });

    // Remove selection mode classes from fields
    document.getElementById('player-field')?.classList.remove('selecting-ally');
    document.querySelectorAll('.opponent-field').forEach(f => {
        f.classList.remove('selecting-target');
        f.classList.remove('selecting-ally');
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

function cancelAction() {
    resetActionState();
    updateActionPrompt('');
    // Re-render board to recalculate usable/unusable classes for the current phase
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
    const container = document.getElementById('player-hand');

    const cardElement = container.querySelector(`[data-card-id="${cardID}"]`);
    if (cardElement) {
        cardElement.classList.toggle('selected');
    }
}

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

// Show floating damage numbers when warriors take damage
let screenFlashShown = false;

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

// Display heal animation on a card with green cross and count-up
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

    // Count-up number
    const countup = document.createElement('div');
    countup.className = 'heal-countup';
    countup.textContent = fromHp;
    cardElement.appendChild(countup);

    let current = fromHp;
    const interval = setInterval(() => {
        current++;
        countup.textContent = current;
        if (current >= toHp) {
            clearInterval(interval);
        }
    }, 80);

    setTimeout(() => countup.remove(), 3500);
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

// ── Desertion Animation ──────────────────────────────────────────────────────

// Captures the warrior card element in the enemy field before re-render.
function prepareDesertionAnimation(previousState, newState) {
    if (newState.last_action !== 'desertion') return null;
    // Only animate from the attacker's perspective
    if (newState.turn_player !== newState.current_player) return null;

    const deserterID = newState.last_moved_warrior_id; // not used; search by field change
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
function playDesertionAnimation(data) {
    if (!data) return;

    const { clone, rect } = data;
    const playerField = document.getElementById('player-field');
    if (!playerField) return;
    const targetRect = playerField.getBoundingClientRect();

    const dx = (targetRect.left + targetRect.width / 2) - (rect.left + rect.width / 2);
    const dy = (targetRect.top + targetRect.height / 2) - (rect.top + rect.height / 2);

    const ghost = document.createElement('div');
    ghost.className = 'desertion-ghost';
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
        playerField.classList.add('desertion-landing-flash');
        setTimeout(() => playerField.classList.remove('desertion-landing-flash'), 700);
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

function renderGameBoard(status) {
    // Render all opponent boards
    renderOpponents(status.opponents || []);

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

function renderOpponents(opponents) {
    const container = document.getElementById('opponents-container');
    container.innerHTML = '';
    container.setAttribute('data-count', opponents.length);

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

        // Header
        const header = document.createElement('div');
        header.className = 'opponent-header';
        const badgeHtml = opponent.is_eliminated
            ? '<span class="opponent-badge eliminated-badge">Eliminated</span>'
            : opponent.is_disconnected
                ? '<span class="opponent-badge eliminated-badge">Disconnected</span>'
                : '';
        header.innerHTML = `
            <span class="opponent-name">${opponent.player_name}</span>
            ${opponent.is_ally ? '<span class="opponent-badge ally-badge">Ally</span>' : ''}
            ${badgeHtml}
        `;
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

        container.appendChild(board);
    });
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
        const castleGoal = gameState.gameMode === '2v2' ? 30 : 25;
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

    // Create card HTML
    if (imageUrl) {
        div.innerHTML = `
            ${shieldHtml}
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
        const castleGoal = gameState.gameMode === '2v2' ? 30 : 25;
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

// Timer functions
function startTimers(status) {
    if (timerInterval) {
        clearInterval(timerInterval);
    }

    const gameStartedAt = new Date(status.game_started_at);
    const turnStartedAt = new Date(status.turn_started_at);
    const turnLimit = status.turn_time_limit_secs || 60;
    const isGameOver = status.game_over_msg && status.game_over_msg.length > 0;

    if (isGameOver) {
        const turnTimerEl = document.getElementById('turn-timer');
        if (turnTimerEl) {
            turnTimerEl.textContent = '--';
            turnTimerEl.classList.remove('warning');
        }
        return;
    }

    function updateTimers() {
        const now = new Date();

        // Game timer (counting up)
        const gameElapsed = Math.floor((now - gameStartedAt) / 1000);
        const gameTimerEl = document.getElementById('game-timer');
        if (gameTimerEl) {
            gameTimerEl.textContent = formatTime(gameElapsed);
        }

        // Turn timer (counting down)
        const turnElapsed = Math.floor((now - turnStartedAt) / 1000);
        const turnRemaining = Math.max(0, turnLimit - turnElapsed);
        const turnTimerEl = document.getElementById('turn-timer');
        if (turnTimerEl) {
            turnTimerEl.textContent = formatCountdown(turnRemaining);
            if (turnRemaining <= 10) {
                turnTimerEl.classList.add('warning');
            } else {
                turnTimerEl.classList.remove('warning');
            }
        }
    }

    updateTimers();
    timerInterval = setInterval(updateTimers, 1000);
}

function formatTime(totalSeconds) {
    const hours = Math.floor(totalSeconds / 3600);
    const mins = Math.floor((totalSeconds % 3600) / 60);
    const secs = totalSeconds % 60;
    if (hours > 0) {
        return `${hours}:${String(mins).padStart(2, '0')}:${String(secs).padStart(2, '0')}`;
    }
    return `${String(mins).padStart(2, '0')}:${String(secs).padStart(2, '0')}`;
}

function formatCountdown(totalSeconds) {
    const mins = Math.floor(totalSeconds / 60);
    const secs = totalSeconds % 60;
    return `${mins}:${String(secs).padStart(2, '0')}`;
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
    if (type === 'spy' || type === 'thief' || type === 'catapult' || type === 'fortress' || type === 'resurrection' || type === 'sabotage' || type === 'desertion' || type === 'ambush') return 'special';
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

// Card metadata fetched from /api/card-config on load.
// Each entry: { description: string, image: string }
let cardConfig = {};

function getCardImageUrl(card) {
    const key = (card.sub_type || card.type || '').toLowerCase();
    const image = cardConfig[key]?.image;
    return image ? `/static/img/cards/${image}` : null;
}

function getCardStatBadge(card, cardType) {
    if (cardType === 'warrior') {
        return `<span class="card-stat-badge warrior">HP ${card.value || 0}</span>`;
    } else if (cardType === 'weapon') {
        return `<span class="card-stat-badge weapon">DMG ${card.value || 0}</span>`;
    } else if (cardType === 'resource') {
        return `<span class="card-stat-badge resource">${card.value || 0}</span>`;
    }
    return '';
}

function isWarrior(card) {
    const type = getCardType(card);
    return type === 'warrior';
}

function updateTurnIndicator() {
    // Update the phase tracker turn status
    updatePhaseTracker();
}

function updateActionButtons() {
    const isYourTurn = gameState.isYourTurn;
    const status = gameState.currentState;

    // Disable all action buttons first
    document.querySelectorAll('.btn-action, #skip-phase-btn, #end-turn-btn').forEach(btn => {
        btn.disabled = true;
    });

    // Hide endturn popup by default
    const endturnPopup = document.getElementById('endturn-popup');
    endturnPopup.classList.add('hidden');

    if (!isYourTurn || !status) {
        clearEndTurnCountdown();
        return;
    }

    // In endturn phase, show the popup and start auto-countdown
    if (status.current_action === 'endturn') {
        document.getElementById('end-turn-btn').disabled = false;
        endturnPopup.classList.remove('hidden');
        if (!endTurnCountdownTimer) {
            startEndTurnCountdown(status.next_turn_player);
        }
        return;
    }

    // Left endturn phase while still our turn (shouldn't normally happen)
    clearEndTurnCountdown();

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
    const phaseTracker = document.getElementById('phase-tracker');
    const turnStatusElement = document.getElementById('phase-turn-status');
    const turnTextElement = turnStatusElement?.querySelector('.turn-text');
    const gameScreen = document.getElementById('game-screen');

    // Update turn status
    if (turnStatusElement && turnTextElement) {
        turnStatusElement.classList.remove('your-turn', 'enemy-turn');
        phaseTracker?.classList.remove('your-turn', 'enemy-turn');
        gameScreen?.classList.remove('your-turn', 'enemy-turn');
        if (gameState.isYourTurn) {
            turnStatusElement.classList.add('your-turn');
            phaseTracker?.classList.add('your-turn');
            gameScreen?.classList.add('your-turn');
            turnTextElement.textContent = 'Your Turn';
        } else {
            turnStatusElement.classList.add('enemy-turn');
            phaseTracker?.classList.add('enemy-turn');
            gameScreen?.classList.add('enemy-turn');
            const turnPlayer = gameState.currentState?.turn_player || 'Enemy';
            turnTextElement.textContent = `${turnPlayer}'s Turn`;
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

function updatePlayerListPanel() {
    const panel = document.getElementById('player-list-panel');
    const list = document.getElementById('player-list');
    if (!panel || !list) return;

    // Hide for 1v1
    if (gameState.gameMode === '1v1') {
        panel.classList.add('hidden');
        repositionLeftPanels();
        return;
    }

    panel.classList.remove('hidden');

    const status = gameState.currentState;
    if (!status) return;

    const turnPlayer = status.turn_player;
    const opponents = status.opponents || [];
    const playersOrder = status.players_order || [];

    // Build a lookup for opponent info
    const opponentMap = {};
    for (const opp of opponents) {
        opponentMap[opp.player_name] = opp;
    }

    list.innerHTML = '';

    // Use server's turn order so all clients see the same list
    for (const name of playersOrder) {
        const isSelf = name === gameState.playerName;
        const opp = opponentMap[name];
        const isAlly = opp ? (opp.is_ally || false) : false;
        const isEliminated = isSelf ? (status.is_eliminated || false) : (opp ? opp.is_eliminated : false);
        const isTurn = name === turnPlayer;

        const li = document.createElement('li');
        let classes = 'player-list-item';
        if (isTurn) classes += ' is-current-turn';
        if (isSelf) classes += ' is-self';
        if (isAlly) classes += ' is-ally';
        if (isEliminated) classes += ' is-eliminated';
        li.className = classes;

        let tag = '';
        if (isSelf) tag = 'YOU';
        else if (isAlly) tag = 'ALLY';

        li.innerHTML = `
            <span class="player-list-dot"></span>
            <span>${name}</span>
            ${tag ? `<span class="player-list-tag">${tag}</span>` : ''}
        `;
        list.appendChild(li);
    }

    repositionLeftPanels();
}

function repositionLeftPanels() {
    const phaseTracker = document.getElementById('phase-tracker');
    const playerPanel = document.getElementById('player-list-panel');
    const historyPanel = document.querySelector('.history-panel');

    if (!phaseTracker) return;

    const phaseBottom = phaseTracker.offsetTop + phaseTracker.offsetHeight + 10;

    if (playerPanel && !playerPanel.classList.contains('hidden')) {
        playerPanel.style.top = phaseBottom + 'px';
        const playerBottom = phaseBottom + playerPanel.offsetHeight + 10;
        if (historyPanel) {
            historyPanel.style.top = playerBottom + 'px';
        }
    } else {
        if (historyPanel) {
            historyPanel.style.top = phaseBottom + 'px';
        }
    }
}

function updateActionPrompt(text) {
    const prompt = document.getElementById('action-prompt');
    const container = document.getElementById('action-prompt-container');
    prompt.textContent = text;

    // Show/hide the prompt container based on whether there's text
    if (text) {
        container.classList.remove('hidden');
    } else {
        container.classList.add('hidden');
    }
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

    // Play deferred animations that were waiting for modal close
    if (pendingAnimationsCallback) {
        const callback = pendingAnimationsCallback;
        pendingAnimationsCallback = null;
        callback();
    }
}

// End Turn Countdown
function startEndTurnCountdown(nextPlayer) {
    const nameEl = document.getElementById('endturn-next-player-name');
    const bar = document.getElementById('endturn-countdown-bar');

    if (nameEl) nameEl.textContent = nextPlayer || '—';

    if (bar) {
        bar.style.animation = 'none';
        bar.offsetHeight; // force reflow
        bar.style.animation = `endturnCountdown ${END_TURN_COUNTDOWN_SECS}s linear forwards`;
    }

    endTurnCountdownTimer = setTimeout(() => {
        endTurnCountdownTimer = null;
        sendAction('end_turn');
    }, END_TURN_COUNTDOWN_SECS * 1000);
}

function clearEndTurnCountdown() {
    if (endTurnCountdownTimer) {
        clearTimeout(endTurnCountdownTimer);
        endTurnCountdownTimer = null;
    }
    const bar = document.getElementById('endturn-countdown-bar');
    if (bar) bar.style.animation = 'none';
}

// Turn Transition Modal
let turnTransitionTimer = null;

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

// Stolen Card Notification Modal
let stolenCardTimer = null;

function showStolenCardModal(card) {
    const modal = document.getElementById('stolen-card-modal');
    const container = document.getElementById('stolen-card-container');
    const text = document.getElementById('stolen-card-text');

    if (!modal || !container || !text) return;

    // Render the stolen card
    const cardName = card.sub_type || card.type;
    container.innerHTML = renderCardForModal(card);
    text.textContent = `${cardName} was stolen from you!`;

    modal.classList.remove('hidden');

    if (stolenCardTimer) clearTimeout(stolenCardTimer);
    stolenCardTimer = setTimeout(() => {
        hideStolenCardModal();
    }, 4000);
}

function hideStolenCardModal() {
    const modal = document.getElementById('stolen-card-modal');
    modal.classList.add('hidden');
    if (stolenCardTimer) {
        clearTimeout(stolenCardTimer);
        stolenCardTimer = null;
    }
}

// Spy Notification Modal
let spyNotificationTimer = null;

function showSpyNotificationModal(message) {
    const modal = document.getElementById('spy-notification-modal');
    const textEl = document.getElementById('spy-notification-text');

    if (!modal || !textEl) return;

    textEl.textContent = message;
    modal.classList.remove('hidden');

    if (spyNotificationTimer) clearTimeout(spyNotificationTimer);
    spyNotificationTimer = setTimeout(() => {
        hideSpyNotificationModal();
    }, 3000);
}

function hideSpyNotificationModal() {
    const modal = document.getElementById('spy-notification-modal');
    if (modal) modal.classList.add('hidden');
    if (spyNotificationTimer) {
        clearTimeout(spyNotificationTimer);
        spyNotificationTimer = null;
    }
}

// Desertion Notification Modal — shown to the player whose warrior was stolen
let desertionNotificationTimer = null;

function showDesertionNotificationModal(notification) {
    const modal = document.getElementById('desertion-notification-modal');
    const container = document.getElementById('desertion-warrior-container');
    const text = document.getElementById('desertion-notification-text');
    if (!modal || !container || !text) return;

    const warrior = notification.warrior_card;
    const stolenBy = notification.stolen_by;
    container.innerHTML = renderCardForModal(warrior);
    const warriorName = warrior.sub_type || warrior.type || 'Warrior';
    text.textContent = `${warriorName} (${warrior.value} HP) deserted to ${stolenBy}!`;

    modal.classList.remove('hidden');

    if (desertionNotificationTimer) clearTimeout(desertionNotificationTimer);
    desertionNotificationTimer = setTimeout(() => hideDesertionNotificationModal(), 5000);
}

function hideDesertionNotificationModal() {
    const modal = document.getElementById('desertion-notification-modal');
    if (modal) modal.classList.add('hidden');
    if (desertionNotificationTimer) {
        clearTimeout(desertionNotificationTimer);
        desertionNotificationTimer = null;
    }
}

// Ambush placed animation — shown to all players when a player places an ambush card
function showAmbushPlacedAnimation(gameStatus) {
    const isOwnField = gameStatus.turn_player === gameStatus.current_player;

    let fieldEl;
    if (isOwnField) {
        fieldEl = document.getElementById('player-field');
    } else {
        const opponentFields = document.querySelectorAll('.opponent-field');
        for (const f of opponentFields) {
            if (f.dataset.opponentName === gameStatus.turn_player) {
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

// Ambush Place Confirmation Modal
function showAmbushPlaceConfirmModal(card, cardID) {
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
            sendMessage('place_ambush', { card_id: cardID });
            resetActionState();
        },
    });
}

// Ambush Triggered Modal
const ambushEffectDescriptions = {
    'Reflect Damage':   'Your weapon damage was reflected back — your warrior took the hit instead.',
    'Attack Cancelled': 'The attack was cancelled. Your weapon was discarded.',
    'Weapon Stolen':    'Your weapon was intercepted and added to the defender\'s hand.',
    'Drain Life':       'The attack was absorbed — the warrior took no damage and gained HP equal to the weapon\'s damage.',
    'Instant Kill':     'One of your warriors was instantly killed.',
};

const ambushEffectColorClass = {
    'Reflect Damage':   'ambush-effect-reflect',
    'Attack Cancelled': 'ambush-effect-cancel',
    'Weapon Stolen':    'ambush-effect-steal',
    'Drain Life':       'ambush-effect-drain',
    'Instant Kill':     'ambush-effect-instant',
};

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
}

function hideAmbushTriggeredModal() {
    const modal = document.getElementById('ambush-triggered-modal');
    if (modal) modal.classList.add('hidden');
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

// Action Confirm Modal Functions
let actionConfirmCallback = null;

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
    const defaultDetail = (opp) => {
        const castle = opp.castle || {};
        return `Castle: ${castle.value || 0}/25 gold, ${castle.resource_cards || 0} resource cards`;
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
    sendAction('steal', { target_player: gameState.actionState.targetPlayer, card_position: position });
    hideGameModal();
}

// Desertion Modal — shows eligible warriors (≤5 HP) from the target opponent's field
function showDesertionModal() {
    const targetName = gameState.actionState.targetPlayer;
    const opponent = getOpponentByName(targetName);
    const field = (opponent?.field || []).filter(w => w.value <= 5);

    if (field.length === 0) {
        updateActionPrompt(`${targetName} has no warriors with 5 HP or less!`);
        resetActionState();
        return;
    }

    let content = '<div class="desertion-warrior-grid">';
    field.forEach(warrior => {
        const subType = warrior.sub_type || warrior.type || 'Warrior';
        const imgKey = subType.toLowerCase();
        const imgSrc = `/static/img/cards/${imgKey}.webp`;
        content += `
            <div class="desertion-warrior-option" onclick="selectDesertionWarrior('${warrior.id}')">
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

    showGameModal(`Desertion — ${targetName}`, 'Choose a weakened warrior to convince (≤5 HP)', content, true);
}

function selectDesertionWarrior(warriorID) {
    sendAction('desertion', { target_player: gameState.actionState.targetPlayer, warrior_id: warriorID });
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
    hideGameModal();

    if (option === 1) {
        // Reveal deck - no target player needed
        gameState.pendingModalAction = 'spy_deck';
        const enemies = getEnemyOpponents();
        // Backend requires a target_player even for deck spy; use first enemy
        sendAction('spy', { target_player: enemies[0]?.player_name || '', option: option });
    } else {
        // Reveal hand - need to select target player
        const enemies = getEnemyOpponents();
        if (enemies.length === 1) {
            gameState.pendingModalAction = 'spy_hand';
            sendAction('spy', { target_player: enemies[0].player_name, option: option });
        } else {
            showTargetPlayerModal('Whose hand do you want to reveal?', enemies, (playerName) => {
                gameState.pendingModalAction = 'spy_hand';
                sendAction('spy', { target_player: playerName, option: option });
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
        const content = `<div style="text-align:center;padding:16px 8px;">
            <svg style="width:56px;height:68px;filter:drop-shadow(0 0 10px rgba(212,168,37,0.9))" viewBox="0 0 80 96" xmlns="http://www.w3.org/2000/svg">
                <defs><linearGradient id="shieldGradM" x1="30%" y1="0%" x2="70%" y2="100%"><stop offset="0%" stop-color="#3b82f6"/><stop offset="100%" stop-color="#1e3a8a"/></linearGradient></defs>
                <path d="M40 4 L76 18 L76 52 Q76 80 40 92 Q4 80 4 52 L4 18 Z" fill="url(#shieldGradM)" stroke="#d4a825" stroke-width="4.5" stroke-linejoin="round"/>
                <line x1="40" y1="24" x2="40" y2="70" stroke="rgba(255,255,255,0.4)" stroke-width="3" stroke-linecap="round"/>
                <line x1="21" y1="47" x2="59" y2="47" stroke="rgba(255,255,255,0.4)" stroke-width="3" stroke-linecap="round"/>
            </svg>
            <p style="color:#e8c244;font-weight:700;margin:12px 0 4px;">Fortress Wall</p>
            <p style="color:#a0a0a0;font-size:0.85em;margin:0 0 16px;">The fortress wall will be destroyed instead of gold</p>
            <button class="btn btn-danger" onclick="confirmCatapultFortress('${targetName}')">Destroy Wall</button>
        </div>`;
        showGameModal(`Catapult ${targetName}'s Castle`, 'Castle is fortified', content, true);
        return;
    }

    if (resourceCount === 0) {
        updateActionPrompt('Castle has no resources to attack!');
        resetActionState();
        return;
    }

    let content = '<div class="catapult-cards-grid">';
    for (let i = 1; i <= resourceCount; i++) {
        content += `
            <div class="card-face-down catapult-target" data-position="${i}" onclick="selectCatapultPosition(${i})">
                <div class="card-back-design">
                    <span class="card-back-icon">💰</span>
                </div>
                <span class="card-position">#${i}</span>
            </div>
        `;
    }
    content += '</div>';

    showGameModal(`Catapult ${targetName}'s Castle`, 'Select a resource card to destroy', content, true);
}

function selectCatapultPosition(position) {
    sendAction('catapult', { target_player: gameState.actionState.targetPlayer, card_position: position });
    hideGameModal();
}

function confirmCatapultFortress(targetName) {
    sendAction('catapult', { target_player: targetName, card_position: 1 });
    hideGameModal();
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
