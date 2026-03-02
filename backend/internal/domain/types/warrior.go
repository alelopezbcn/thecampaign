package types

type WarriorType string

const (
	ArcherWarriorType    WarriorType = "Archer"
	KnightWarriorType    WarriorType = "Knight"
	MageWarriorType      WarriorType = "Mage"
	DragonWarriorType    WarriorType = "Dragon"
	MercenaryWarriorType WarriorType = "Mercenary"
)

func (wt WarriorType) String() string {
	return string(wt)
}
