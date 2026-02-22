package types

type WeaponType string

const (
	SwordWeaponType        WeaponType = "Sword"
	ArrowWeaponType        WeaponType = "Arrow"
	PoisonWeaponType       WeaponType = "Poison"
	SpecialPowerWeaponType WeaponType = "Special Power"
	HarpoonWeaponType      WeaponType = "Harpoon"
	BloodRainWeaponType    WeaponType = "Blood Rain"
)

func (wt WeaponType) String() string {
	return string(wt)
}
