package ports

type WarriorType string

const (
	ArcherWarriorType WarriorType = "Archer"
	KnightWarriorType WarriorType = "Knight"
	MageWarriorType   WarriorType = "Mage"
	DragonWarriorType WarriorType = "Dragon"
)

func (wt WarriorType) String() string {
	return string(wt)
}

type WeaponType string

const (
	SwordWeaponType        WeaponType = "Sword"
	ArrowWeaponType        WeaponType = "Arrow"
	PoisonWeaponType       WeaponType = "Poison"
	SpecialPowerWeaponType WeaponType = "Special Power"
)

func (wt WeaponType) String() string {
	return string(wt)
}
