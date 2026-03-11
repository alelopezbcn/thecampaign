function updateTurnIndicator() {
    updatePhaseBadge();
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

    // Forge - enabled if can_forge is true (from backend)
    document.getElementById('forge-btn').disabled = !status.can_forge;

    // Skip Phase and End Turn - always enabled during your turn
    document.getElementById('skip-phase-btn').disabled = false;
    document.getElementById('end-turn-btn').disabled = false;
}

function updatePhaseBadge() {
    const status = gameState.currentState;
    const gameScreen = document.getElementById('game-screen');

    // Update your-turn/enemy-turn class on game screen (drives card hover styles)
    gameScreen?.classList.remove('your-turn', 'enemy-turn');
    gameScreen?.classList.add(gameState.isYourTurn ? 'your-turn' : 'enemy-turn');

    // Update phase badge strip inside the player board header
    const badge = document.getElementById('phase-badge');
    if (!badge) return;

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
                el.classList.add(gameState.executedPhases.includes(phase) ? 'done' : 'skipped');
            }
        });
    } else {
        badge.classList.add('hidden');
    }
}

// Keep old function names for compatibility
function updatePhaseTracker() { updatePhaseBadge(); }
function updatePhaseIndicator() { updatePhaseBadge(); }

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
    const TOP = 60;
    const playerPanel = document.getElementById('player-list-panel');
    const historyPanel = document.getElementById('history-panel');

    if (playerPanel && !playerPanel.classList.contains('hidden')) {
        playerPanel.style.top = TOP + 'px';
        const playerBottom = TOP + playerPanel.offsetHeight + 10;
        if (historyPanel) historyPanel.style.top = playerBottom + 'px';
    } else {
        if (historyPanel) historyPanel.style.top = TOP + 'px';
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

    const turnStartedAt = new Date(status.turn_started_at);
    const turnLimit = status.turn_time_limit_secs || 60;
    const isGameOver = status.game_over_msg && status.game_over_msg.length > 0;
    const turnTimerEl = document.getElementById('turn-timer');

    if (isGameOver) {
        if (turnTimerEl) {
            turnTimerEl.textContent = '--';
            turnTimerEl.classList.remove('warning');
        }
        return;
    }

    function updateTurnTimer() {
        const turnElapsed = Math.floor((new Date() - turnStartedAt) / 1000);
        const turnRemaining = Math.max(0, turnLimit - turnElapsed);
        if (turnTimerEl) {
            turnTimerEl.textContent = formatCountdown(turnRemaining);
            turnTimerEl.classList.toggle('warning', turnRemaining <= 10);
        }
    }

    updateTurnTimer();
    timerInterval = setInterval(updateTurnTimer, 1000);
}

// Format mm:ss or h:mm:ss — used by game-over stats for total duration
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

function initSidePanelToggles() {
    const historyPanel = document.getElementById('history-panel');
    const historyBtn   = document.getElementById('history-toggle-btn');
    const playersPanel = document.getElementById('player-list-panel');
    const playersBtn   = document.getElementById('player-list-toggle-btn');

    if (historyBtn && historyPanel) {
        historyPanel.classList.add('collapsed');
        historyPanel.addEventListener('click', (e) => {
            if (historyPanel.classList.contains('collapsed')) {
                historyPanel.classList.remove('collapsed');
            } else if (e.target === historyBtn || historyBtn.contains(e.target)) {
                historyPanel.classList.add('collapsed');
            }
            repositionLeftPanels();
        });
    }

    if (playersBtn && playersPanel) {
        playersPanel.addEventListener('click', (e) => {
            if (playersPanel.classList.contains('collapsed')) {
                playersPanel.classList.remove('collapsed');
            } else if (e.target === playersBtn || playersBtn.contains(e.target)) {
                playersPanel.classList.add('collapsed');
            }
            repositionLeftPanels();
        });
    }
}
