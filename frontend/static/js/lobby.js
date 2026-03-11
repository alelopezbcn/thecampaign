// --- Game Settings / Presets ---

function loadPresets() {
    fetch('/static/presets.json')
        .then(r => r.json())
        .then(data => {
            presetsData = data.presets || [];
            const select = document.getElementById('settings-preset');
            if (!select) return;
            select.innerHTML = '';
            presetsData.forEach((p, i) => {
                const opt = document.createElement('option');
                opt.value = i;
                opt.textContent = p.name;
                select.appendChild(opt);
            });
            const customOpt = document.createElement('option');
            customOpt.value = 'custom';
            customOpt.textContent = 'Custom';
            select.appendChild(customOpt);
            // Apply first preset as default
            if (presetsData.length > 0) {
                applyPreset(presetsData[0].config);
            }
        })
        .catch(() => {
            // Fallback: fill inputs with DEFAULT_GAME_CONFIG
            applyPreset(DEFAULT_GAME_CONFIG);
        });
}

function applyPreset(config) {
    const fields = {
        'cfg-warriors':            'warriors',
        'cfg-dragons':             'dragons',
        'cfg-harpoons':            'harpoons',
        'cfg-special-powers':      'special_powers',
        'cfg-spies':               'spies',
        'cfg-thieves':             'thieves',
        'cfg-sabotages':           'sabotages',
        'cfg-catapults':           'catapults',
        'cfg-fortresses':          'fortresses',
        'cfg-ambushes':            'ambushes',
        'cfg-blood-rains':         'blood_rains',
        'cfg-resurrections':       'resurrections',
        'cfg-treasons':          'treasons',
        'cfg-construction-cards':  'construction_cards',
        'cfg-high-value-gold':     'high_value_gold_cards',
        'cfg-castle-goal':         'castle_goal'
    };
    for (const [id, key] of Object.entries(fields)) {
        const el = document.getElementById(id);
        if (el && config[key] !== undefined) el.value = config[key];
    }
    gameState.gameConfig = { ...config };
}

function readConfigFromInputs() {
    return {
        warriors:           parseInt(document.getElementById('cfg-warriors').value) || 5,
        dragons:            parseInt(document.getElementById('cfg-dragons').value) || 0,
        harpoons:           parseInt(document.getElementById('cfg-harpoons').value) || 0,
        special_powers:     parseInt(document.getElementById('cfg-special-powers').value) || 0,
        spies:              parseInt(document.getElementById('cfg-spies').value) || 0,
        thieves:            parseInt(document.getElementById('cfg-thieves').value) || 0,
        sabotages:          parseInt(document.getElementById('cfg-sabotages').value) || 0,
        catapults:          parseInt(document.getElementById('cfg-catapults').value) || 0,
        fortresses:         parseInt(document.getElementById('cfg-fortresses').value) || 0,
        ambushes:           parseInt(document.getElementById('cfg-ambushes').value) || 0,
        blood_rains:        parseInt(document.getElementById('cfg-blood-rains').value) || 0,
        resurrections:      parseInt(document.getElementById('cfg-resurrections').value) || 0,
        treasons:         parseInt(document.getElementById('cfg-treasons').value) || 0,
        construction_cards:    parseInt(document.getElementById('cfg-construction-cards').value) || 1,
        high_value_gold_cards: parseInt(document.getElementById('cfg-high-value-gold').value) || 0,
        castle_goal:           parseInt(document.getElementById('cfg-castle-goal').value) || DEFAULT_CASTLE_GOAL
    };
}

function markConfigCustom() {
    const select = document.getElementById('settings-preset');
    if (select) select.value = 'custom';
}

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

    // Settings modal open/close
    document.getElementById('show-settings-link').addEventListener('click', () => {
        document.getElementById('game-settings-modal').classList.remove('hidden');
    });
    document.getElementById('settings-close-btn').addEventListener('click', () => {
        document.getElementById('game-settings-modal').classList.add('hidden');
    });
    document.getElementById('settings-apply-btn').addEventListener('click', () => {
        document.getElementById('game-settings-modal').classList.add('hidden');
    });
    document.getElementById('game-settings-modal').addEventListener('click', (e) => {
        if (e.target === document.getElementById('game-settings-modal')) {
            document.getElementById('game-settings-modal').classList.add('hidden');
        }
    });

    // Preset dropdown change
    document.getElementById('settings-preset').addEventListener('change', (e) => {
        const idx = parseInt(e.target.value);
        if (!isNaN(idx) && presetsData[idx]) {
            applyPreset(presetsData[idx].config);
        }
    });

    // Mark as custom when any input is manually changed
    const cfgInputIds = [
        'cfg-warriors', 'cfg-dragons', 'cfg-harpoons', 'cfg-special-powers',
        'cfg-spies', 'cfg-thieves', 'cfg-sabotages', 'cfg-catapults',
        'cfg-fortresses', 'cfg-ambushes', 'cfg-blood-rains', 'cfg-resurrections',
        'cfg-treasons', 'cfg-construction-cards', 'cfg-castle-goal'
    ];
    cfgInputIds.forEach(id => {
        document.getElementById(id).addEventListener('input', markConfigCustom);
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

    // Game screen actions
    document.getElementById('trade-btn').addEventListener('click', () => startAction('trade'));
    document.getElementById('forge-btn').addEventListener('click', () => startForgeMode());
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

    // Stolen card modal close
    document.getElementById('stolen-card-close').addEventListener('click', hideStolenCardModal);

    // Spy notification modal close
    document.getElementById('spy-notification-close').addEventListener('click', hideSpyNotificationModal);

    // Treason notification modal close
    document.getElementById('treason-notification-close').addEventListener('click', hideTreasonNotificationModal);

    // Event turn modal close
    document.getElementById('event-turn-modal-close').addEventListener('click', hideEventTurnModal);
    document.getElementById('event-turn-modal').addEventListener('click', (e) => {
        if (e.target === e.currentTarget) hideEventTurnModal();
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
        { id: 'treason-notification-modal', hide: hideTreasonNotificationModal },
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

    const stolenCardModal = document.getElementById('stolen-card-modal');
    const spyNotificationModal = document.getElementById('spy-notification-modal');
    const treasonNotificationModal = document.getElementById('treason-notification-modal');
    const eventTurnModal = document.getElementById('event-turn-modal');

    const isStolenCardOpen = stolenCardModal && !stolenCardModal.classList.contains('hidden');
    const isSpyNotificationOpen = spyNotificationModal && !spyNotificationModal.classList.contains('hidden');
    const isTreasonNotificationOpen = treasonNotificationModal && !treasonNotificationModal.classList.contains('hidden');
    const isEventTurnOpen = eventTurnModal && !eventTurnModal.classList.contains('hidden');

    if (e.key === 'Escape') {
        if (isStolenCardOpen) {
            hideStolenCardModal();
        } else if (isSpyNotificationOpen) {
            hideSpyNotificationModal();
        } else if (isTreasonNotificationOpen) {
            hideTreasonNotificationModal();
        } else if (isEventTurnOpen) {
            hideEventTurnModal();
        } else if (isActionConfirmOpen) {
            onActionConfirmNo();
        } else if (isGameModalOpen) {
            hideGameModal();
        } else if (isActionPromptOpen) {
            cancelAction();
        }
    } else if (e.key === 'Enter') {
        if (isTreasonNotificationOpen) {
            hideTreasonNotificationModal();
        } else if (isEventTurnOpen) {
            hideEventTurnModal();
        } else if (isActionConfirmOpen) {
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
    gameState.gameConfig = readConfigFromInputs();
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

// Initialize
document.addEventListener('DOMContentLoaded', () => {
    setupEventListeners();
    checkUrlParams();
    loadPresets();
    initSidePanelToggles();
    initAutoCollapseCheckbox();
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
