// Game state
let ws = null;
let reconnectAttempts = 0;
let reconnectTimer = null;
let timerInterval = null;
let pendingAnimationsCallback = null; // Deferred animations waiting for modal close
let endTurnCountdownTimer = null;
const END_TURN_COUNTDOWN_SECS = 3;
const disconnectCountdowns = {}; // playerName -> { secondsLeft, intervalId }
const MAX_RECONNECT_ATTEMPTS = 20;
const DEFAULT_CASTLE_GOAL = 25;
const DEFAULT_GAME_CONFIG = {
    warriors: 5,
    dragons: 1,
    harpoons: 1,
    special_powers: 4,
    spies: 1,
    thieves: 1,
    sabotages: 1,
    catapults: 1,
    fortresses: 1,
    ambushes: 1,
    blood_rains: 2,
    resurrections: 1,
    treasons: 1,
    construction_cards: 1,
    high_value_gold_cards: 0,
    castle_goal: DEFAULT_CASTLE_GOAL
};

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
    gameConfig: { ...DEFAULT_GAME_CONFIG }, // Game configuration (only used by creator)
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

// Mutable variables declared here so they are global
let presetsData = [];

// Card metadata fetched from /api/card-config on load.
// Each entry: { description: string, image: string }
let cardConfig = {};

// Show floating damage numbers when warriors take damage
let screenFlashShown = false;

// Turn Transition Modal
let turnTransitionTimer = null;

// Event Turn Modal (shown to active player when their turn starts and an event is active)
let eventTurnModalTimer = null;

// Stolen Card Notification Modal
let stolenCardTimer = null;

// Spy Notification Modal
let spyNotificationTimer = null;

// Treason Notification Modal — shown to the player whose warrior was stolen
let treasonNotificationTimer = null;
let championsBountyTimer = null;
let resurrectionTimer = null;
let ambushTriggeredTimer = null;
let cardsModalTimer = null;

// Action Confirm Modal Functions
let actionConfirmCallback = null;
