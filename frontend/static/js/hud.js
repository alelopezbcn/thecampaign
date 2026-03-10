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

    // Forge - enabled if can_forge is true (from backend)
    document.getElementById('forge-btn').disabled = !status.can_forge;

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

    // Update phase badge strip in action bar
    const badge = document.getElementById('phase-badge');
    if (badge) {
        const badgePhaseOrder = ['attack', 'spy/steal', 'buy', 'construct'];
        const currentPhase = status?.current_action;
        if (gameState.isYourTurn && currentPhase && currentPhase !== 'endturn') {
            badge.classList.remove('hidden');
            const currentIdx = badgePhaseOrder.indexOf(currentPhase);
            badge.querySelectorAll('.pb-phase').forEach(el => {
                const phase = el.dataset.phase;
                const idx = badgePhaseOrder.indexOf(phase);
                el.classList.remove('active', 'done', 'skipped');
                if (idx === currentIdx) {
                    el.classList.add('active');
                } else if (idx < currentIdx) {
                    if (gameState.executedPhases.includes(phase)) {
                        el.classList.add('done');
                    } else {
                        el.classList.add('skipped');
                    }
                }
            });
        } else {
            badge.classList.add('hidden');
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
