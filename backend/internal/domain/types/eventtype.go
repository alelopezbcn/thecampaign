package types

type EventType string

const (
	EventTypeNone            EventType = ""
	EventTypeCurse           EventType = "curse"
	EventTypeHarvest         EventType = "harvest"
	EventTypePlague          EventType = "plague"
	EventTypeAbundance       EventType = "abundance"
	EventTypeBloodlust       EventType = "bloodlust"
	EventTypeChampionsBounty EventType = "champions_bounty"
)

var AllEventTypes = []EventType{
	EventTypeNone,
	EventTypeCurse,
	EventTypeHarvest,
	EventTypePlague,
	EventTypeAbundance,
	EventTypeBloodlust,
	EventTypeChampionsBounty,
}

// CurseWeapons contains the three weapon types that can be affected by the Curse event.
var CurseWeapons = []WeaponType{
	SwordWeaponType,
	ArrowWeaponType,
	PoisonWeaponType,
}

// ActiveEvent holds the current global event and all its randomised parameters.
type ActiveEvent struct {
	Type EventType

	// Curse: one weapon is excluded (the other two are affected by CurseModifier)
	CurseExcludedWeapon WeaponType
	CurseModifier       int // [-3,+3] excl. 0

	// Harvest: flat modifier applied to each resource card's value during construction
	HarvestModifier int // [-4,+4] excl. 0

	// Plague: HP modifier applied to the active player's warriors at turn start
	PlagueModifier int // [-3,+3] excl. 0
}
