// Maps regular weapon sub_type → warrior sub_types that can wield it.
// Only Sword/Arrow/Poison require attacker selection; BloodRain/Harpoon keep their existing flow.
const WEAPON_ATTACKER_TYPES = {
    'sword':  ['knight', 'dragon', 'mercenary'],
    'arrow':  ['archer', 'dragon', 'mercenary'],
    'poison': ['mage', 'dragon', 'mercenary'],
};

// Warrior types that can use the Special Power card.
const SPECIAL_POWER_USER_TYPES = ['archer', 'knight', 'mage'];

// Ambush effect descriptions and color classes
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
    if (type === 'spy' || type === 'thief' || type === 'catapult' || type === 'fortress' || type === 'resurrection' || type === 'sabotage' || type === 'treason' || type === 'ambush') return 'special';
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

function generateCardID(card) {
    return `card_${Math.random().toString(36).substr(2, 9)}`;
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
