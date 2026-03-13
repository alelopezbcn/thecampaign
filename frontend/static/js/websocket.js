// WebSocket functions
function connectWebSocket() {
    const protocol = window.location.protocol === 'https:' ? 'wss:' : 'ws:';
    const wsUrl = `${protocol}//${window.location.host}/ws`;

    // Close any existing connection cleanly before opening a new one.
    // Nulling out onclose first prevents the stale handler from clobbering
    // the new ws reference or triggering an unwanted reconnect cycle.
    if (ws) {
        ws.onclose = null;
        ws.onmessage = null;
        ws.onerror = null;
        ws.close();
        ws = null;
    }

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
            const joinPayload = {
                player_name: gameState.playerName,
                game_id: gameState.gameID,
                game_mode: gameState.gameMode
            };
            if (gameState.isCreator) {
                joinPayload.game_config = gameState.gameConfig;
            }
            sendMessage('join_game', joinPayload);
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
        case 'player_disconnected':
            startDisconnectCountdown(message.payload.player_name, message.payload.grace_period_secs);
            break;
        case 'player_reconnected':
            clearDisconnectCountdown(message.payload.player_name);
            hideWaitingForReconnectModal();
            break;
        case 'waiting_for_reconnect':
            showWaitingForReconnectModal(message.payload.disconnected_players, message.payload.secs_until_game_ends);
            break;
        default:
            console.log('Unknown message type:', message.type);
    }
}

function startDisconnectCountdown(playerName, seconds) {
    clearDisconnectCountdown(playerName);
    disconnectCountdowns[playerName] = { secondsLeft: seconds };
    disconnectCountdowns[playerName].intervalId = setInterval(() => {
        const entry = disconnectCountdowns[playerName];
        if (!entry) return;
        entry.secondsLeft--;
        if (entry.secondsLeft <= 0) {
            clearDisconnectCountdown(playerName);
        } else {
            const badge = document.querySelector(`.disconnect-countdown[data-player="${CSS.escape(playerName)}"]`);
            if (badge) badge.textContent = `Disconnected (${entry.secondsLeft}s)`;
        }
    }, 1000);
}

function clearDisconnectCountdown(playerName) {
    const entry = disconnectCountdowns[playerName];
    if (!entry) return;
    clearInterval(entry.intervalId);
    delete disconnectCountdowns[playerName];
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

    // Detect treason (warrior flying from enemy field to own field)
    let treasonData = null;
    if (previousState) {
        treasonData = prepareTreasonAnimation(previousState, payload.game_status);
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
        if (treasonData) {
            playTreasonAnimation(treasonData);
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
        showStolenCardModal(sabotagedCards[0], 'sabotaged');
    }

    // Detect spy notification
    const spyNotification = payload.game_status.spy_notification;
    if (spyNotification) {
        showSpyNotificationModal(spyNotification);
    }

    // Detect ambush triggered
    const ambushTriggered = payload.game_status.ambush_triggered;
    if (ambushTriggered) {
        showAmbushTriggeredModal(ambushTriggered);
    }

    // Detect treason notification (victim only)
    const treasonNotification = payload.game_status.treason_notification;
    if (treasonNotification) {
        showTreasonNotificationModal(treasonNotification);
    }

    // Detect Champion's Bounty (shown to all players)
    const championsBounty = payload.game_status.champions_bounty;
    if (championsBounty) {
        showChampionsBountyModal(championsBounty);
    }

    // Detect Resurrection (shown to all players)
    const resurrectionNotification = payload.game_status.resurrection_notification;
    if (resurrectionNotification) {
        showResurrectionModal(resurrectionNotification);
    }

    // Detect Catapult (target gets modal, others get toast)
    const catapultNotification = payload.game_status.catapult_notification;
    if (catapultNotification) {
        showCatapultNotificationModal(catapultNotification);
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
            } else if (gameState.pendingAction === 'forge') {
                showForgeResultModal(acquiredCards);
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
        showGameOverModal(isWinner, gameOverMsg, payload.game_status.player_stats, payload.game_status.game_started_at);
    }

    // Check for error message
    const errorMsg = payload.game_status.error_msg;
    if (errorMsg && errorMsg.length > 0) {
        showErrorToast(errorMsg);
    }

    // Event banner (always update)
    renderEventBanner(payload.game_status);

    // Event change toast for all players when the round event changes
    const prevEvent = previousState?.current_event;
    if (prevEvent !== undefined && prevEvent !== payload.game_status.current_event) {
        showEventChangeToast(payload.game_status);
    }

    // Event turn modal for active player at turn start
    if (!wasYourTurn && isNowYourTurn) {
        showEventTurnModal(payload.game_status);
    }
}

function checkIsWinner(gameOverMsg, status) {
    return !!status.is_winner;
}

function handleGameEnded() {
    hideWaitingForReconnectModal();
    const gameOverMsg = gameState.currentState?.game_over_msg || 'Game Over!';
    const isWinner = checkIsWinner(gameOverMsg, gameState.currentState || {});
    showGameOverModal(isWinner, gameOverMsg, gameState.currentState?.player_stats, gameState.currentState?.game_started_at);
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
        'place_ambush': 'attack',
        'resurrection': 'attack',
        'treason': 'attack',
        'catapult': 'attack',
        'spy': 'spy/steal',
        'steal': 'spy/steal',
        'sabotage': 'spy/steal',
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
