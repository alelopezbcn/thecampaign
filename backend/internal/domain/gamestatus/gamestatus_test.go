package gamestatus

// import (
// 	"testing"

// 	"github.com/alelopezbcn/thecampaign/internal/domain/cards"
// 	"github.com/alelopezbcn/thecampaign/internal/domain/ports"
// 	"github.com/alelopezbcn/thecampaign/internal/domain/types"
// 	"github.com/alelopezbcn/thecampaign/test/mocks"
// 	"github.com/stretchr/testify/assert"
// 	"go.uber.org/mock/gomock"
// )

// func TestGameStatus_WarriorsInHand(t *testing.T) {
// 	ctrl := gomock.NewController(t)
// 	defer ctrl.Finish()

// 	// Create real warrior cards for the hand
// 	k := cards.NewKnight("k1")
// 	a := cards.NewArcher("a1")
// 	m := cards.NewMage("m1")
// 	d := cards.NewDragon("d1")
// 	cardsInHand := []ports.Card{k, a, m, d}

// 	// Setup p1 (current player) mocks
// 	p1 := mocks.NewMockPlayer(ctrl)
// 	p1Hand := mocks.NewMockHand(ctrl)
// 	p1Field := mocks.NewMockField(ctrl)
// 	p1Castle := mocks.NewMockCastle(ctrl)

// 	p1.EXPECT().Name().Return("p1")
// 	p1.EXPECT().Hand().Return(p1Hand)
// 	p1.EXPECT().Field().Return(p1Field)
// 	p1.EXPECT().Castle().Return(p1Castle)

// 	p1Hand.EXPECT().ShowCards().Return(cardsInHand)
// 	p1Field.EXPECT().Warriors().Return([]ports.Warrior{})
// 	p1Castle.EXPECT().IsConstructed().Return(false)
// 	p1Castle.EXPECT().ResourceCards().Return(0)
// 	p1Castle.EXPECT().Value().Return(0)

// 	// Setup p2 (enemy) mocks
// 	p2 := mocks.NewMockPlayer(ctrl)
// 	p2Field := mocks.NewMockField(ctrl)
// 	p2Castle := mocks.NewMockCastle(ctrl)

// 	p2.EXPECT().Field().Return(p2Field)
// 	p2.EXPECT().Castle().Return(p2Castle).Times(2)
// 	p2.EXPECT().CardsInHand().Return(0)

// 	p2Field.EXPECT().Warriors().Return([]ports.Warrior{})
// 	p2Castle.EXPECT().IsConstructed().Return(false)
// 	p2Castle.EXPECT().ResourceCards().Return(0).Times(2)
// 	p2Castle.EXPECT().Value().Return(0)

// 	// Create game status and verify
// 	gameStatus := NewGameStatus(p1, p2, types.ActionTypeSpySteal)

// 	assert.Equal(t, 4, len(gameStatus.CurrentPlayerHand))
// 	assert.True(t, gameStatus.CanMoveWarrior)

// 	// Verify each warrior is in the hand with correct ID and type
// 	handCardIDs := make([]string, len(gameStatus.CurrentPlayerHand))
// 	for i, hc := range gameStatus.CurrentPlayerHand {
// 		handCardIDs[i] = hc.CardID
// 	}
// 	assert.Contains(t, handCardIDs, "K1")
// 	assert.Contains(t, handCardIDs, "A1")
// 	assert.Contains(t, handCardIDs, "M1")
// 	assert.Contains(t, handCardIDs, "D1")
// }

// // Helper function to create base player and castle mocks for weapon tests
// // Note: Does NOT set up Field().Warriors() - each test must do that
// func setupBaseMocks(ctrl *gomock.Controller) (
// 	p1 *mocks.MockPlayer,
// 	p1Hand *mocks.MockHand,
// 	p1Field *mocks.MockField,
// 	p1Castle *mocks.MockCastle,
// 	p2 *mocks.MockPlayer,
// 	p2Field *mocks.MockField,
// 	p2Castle *mocks.MockCastle,
// ) {
// 	p1 = mocks.NewMockPlayer(ctrl)
// 	p1Hand = mocks.NewMockHand(ctrl)
// 	p1Field = mocks.NewMockField(ctrl)
// 	p1Castle = mocks.NewMockCastle(ctrl)

// 	p2 = mocks.NewMockPlayer(ctrl)
// 	p2Field = mocks.NewMockField(ctrl)
// 	p2Castle = mocks.NewMockCastle(ctrl)

// 	// Setup p1 base expectations
// 	p1.EXPECT().Name().Return("p1")
// 	p1.EXPECT().Hand().Return(p1Hand)
// 	p1.EXPECT().Field().Return(p1Field).AnyTimes()
// 	p1.EXPECT().Castle().Return(p1Castle)

// 	// Note: p1Field.Warriors() is NOT set here - tests must set it
// 	p1Castle.EXPECT().IsConstructed().Return(false)
// 	p1Castle.EXPECT().ResourceCards().Return(0)
// 	p1Castle.EXPECT().Value().Return(0)

// 	// Setup p2 base expectations
// 	p2.EXPECT().Field().Return(p2Field).AnyTimes()
// 	p2.EXPECT().Castle().Return(p2Castle).Times(2)
// 	p2.EXPECT().CardsInHand().Return(0)

// 	// Note: p2Field.Warriors() is NOT set here - tests must set it
// 	p2Castle.EXPECT().IsConstructed().Return(false)
// 	p2Castle.EXPECT().ResourceCards().Return(0).Times(2)
// 	p2Castle.EXPECT().Value().Return(0)

// 	return
// }

// /*
// func TestGameStatus_SwordInHand_WithKnight_CanAttack(t *testing.T) {
// 	ctrl := gomock.NewController(t)
// 	defer ctrl.Finish()

// 	p1, p1Hand, p1Field, _, p2, p2Field, _ := setupBaseMocks(ctrl)

// 	// Create sword weapon mock
// 	sword := mocks.NewMockWeapon(ctrl)
// 	sword.EXPECT().CanConstruct().Return(false).Times(2) // Called in gamestatus.go and handcard.go
// 	sword.EXPECT().Type().Return(ports.SwordWeaponType).Times(2)
// 	sword.EXPECT().GetID().Return("S1")
// 	sword.EXPECT().DamageAmount().Return(5)

// 	p1Hand.EXPECT().ShowCards().Return([]ports.Card{sword})

// 	// Knight on field - sword is usable (HasDragon not called due to short-circuit)
// 	p1Field.EXPECT().HasKnight().Return(true)
// 	p1Field.EXPECT().Warriors().Return([]ports.Warrior{})

// 	// Enemy has attackable targets
// 	p2Field.EXPECT().AttackableIDs().Return([]string{"EK1", "EA1"})
// 	p2Field.EXPECT().Warriors().Return([]ports.Warrior{})

// 	gameStatus := NewGameStatus(p1, p2)

// 	assert.True(t, gameStatus.CanAttack)
// 	assert.Equal(t, 1, len(gameStatus.CurrentPlayerHand))
// 	assert.Equal(t, "S1", gameStatus.CurrentPlayerHand[0].CardID)
// 	assert.Equal(t, CardTypeSword, gameStatus.CurrentPlayerHand[0].CardType)
// 	assert.Equal(t, 5, gameStatus.CurrentPlayerHand[0].Value)
// 	assert.Equal(t, []string{"EK1", "EA1"}, gameStatus.CurrentPlayerHand[0].CanBeUsedOnIDs)
// }

// func TestGameStatus_ArrowInHand_WithArcher_CanAttack(t *testing.T) {
// 	ctrl := gomock.NewController(t)
// 	defer ctrl.Finish()

// 	p1, p1Hand, p1Field, _, p2, p2Field, _ := setupBaseMocks(ctrl)

// 	// Create arrow weapon mock
// 	arrow := mocks.NewMockWeapon(ctrl)
// 	arrow.EXPECT().CanConstruct().Return(false).Times(2)
// 	arrow.EXPECT().Type().Return(ports.ArrowWeaponType).Times(2)
// 	arrow.EXPECT().GetID().Return("A1")
// 	arrow.EXPECT().DamageAmount().Return(3)

// 	p1Hand.EXPECT().ShowCards().Return([]ports.Card{arrow})

// 	// Archer on field - arrow is usable (HasDragon not called due to short-circuit)
// 	p1Field.EXPECT().HasArcher().Return(true)
// 	p1Field.EXPECT().Warriors().Return([]ports.Warrior{})

// 	// Enemy has attackable targets
// 	p2Field.EXPECT().AttackableIDs().Return([]string{"EK1"})
// 	p2Field.EXPECT().Warriors().Return([]ports.Warrior{})

// 	gameStatus := NewGameStatus(p1, p2)

// 	assert.True(t, gameStatus.CanAttack)
// 	assert.Equal(t, 1, len(gameStatus.CurrentPlayerHand))
// 	assert.Equal(t, "A1", gameStatus.CurrentPlayerHand[0].CardID)
// 	assert.Equal(t, CardTypeArrow, gameStatus.CurrentPlayerHand[0].CardType)
// 	assert.Equal(t, []string{"EK1"}, gameStatus.CurrentPlayerHand[0].CanBeUsedOnIDs)
// }

// func TestGameStatus_PoisonInHand_WithMage_CanAttack(t *testing.T) {
// 	ctrl := gomock.NewController(t)
// 	defer ctrl.Finish()

// 	p1, p1Hand, p1Field, _, p2, p2Field, _ := setupBaseMocks(ctrl)

// 	// Create poison weapon mock
// 	poison := mocks.NewMockWeapon(ctrl)
// 	poison.EXPECT().CanConstruct().Return(false).Times(2)
// 	poison.EXPECT().Type().Return(ports.PoisonWeaponType).Times(2)
// 	poison.EXPECT().GetID().Return("P1")
// 	poison.EXPECT().DamageAmount().Return(4)

// 	p1Hand.EXPECT().ShowCards().Return([]ports.Card{poison})

// 	// Mage on field - poison is usable (HasDragon not called due to short-circuit)
// 	p1Field.EXPECT().HasMage().Return(true)
// 	p1Field.EXPECT().Warriors().Return([]ports.Warrior{})

// 	// Enemy has attackable targets
// 	p2Field.EXPECT().AttackableIDs().Return([]string{"EM1"})
// 	p2Field.EXPECT().Warriors().Return([]ports.Warrior{})

// 	gameStatus := NewGameStatus(p1, p2)

// 	assert.True(t, gameStatus.CanAttack)
// 	assert.Equal(t, 1, len(gameStatus.CurrentPlayerHand))
// 	assert.Equal(t, "P1", gameStatus.CurrentPlayerHand[0].CardID)
// 	assert.Equal(t, CardTypePoison, gameStatus.CurrentPlayerHand[0].CardType)
// 	assert.Equal(t, []string{"EM1"}, gameStatus.CurrentPlayerHand[0].CanBeUsedOnIDs)
// }

// func TestGameStatus_SwordInHand_NoMatchingWarrior_CannotUse(t *testing.T) {
// 	ctrl := gomock.NewController(t)
// 	defer ctrl.Finish()

// 	p1, p1Hand, p1Field, _, p2, p2Field, _ := setupBaseMocks(ctrl)

// 	// Create sword weapon mock
// 	sword := mocks.NewMockWeapon(ctrl)
// 	sword.EXPECT().CanConstruct().Return(false).Times(2)
// 	sword.EXPECT().Type().Return(ports.SwordWeaponType).Times(2)
// 	sword.EXPECT().GetID().Return("S1")
// 	sword.EXPECT().DamageAmount().Return(5)

// 	p1Hand.EXPECT().ShowCards().Return([]ports.Card{sword})

// 	// No knight or dragon on field - sword not usable
// 	p1Field.EXPECT().HasKnight().Return(false)
// 	p1Field.EXPECT().HasDragon().Return(false)
// 	p1Field.EXPECT().Warriors().Return([]ports.Warrior{})

// 	// Even if enemy has targets, sword can't be used
// 	p2Field.EXPECT().AttackableIDs().Return([]string{"EK1"})
// 	p2Field.EXPECT().Warriors().Return([]ports.Warrior{})

// 	gameStatus := NewGameStatus(p1, p2)

// 	assert.True(t, gameStatus.CanAttack) // CanAttack is still true (weapon exists)
// 	assert.Equal(t, 1, len(gameStatus.CurrentPlayerHand))
// 	assert.Equal(t, "S1", gameStatus.CurrentPlayerHand[0].CardID)
// 	assert.Empty(t, gameStatus.CurrentPlayerHand[0].CanBeUsedOnIDs) // No targets since weapon not usable
// }

// func TestGameStatus_WeaponInHand_WithDragon_AllUsable(t *testing.T) {
// 	ctrl := gomock.NewController(t)
// 	defer ctrl.Finish()

// 	p1, p1Hand, p1Field, _, p2, p2Field, _ := setupBaseMocks(ctrl)

// 	// Create all weapon types
// 	sword := mocks.NewMockWeapon(ctrl)
// 	sword.EXPECT().CanConstruct().Return(false).Times(2)
// 	sword.EXPECT().Type().Return(ports.SwordWeaponType).Times(2)
// 	sword.EXPECT().GetID().Return("S1")
// 	sword.EXPECT().DamageAmount().Return(5)

// 	arrow := mocks.NewMockWeapon(ctrl)
// 	arrow.EXPECT().CanConstruct().Return(false).Times(2)
// 	arrow.EXPECT().Type().Return(ports.ArrowWeaponType).Times(2)
// 	arrow.EXPECT().GetID().Return("A1")
// 	arrow.EXPECT().DamageAmount().Return(3)

// 	poison := mocks.NewMockWeapon(ctrl)
// 	poison.EXPECT().CanConstruct().Return(false).Times(2)
// 	poison.EXPECT().Type().Return(ports.PoisonWeaponType).Times(2)
// 	poison.EXPECT().GetID().Return("P1")
// 	poison.EXPECT().DamageAmount().Return(4)

// 	p1Hand.EXPECT().ShowCards().Return([]ports.Card{sword, arrow, poison})

// 	// Dragon on field - all weapons are usable
// 	p1Field.EXPECT().HasKnight().Return(false)
// 	p1Field.EXPECT().HasDragon().Return(true).Times(3)
// 	p1Field.EXPECT().HasArcher().Return(false)
// 	p1Field.EXPECT().HasMage().Return(false)
// 	p1Field.EXPECT().Warriors().Return([]ports.Warrior{})

// 	// Enemy has attackable targets
// 	p2Field.EXPECT().AttackableIDs().Return([]string{"EK1"}).Times(3)
// 	p2Field.EXPECT().Warriors().Return([]ports.Warrior{})

// 	gameStatus := NewGameStatus(p1, p2)

// 	assert.True(t, gameStatus.CanAttack)
// 	assert.Equal(t, 3, len(gameStatus.CurrentPlayerHand))

// 	// All weapons should have targets
// 	for _, hc := range gameStatus.CurrentPlayerHand {
// 		assert.Equal(t, []string{"EK1"}, hc.CanBeUsedOnIDs)
// 	}
// }

// func TestGameStatus_WeaponInHand_CanConstruct_True(t *testing.T) {
// 	ctrl := gomock.NewController(t)
// 	defer ctrl.Finish()

// 	p1, p1Hand, p1Field, _, p2, p2Field, _ := setupBaseMocks(ctrl)

// 	// Create weapon that can construct (value 1)
// 	sword := mocks.NewMockWeapon(ctrl)
// 	sword.EXPECT().CanConstruct().Return(true).Times(2)
// 	sword.EXPECT().Type().Return(ports.SwordWeaponType).Times(2)
// 	sword.EXPECT().GetID().Return("S1")
// 	sword.EXPECT().DamageAmount().Return(1)

// 	p1Hand.EXPECT().ShowCards().Return([]ports.Card{sword})

// 	p1Field.EXPECT().HasKnight().Return(true)
// 	p1Field.EXPECT().Warriors().Return([]ports.Warrior{})

// 	p2Field.EXPECT().AttackableIDs().Return([]string{})
// 	p2Field.EXPECT().Warriors().Return([]ports.Warrior{})

// 	gameStatus := NewGameStatus(p1, p2)

// 	assert.True(t, gameStatus.CanInitiateCastle)
// 	assert.True(t, gameStatus.CurrentPlayerHand[0].CanConstruct)
// }

// func TestGameStatus_WeaponInHand_CanConstruct_False(t *testing.T) {
// 	ctrl := gomock.NewController(t)
// 	defer ctrl.Finish()

// 	p1, p1Hand, p1Field, _, p2, p2Field, _ := setupBaseMocks(ctrl)

// 	// Create weapon that cannot construct (value > 1)
// 	sword := mocks.NewMockWeapon(ctrl)
// 	sword.EXPECT().CanConstruct().Return(false).Times(2)
// 	sword.EXPECT().Type().Return(ports.SwordWeaponType).Times(2)
// 	sword.EXPECT().GetID().Return("S1")
// 	sword.EXPECT().DamageAmount().Return(5)

// 	p1Hand.EXPECT().ShowCards().Return([]ports.Card{sword})

// 	p1Field.EXPECT().HasKnight().Return(true)
// 	p1Field.EXPECT().Warriors().Return([]ports.Warrior{})

// 	p2Field.EXPECT().AttackableIDs().Return([]string{})
// 	p2Field.EXPECT().Warriors().Return([]ports.Warrior{})

// 	gameStatus := NewGameStatus(p1, p2)

// 	assert.False(t, gameStatus.CanInitiateCastle)
// 	assert.False(t, gameStatus.CurrentPlayerHand[0].CanConstruct)
// }

// func TestGameStatus_SpecialPowerInHand_WithArcher_CanInstantKill(t *testing.T) {
// 	ctrl := gomock.NewController(t)
// 	defer ctrl.Finish()

// 	p1, p1Hand, p1Field, _, p2, p2Field, _ := setupBaseMocks(ctrl)

// 	// Create special power mock
// 	sp := mocks.NewMockSpecialPower(ctrl)
// 	sp.EXPECT().CanConstruct().Return(false)
// 	sp.EXPECT().Type().Return(ports.SpecialPowerWeaponType)
// 	sp.EXPECT().GetID().Return("SP1")

// 	p1Hand.EXPECT().ShowCards().Return([]ports.Card{sp})

// 	// Archer on field - can instant kill
// 	p1Field.EXPECT().HasArcher().Return(true)
// 	p1Field.EXPECT().HasKnight().Return(false)
// 	p1Field.EXPECT().HasMage().Return(false)
// 	p1Field.EXPECT().Warriors().Return([]ports.Warrior{})

// 	// Enemy warrior (unprotected) - needs all methods for newFieldCard
// 	enemyWarrior := mocks.NewMockWarrior(ctrl)
// 	enemyWarrior.EXPECT().IsProtected().Return(false, nil).Times(2) // Once for handcard, once for fieldcard
// 	enemyWarrior.EXPECT().GetID().Return("EK1").Times(2)            // Once for handcard, once for fieldcard
// 	enemyWarrior.EXPECT().Type().Return(ports.KnightWarriorType)    // For fieldcard
// 	enemyWarrior.EXPECT().AttackedBy().Return([]ports.Weapon{})     // For fieldcard
// 	enemyWarrior.EXPECT().Health().Return(10)                       // For fieldcard

// 	p2Field.EXPECT().Warriors().Return([]ports.Warrior{enemyWarrior}).Times(2)

// 	gameStatus := NewGameStatus(p1, p2)

// 	assert.True(t, gameStatus.CanAttack)
// 	assert.Equal(t, 1, len(gameStatus.CurrentPlayerHand))
// 	assert.Equal(t, "SP1", gameStatus.CurrentPlayerHand[0].CardID)
// 	assert.Equal(t, CardTypeSpecialPower, gameStatus.CurrentPlayerHand[0].CardType)
// 	assert.Contains(t, gameStatus.CurrentPlayerHand[0].CanBeUsedOnIDs, "EK1")
// }

// func TestGameStatus_SpecialPowerInHand_WithKnight_CanProtect(t *testing.T) {
// 	ctrl := gomock.NewController(t)
// 	defer ctrl.Finish()

// 	p1, p1Hand, p1Field, _, p2, p2Field, _ := setupBaseMocks(ctrl)

// 	// Create special power mock
// 	sp := mocks.NewMockSpecialPower(ctrl)
// 	sp.EXPECT().CanConstruct().Return(false)
// 	sp.EXPECT().Type().Return(ports.SpecialPowerWeaponType)
// 	sp.EXPECT().GetID().Return("SP1")

// 	p1Hand.EXPECT().ShowCards().Return([]ports.Card{sp})

// 	// Knight on field - can protect
// 	p1Field.EXPECT().HasArcher().Return(false)
// 	p1Field.EXPECT().HasKnight().Return(true)
// 	p1Field.EXPECT().HasMage().Return(false)

// 	// My warrior (knight, unprotected) - needs all methods for newFieldCard
// 	myWarrior := mocks.NewMockWarrior(ctrl)
// 	myWarrior.EXPECT().IsProtected().Return(false, nil).Times(2) // Once for handcard, once for fieldcard
// 	myWarrior.EXPECT().Type().Return(ports.KnightWarriorType).Times(2)
// 	myWarrior.EXPECT().GetID().Return("K1").Times(2)
// 	myWarrior.EXPECT().AttackedBy().Return([]ports.Weapon{})
// 	myWarrior.EXPECT().Health().Return(10)

// 	p1Field.EXPECT().Warriors().Return([]ports.Warrior{myWarrior}).Times(2)

// 	p2Field.EXPECT().Warriors().Return([]ports.Warrior{})

// 	gameStatus := NewGameStatus(p1, p2)

// 	assert.True(t, gameStatus.CanAttack)
// 	assert.Equal(t, 1, len(gameStatus.CurrentPlayerHand))
// 	assert.Contains(t, gameStatus.CurrentPlayerHand[0].CanBeUsedOnIDs, "K1")
// }

// func TestGameStatus_SpecialPowerInHand_WithMage_CanHeal(t *testing.T) {
// 	ctrl := gomock.NewController(t)
// 	defer ctrl.Finish()

// 	p1, p1Hand, p1Field, _, p2, p2Field, _ := setupBaseMocks(ctrl)

// 	// Create special power mock
// 	sp := mocks.NewMockSpecialPower(ctrl)
// 	sp.EXPECT().CanConstruct().Return(false)
// 	sp.EXPECT().Type().Return(ports.SpecialPowerWeaponType)
// 	sp.EXPECT().GetID().Return("SP1")

// 	p1Hand.EXPECT().ShowCards().Return([]ports.Card{sp})

// 	// Mage on field - can heal
// 	p1Field.EXPECT().HasArcher().Return(false)
// 	p1Field.EXPECT().HasKnight().Return(false)
// 	p1Field.EXPECT().HasMage().Return(true)

// 	// My damaged warrior (not dragon) - needs all methods for newFieldCard
// 	myWarrior := mocks.NewMockWarrior(ctrl)
// 	myWarrior.EXPECT().Type().Return(ports.KnightWarriorType).Times(2)
// 	myWarrior.EXPECT().IsDamaged().Return(true)
// 	myWarrior.EXPECT().GetID().Return("K1").Times(2)
// 	myWarrior.EXPECT().IsProtected().Return(false, nil)
// 	myWarrior.EXPECT().AttackedBy().Return([]ports.Weapon{})
// 	myWarrior.EXPECT().Health().Return(5) // Damaged so less than 10

// 	p1Field.EXPECT().Warriors().Return([]ports.Warrior{myWarrior}).Times(2)

// 	p2Field.EXPECT().Warriors().Return([]ports.Warrior{})

// 	gameStatus := NewGameStatus(p1, p2)

// 	assert.True(t, gameStatus.CanAttack)
// 	assert.Equal(t, 1, len(gameStatus.CurrentPlayerHand))
// 	assert.Contains(t, gameStatus.CurrentPlayerHand[0].CanBeUsedOnIDs, "K1")
// }

// func TestGameStatus_SpecialPowerInHand_ProtectedEnemy_TargetsShield(t *testing.T) {
// 	ctrl := gomock.NewController(t)
// 	defer ctrl.Finish()

// 	p1, p1Hand, p1Field, _, p2, p2Field, _ := setupBaseMocks(ctrl)

// 	// Create special power mock
// 	sp := mocks.NewMockSpecialPower(ctrl)
// 	sp.EXPECT().CanConstruct().Return(false)
// 	sp.EXPECT().Type().Return(ports.SpecialPowerWeaponType)
// 	sp.EXPECT().GetID().Return("SP1")

// 	p1Hand.EXPECT().ShowCards().Return([]ports.Card{sp})

// 	// Archer on field - can instant kill
// 	p1Field.EXPECT().HasArcher().Return(true)
// 	p1Field.EXPECT().HasKnight().Return(false)
// 	p1Field.EXPECT().HasMage().Return(false)
// 	p1Field.EXPECT().Warriors().Return([]ports.Warrior{})

// 	// Enemy warrior protected by shield - needs all methods for newFieldCard
// 	shield := mocks.NewMockSpecialPower(ctrl)
// 	shield.EXPECT().GetID().Return("SHIELD1").Times(2) // Once for handcard, once for fieldcard

// 	enemyWarrior := mocks.NewMockWarrior(ctrl)
// 	enemyWarrior.EXPECT().IsProtected().Return(true, shield).Times(2)
// 	enemyWarrior.EXPECT().Type().Return(ports.KnightWarriorType)
// 	enemyWarrior.EXPECT().GetID().Return("EK1")
// 	enemyWarrior.EXPECT().AttackedBy().Return([]ports.Weapon{})
// 	enemyWarrior.EXPECT().Health().Return(10)

// 	p2Field.EXPECT().Warriors().Return([]ports.Warrior{enemyWarrior}).Times(2)

// 	gameStatus := NewGameStatus(p1, p2)

// 	assert.True(t, gameStatus.CanAttack)
// 	assert.Equal(t, 1, len(gameStatus.CurrentPlayerHand))
// 	// Should target the shield, not the warrior
// 	assert.Contains(t, gameStatus.CurrentPlayerHand[0].CanBeUsedOnIDs, "SHIELD1")
// }

// // ========== Catapult Tests (line 84-92) ==========

// func TestGameStatus_CatapultInHand_EnemyCastleCanBeAttacked(t *testing.T) {
// 	ctrl := gomock.NewController(t)
// 	defer ctrl.Finish()

// 	// Setup p1 mocks
// 	p1 := mocks.NewMockPlayer(ctrl)
// 	p1Hand := mocks.NewMockHand(ctrl)
// 	p1Field := mocks.NewMockField(ctrl)
// 	p1Castle := mocks.NewMockCastle(ctrl)

// 	p1.EXPECT().Name().Return("p1")
// 	p1.EXPECT().Hand().Return(p1Hand)
// 	p1.EXPECT().Field().Return(p1Field)
// 	p1.EXPECT().Castle().Return(p1Castle)

// 	p1Castle.EXPECT().IsConstructed().Return(false)
// 	p1Castle.EXPECT().ResourceCards().Return(0)
// 	p1Castle.EXPECT().Value().Return(0)

// 	// Setup p2 mocks - note extra Castle() calls for catapult
// 	p2 := mocks.NewMockPlayer(ctrl)
// 	p2Field := mocks.NewMockField(ctrl)
// 	p2Castle := mocks.NewMockCastle(ctrl)

// 	p2.EXPECT().Field().Return(p2Field)
// 	p2.EXPECT().Castle().Return(p2Castle).Times(4) // CanBeAttacked, GetID, newCastle, ResourceCards
// 	p2.EXPECT().CardsInHand().Return(0)

// 	// Create catapult mock
// 	catapult := mocks.NewMockCatapult(ctrl)
// 	catapult.EXPECT().GetID().Return("CAT1")

// 	p1Hand.EXPECT().ShowCards().Return([]ports.Card{catapult})

// 	// Enemy castle can be attacked
// 	p2Castle.EXPECT().CanBeAttacked().Return(true)
// 	p2Castle.EXPECT().GetID().Return("CASTLE1")
// 	p2Castle.EXPECT().IsConstructed().Return(true)
// 	p2Castle.EXPECT().ResourceCards().Return(3).Times(2)
// 	p2Castle.EXPECT().Value().Return(5)

// 	p1Field.EXPECT().Warriors().Return([]ports.Warrior{})
// 	p2Field.EXPECT().Warriors().Return([]ports.Warrior{})

// 	gameStatus := NewGameStatus(p1, p2)

// 	assert.True(t, gameStatus.CanAttack)
// 	assert.Equal(t, 1, len(gameStatus.CurrentPlayerHand))
// 	assert.Equal(t, "CAT1", gameStatus.CurrentPlayerHand[0].CardID)
// 	assert.Equal(t, CardTypeCatapult, gameStatus.CurrentPlayerHand[0].CardType)
// 	assert.Equal(t, []string{"CASTLE1"}, gameStatus.CurrentPlayerHand[0].CanBeUsedOnIDs)
// }

// func TestGameStatus_CatapultInHand_EnemyCastleCannotBeAttacked(t *testing.T) {
// 	ctrl := gomock.NewController(t)
// 	defer ctrl.Finish()

// 	// Setup p1 mocks
// 	p1 := mocks.NewMockPlayer(ctrl)
// 	p1Hand := mocks.NewMockHand(ctrl)
// 	p1Field := mocks.NewMockField(ctrl)
// 	p1Castle := mocks.NewMockCastle(ctrl)

// 	p1.EXPECT().Name().Return("p1")
// 	p1.EXPECT().Hand().Return(p1Hand)
// 	p1.EXPECT().Field().Return(p1Field)
// 	p1.EXPECT().Castle().Return(p1Castle)

// 	p1Castle.EXPECT().IsConstructed().Return(false)
// 	p1Castle.EXPECT().ResourceCards().Return(0)
// 	p1Castle.EXPECT().Value().Return(0)

// 	// Setup p2 mocks - note extra Castle() call for catapult
// 	p2 := mocks.NewMockPlayer(ctrl)
// 	p2Field := mocks.NewMockField(ctrl)
// 	p2Castle := mocks.NewMockCastle(ctrl)

// 	p2.EXPECT().Field().Return(p2Field)
// 	p2.EXPECT().Castle().Return(p2Castle).Times(3) // CanBeAttacked, newCastle, ResourceCards
// 	p2.EXPECT().CardsInHand().Return(0)

// 	// Create catapult mock
// 	catapult := mocks.NewMockCatapult(ctrl)
// 	catapult.EXPECT().GetID().Return("CAT1")

// 	p1Hand.EXPECT().ShowCards().Return([]ports.Card{catapult})

// 	// Enemy castle cannot be attacked
// 	p2Castle.EXPECT().CanBeAttacked().Return(false)
// 	p2Castle.EXPECT().IsConstructed().Return(false)
// 	p2Castle.EXPECT().ResourceCards().Return(0).Times(2)
// 	p2Castle.EXPECT().Value().Return(0)

// 	p1Field.EXPECT().Warriors().Return([]ports.Warrior{})
// 	p2Field.EXPECT().Warriors().Return([]ports.Warrior{})

// 	gameStatus := NewGameStatus(p1, p2)

// 	assert.False(t, gameStatus.CanAttack) // No weapon and catapult can't be used
// 	assert.Equal(t, 1, len(gameStatus.CurrentPlayerHand))
// 	assert.Equal(t, "CAT1", gameStatus.CurrentPlayerHand[0].CardID)
// 	assert.Equal(t, CardTypeCatapult, gameStatus.CurrentPlayerHand[0].CardType)
// 	assert.Empty(t, gameStatus.CurrentPlayerHand[0].CanBeUsedOnIDs) // No target
// }

// func TestGameStatus_CatapultInHand_CanAttackAlreadyTrue(t *testing.T) {
// 	ctrl := gomock.NewController(t)
// 	defer ctrl.Finish()

// 	// Setup p1 mocks
// 	p1 := mocks.NewMockPlayer(ctrl)
// 	p1Hand := mocks.NewMockHand(ctrl)
// 	p1Field := mocks.NewMockField(ctrl)
// 	p1Castle := mocks.NewMockCastle(ctrl)

// 	p1.EXPECT().Name().Return("p1")
// 	p1.EXPECT().Hand().Return(p1Hand)
// 	p1.EXPECT().Field().Return(p1Field).AnyTimes()
// 	p1.EXPECT().Castle().Return(p1Castle)

// 	p1Castle.EXPECT().IsConstructed().Return(false)
// 	p1Castle.EXPECT().ResourceCards().Return(0)
// 	p1Castle.EXPECT().Value().Return(0)

// 	// Setup p2 mocks
// 	p2 := mocks.NewMockPlayer(ctrl)
// 	p2Field := mocks.NewMockField(ctrl)
// 	p2Castle := mocks.NewMockCastle(ctrl)

// 	p2.EXPECT().Field().Return(p2Field).AnyTimes()
// 	p2.EXPECT().Castle().Return(p2Castle).Times(3) // CanBeAttacked, newCastle, ResourceCards
// 	p2.EXPECT().CardsInHand().Return(0)

// 	// Create a weapon first (sets CanAttack to true)
// 	sword := mocks.NewMockWeapon(ctrl)
// 	sword.EXPECT().CanConstruct().Return(false).Times(2)
// 	sword.EXPECT().Type().Return(ports.SwordWeaponType).Times(2)
// 	sword.EXPECT().GetID().Return("S1")
// 	sword.EXPECT().DamageAmount().Return(5)

// 	// Create catapult
// 	catapult := mocks.NewMockCatapult(ctrl)
// 	catapult.EXPECT().GetID().Return("CAT1")

// 	p1Hand.EXPECT().ShowCards().Return([]ports.Card{sword, catapult})

// 	// Sword is usable (knight on field)
// 	p1Field.EXPECT().HasKnight().Return(true)
// 	p1Field.EXPECT().Warriors().Return([]ports.Warrior{})

// 	p2Field.EXPECT().AttackableIDs().Return([]string{"EK1"})
// 	p2Field.EXPECT().Warriors().Return([]ports.Warrior{})

// 	// Enemy castle cannot be attacked, but CanAttack should remain true from sword
// 	p2Castle.EXPECT().CanBeAttacked().Return(false)
// 	p2Castle.EXPECT().IsConstructed().Return(false)
// 	p2Castle.EXPECT().ResourceCards().Return(0).Times(2)
// 	p2Castle.EXPECT().Value().Return(0)

// 	gameStatus := NewGameStatus(p1, p2)

// 	assert.True(t, gameStatus.CanAttack) // Still true because of the sword
// 	assert.Equal(t, 2, len(gameStatus.CurrentPlayerHand))

// 	// Find catapult in hand and verify it has no target
// 	var catapultCard *HandCard
// 	for i := range gameStatus.CurrentPlayerHand {
// 		if gameStatus.CurrentPlayerHand[i].CardType == CardTypeCatapult {
// 			catapultCard = &gameStatus.CurrentPlayerHand[i]
// 			break
// 		}
// 	}
// 	assert.NotNil(t, catapultCard)
// 	assert.Empty(t, catapultCard.CanBeUsedOnIDs)
// }

// // ========== Spy Tests (line 92-95) ==========

// func TestGameStatus_SpyInHand_SetsCanSpy(t *testing.T) {
// 	ctrl := gomock.NewController(t)
// 	defer ctrl.Finish()

// 	p1, p1Hand, p1Field, _, p2, p2Field, _ := setupBaseMocks(ctrl)

// 	// Create spy mock
// 	spy := mocks.NewMockSpy(ctrl)
// 	spy.EXPECT().GetID().Return("SPY1")

// 	p1Hand.EXPECT().ShowCards().Return([]ports.Card{spy})

// 	p1Field.EXPECT().Warriors().Return([]ports.Warrior{})
// 	p2Field.EXPECT().Warriors().Return([]ports.Warrior{})

// 	gameStatus := NewGameStatus(p1, p2)

// 	assert.True(t, gameStatus.CanSpy)
// 	assert.Equal(t, 1, len(gameStatus.CurrentPlayerHand))
// 	assert.Equal(t, "SPY1", gameStatus.CurrentPlayerHand[0].CardID)
// 	assert.Equal(t, CardTypeSpy, gameStatus.CurrentPlayerHand[0].CardType)
// }

// func TestGameStatus_MultipleSpiesInHand(t *testing.T) {
// 	ctrl := gomock.NewController(t)
// 	defer ctrl.Finish()

// 	p1, p1Hand, p1Field, _, p2, p2Field, _ := setupBaseMocks(ctrl)

// 	// Create multiple spy mocks
// 	spy1 := mocks.NewMockSpy(ctrl)
// 	spy1.EXPECT().GetID().Return("SPY1")

// 	spy2 := mocks.NewMockSpy(ctrl)
// 	spy2.EXPECT().GetID().Return("SPY2")

// 	p1Hand.EXPECT().ShowCards().Return([]ports.Card{spy1, spy2})

// 	p1Field.EXPECT().Warriors().Return([]ports.Warrior{})
// 	p2Field.EXPECT().Warriors().Return([]ports.Warrior{})

// 	gameStatus := NewGameStatus(p1, p2)

// 	assert.True(t, gameStatus.CanSpy)
// 	assert.Equal(t, 2, len(gameStatus.CurrentPlayerHand))
// }

// // ========== Thief Tests (line 96-99) ==========

// func TestGameStatus_ThiefInHand_SetsCanSteal(t *testing.T) {
// 	ctrl := gomock.NewController(t)
// 	defer ctrl.Finish()

// 	p1, p1Hand, p1Field, _, p2, p2Field, _ := setupBaseMocks(ctrl)

// 	// Create thief mock
// 	thief := mocks.NewMockThief(ctrl)
// 	thief.EXPECT().GetID().Return("THIEF1")

// 	p1Hand.EXPECT().ShowCards().Return([]ports.Card{thief})

// 	p1Field.EXPECT().Warriors().Return([]ports.Warrior{})
// 	p2Field.EXPECT().Warriors().Return([]ports.Warrior{})

// 	gameStatus := NewGameStatus(p1, p2)

// 	assert.True(t, gameStatus.CanSteal)
// 	assert.Equal(t, 1, len(gameStatus.CurrentPlayerHand))
// 	assert.Equal(t, "THIEF1", gameStatus.CurrentPlayerHand[0].CardID)
// 	assert.Equal(t, CardTypeThief, gameStatus.CurrentPlayerHand[0].CardType)
// }

// func TestGameStatus_MultipleThievesInHand(t *testing.T) {
// 	ctrl := gomock.NewController(t)
// 	defer ctrl.Finish()

// 	p1, p1Hand, p1Field, _, p2, p2Field, _ := setupBaseMocks(ctrl)

// 	// Create multiple thief mocks
// 	thief1 := mocks.NewMockThief(ctrl)
// 	thief1.EXPECT().GetID().Return("THIEF1")

// 	thief2 := mocks.NewMockThief(ctrl)
// 	thief2.EXPECT().GetID().Return("THIEF2")

// 	p1Hand.EXPECT().ShowCards().Return([]ports.Card{thief1, thief2})

// 	p1Field.EXPECT().Warriors().Return([]ports.Warrior{})
// 	p2Field.EXPECT().Warriors().Return([]ports.Warrior{})

// 	gameStatus := NewGameStatus(p1, p2)

// 	assert.True(t, gameStatus.CanSteal)
// 	assert.Equal(t, 2, len(gameStatus.CurrentPlayerHand))
// }

// func TestGameStatus_SpyAndThiefInHand(t *testing.T) {
// 	ctrl := gomock.NewController(t)
// 	defer ctrl.Finish()

// 	p1, p1Hand, p1Field, _, p2, p2Field, _ := setupBaseMocks(ctrl)

// 	// Create spy and thief mocks
// 	spy := mocks.NewMockSpy(ctrl)
// 	spy.EXPECT().GetID().Return("SPY1")

// 	thief := mocks.NewMockThief(ctrl)
// 	thief.EXPECT().GetID().Return("THIEF1")

// 	p1Hand.EXPECT().ShowCards().Return([]ports.Card{spy, thief})

// 	p1Field.EXPECT().Warriors().Return([]ports.Warrior{})
// 	p2Field.EXPECT().Warriors().Return([]ports.Warrior{})

// 	gameStatus := NewGameStatus(p1, p2)

// 	assert.True(t, gameStatus.CanSpy)
// 	assert.True(t, gameStatus.CanSteal)
// 	assert.Equal(t, 2, len(gameStatus.CurrentPlayerHand))

// 	// Verify both card types are present
// 	cardTypes := make(map[CardType]bool)
// 	for _, hc := range gameStatus.CurrentPlayerHand {
// 		cardTypes[hc.CardType] = true
// 	}
// 	assert.True(t, cardTypes[CardTypeSpy])
// 	assert.True(t, cardTypes[CardTypeThief])
// }

// // ========== Resource Tests (line 100-106) ==========

// func TestGameStatus_ResourceInHand_CanConstruct_SetsCanInitiateCastle(t *testing.T) {
// 	ctrl := gomock.NewController(t)
// 	defer ctrl.Finish()

// 	p1, p1Hand, p1Field, _, p2, p2Field, _ := setupBaseMocks(ctrl)

// 	// Create resource mock that can construct (value 1)
// 	resource := mocks.NewMockResource(ctrl)
// 	resource.EXPECT().CanConstruct().Return(true).Times(2) // Once in gamestatus.go, once in handcard.go
// 	resource.EXPECT().CanBuy().Return(false)
// 	resource.EXPECT().GetID().Return("G1")
// 	resource.EXPECT().Value().Return(1)

// 	p1Hand.EXPECT().ShowCards().Return([]ports.Card{resource})

// 	p1Field.EXPECT().Warriors().Return([]ports.Warrior{})
// 	p2Field.EXPECT().Warriors().Return([]ports.Warrior{})

// 	gameStatus := NewGameStatus(p1, p2)

// 	assert.True(t, gameStatus.CanInitiateCastle)
// 	assert.True(t, gameStatus.CanGrowCastle)
// 	assert.False(t, gameStatus.CanBuy)
// 	assert.Equal(t, 1, len(gameStatus.CurrentPlayerHand))
// 	assert.Equal(t, "G1", gameStatus.CurrentPlayerHand[0].CardID)
// 	assert.Equal(t, CardTypeResource, gameStatus.CurrentPlayerHand[0].CardType)
// 	assert.True(t, gameStatus.CurrentPlayerHand[0].CanConstruct)
// }

// func TestGameStatus_ResourceInHand_CannotConstruct(t *testing.T) {
// 	ctrl := gomock.NewController(t)
// 	defer ctrl.Finish()

// 	p1, p1Hand, p1Field, _, p2, p2Field, _ := setupBaseMocks(ctrl)

// 	// Create resource mock that cannot construct (value > 1)
// 	resource := mocks.NewMockResource(ctrl)
// 	resource.EXPECT().CanConstruct().Return(false).Times(2)
// 	resource.EXPECT().CanBuy().Return(true)
// 	resource.EXPECT().GetID().Return("G1")
// 	resource.EXPECT().Value().Return(5)

// 	p1Hand.EXPECT().ShowCards().Return([]ports.Card{resource})

// 	p1Field.EXPECT().Warriors().Return([]ports.Warrior{})
// 	p2Field.EXPECT().Warriors().Return([]ports.Warrior{})

// 	gameStatus := NewGameStatus(p1, p2)

// 	assert.False(t, gameStatus.CanInitiateCastle)
// 	assert.True(t, gameStatus.CanGrowCastle) // Always true when resource in hand
// 	assert.True(t, gameStatus.CanBuy)
// 	assert.Equal(t, 1, len(gameStatus.CurrentPlayerHand))
// 	assert.False(t, gameStatus.CurrentPlayerHand[0].CanConstruct)
// }

// func TestGameStatus_ResourceInHand_CanBuy_True(t *testing.T) {
// 	ctrl := gomock.NewController(t)
// 	defer ctrl.Finish()

// 	p1, p1Hand, p1Field, _, p2, p2Field, _ := setupBaseMocks(ctrl)

// 	// Create resource mock that can buy
// 	resource := mocks.NewMockResource(ctrl)
// 	resource.EXPECT().CanConstruct().Return(false).Times(2)
// 	resource.EXPECT().CanBuy().Return(true)
// 	resource.EXPECT().GetID().Return("G1")
// 	resource.EXPECT().Value().Return(9)

// 	p1Hand.EXPECT().ShowCards().Return([]ports.Card{resource})

// 	p1Field.EXPECT().Warriors().Return([]ports.Warrior{})
// 	p2Field.EXPECT().Warriors().Return([]ports.Warrior{})

// 	gameStatus := NewGameStatus(p1, p2)

// 	assert.True(t, gameStatus.CanBuy)
// }

// func TestGameStatus_ResourceInHand_CanBuy_False(t *testing.T) {
// 	ctrl := gomock.NewController(t)
// 	defer ctrl.Finish()

// 	p1, p1Hand, p1Field, _, p2, p2Field, _ := setupBaseMocks(ctrl)

// 	// Create resource mock that cannot buy (low value)
// 	resource := mocks.NewMockResource(ctrl)
// 	resource.EXPECT().CanConstruct().Return(true).Times(2)
// 	resource.EXPECT().CanBuy().Return(false)
// 	resource.EXPECT().GetID().Return("G1")
// 	resource.EXPECT().Value().Return(1)

// 	p1Hand.EXPECT().ShowCards().Return([]ports.Card{resource})

// 	p1Field.EXPECT().Warriors().Return([]ports.Warrior{})
// 	p2Field.EXPECT().Warriors().Return([]ports.Warrior{})

// 	gameStatus := NewGameStatus(p1, p2)

// 	assert.False(t, gameStatus.CanBuy)
// }

// func TestGameStatus_MultipleResourcesInHand_CanInitiateCastleIsOr(t *testing.T) {
// 	ctrl := gomock.NewController(t)
// 	defer ctrl.Finish()

// 	p1, p1Hand, p1Field, _, p2, p2Field, _ := setupBaseMocks(ctrl)

// 	// First resource cannot construct
// 	resource1 := mocks.NewMockResource(ctrl)
// 	resource1.EXPECT().CanConstruct().Return(false).Times(2)
// 	resource1.EXPECT().CanBuy().Return(true)
// 	resource1.EXPECT().GetID().Return("G1")
// 	resource1.EXPECT().Value().Return(5)

// 	// Second resource can construct - should set CanInitiateCastle to true
// 	resource2 := mocks.NewMockResource(ctrl)
// 	resource2.EXPECT().CanConstruct().Return(true).Times(2)
// 	resource2.EXPECT().CanBuy().Return(false)
// 	resource2.EXPECT().GetID().Return("G2")
// 	resource2.EXPECT().Value().Return(1)

// 	p1Hand.EXPECT().ShowCards().Return([]ports.Card{resource1, resource2})

// 	p1Field.EXPECT().Warriors().Return([]ports.Warrior{})
// 	p2Field.EXPECT().Warriors().Return([]ports.Warrior{})

// 	gameStatus := NewGameStatus(p1, p2)

// 	assert.True(t, gameStatus.CanInitiateCastle) // true because resource2 can construct
// 	assert.True(t, gameStatus.CanGrowCastle)
// 	assert.False(t, gameStatus.CanBuy) // Last resource's CanBuy value wins
// 	assert.Equal(t, 2, len(gameStatus.CurrentPlayerHand))
// }

// func TestGameStatus_ResourceInHand_AlwaysSetsCanGrowCastle(t *testing.T) {
// 	ctrl := gomock.NewController(t)
// 	defer ctrl.Finish()

// 	p1, p1Hand, p1Field, _, p2, p2Field, _ := setupBaseMocks(ctrl)

// 	// Any resource sets CanGrowCastle to true
// 	resource := mocks.NewMockResource(ctrl)
// 	resource.EXPECT().CanConstruct().Return(false).Times(2)
// 	resource.EXPECT().CanBuy().Return(false)
// 	resource.EXPECT().GetID().Return("G1")
// 	resource.EXPECT().Value().Return(3)

// 	p1Hand.EXPECT().ShowCards().Return([]ports.Card{resource})

// 	p1Field.EXPECT().Warriors().Return([]ports.Warrior{})
// 	p2Field.EXPECT().Warriors().Return([]ports.Warrior{})

// 	gameStatus := NewGameStatus(p1, p2)

// 	assert.True(t, gameStatus.CanGrowCastle) // Always true when resource in hand
// 	assert.False(t, gameStatus.CanInitiateCastle)
// 	assert.False(t, gameStatus.CanBuy)
// }

// func TestGameStatus_CardValues_RealCards(t *testing.T) {
// 	ctrl := gomock.NewController(t)
// 	defer ctrl.Finish()

// 	// Create real warrior and weapon cards
// 	knight := cards.NewKnight("k1")
// 	sword := cards.NewSword("s1", 7)
// 	arrow := cards.NewArrow("a1", 5)
// 	gold := cards.NewGold("g1", 9)

// 	cardsInHand := []ports.Card{knight, sword, arrow, gold}

// 	// Setup mocks
// 	p1 := mocks.NewMockPlayer(ctrl)
// 	p1Hand := mocks.NewMockHand(ctrl)
// 	p1Field := mocks.NewMockField(ctrl)
// 	p1Castle := mocks.NewMockCastle(ctrl)

// 	p1.EXPECT().Name().Return("p1")
// 	p1.EXPECT().Hand().Return(p1Hand)
// 	p1.EXPECT().Field().Return(p1Field).AnyTimes()
// 	p1.EXPECT().Castle().Return(p1Castle)

// 	p1Hand.EXPECT().ShowCards().Return(cardsInHand)
// 	p1Field.EXPECT().Warriors().Return([]ports.Warrior{})
// 	p1Field.EXPECT().HasKnight().Return(true).AnyTimes()
// 	p1Field.EXPECT().HasArcher().Return(true).AnyTimes()
// 	p1Field.EXPECT().HasMage().Return(false).AnyTimes()
// 	p1Field.EXPECT().HasDragon().Return(false).AnyTimes()
// 	p1Castle.EXPECT().IsConstructed().Return(false)
// 	p1Castle.EXPECT().ResourceCards().Return(0)
// 	p1Castle.EXPECT().Value().Return(0)

// 	p2 := mocks.NewMockPlayer(ctrl)
// 	p2Field := mocks.NewMockField(ctrl)
// 	p2Castle := mocks.NewMockCastle(ctrl)

// 	p2.EXPECT().Field().Return(p2Field).AnyTimes()
// 	p2.EXPECT().Castle().Return(p2Castle).AnyTimes()
// 	p2.EXPECT().CardsInHand().Return(0)

// 	p2Field.EXPECT().Warriors().Return([]ports.Warrior{})
// 	p2Field.EXPECT().AttackableIDs().Return([]string{}).AnyTimes()
// 	p2Castle.EXPECT().IsConstructed().Return(false)
// 	p2Castle.EXPECT().ResourceCards().Return(0).AnyTimes()
// 	p2Castle.EXPECT().Value().Return(0)

// 	gameStatus := NewGameStatus(p1, p2)

// 	// Verify we have 4 cards
// 	assert.Equal(t, 4, len(gameStatus.CurrentPlayerHand))

// 	// Find each card and verify its value
// 	for _, hc := range gameStatus.CurrentPlayerHand {
// 		switch hc.CardID {
// 		case "K1":
// 			assert.Equal(t, 20, hc.Value, "Knight should have health of 20")
// 		case "S1":
// 			assert.Equal(t, 7, hc.Value, "Sword should have damage of 7")
// 		case "A1":
// 			assert.Equal(t, 5, hc.Value, "Arrow should have damage of 5")
// 		case "G1":
// 			assert.Equal(t, 9, hc.Value, "Gold should have value of 9")
// 		}
// 	}
// }

// /* func TestGameStatus_UsableWeapons_All(t *testing.T) {
// 	k := cards.NewKnight("k1")
// 	a := cards.NewArcher("a1")
// 	m := cards.NewMage("m1")

// 	cardsInField := []ports.Warrior{k, a, m}
// 	cardsInHand := []ports.Card{
// 		cards.NewSword("s1", 5),
// 		cards.NewArrow("a1", 3),
// 		cards.NewPoison("p1", 4),
// 	}
// 	p1 := newPlayerWithCards("p1", cardsInHand, cardsInField)
// 	p2 := newPlayerWithCards("p2", nil, nil)
// 	gameStatus := NewGameStatus(p1, p2)

// 	assert.Equal(t, 3, len(gameStatus.UsableWeaponIDs))
// 	assert.Contains(t, gameStatus.UsableWeaponIDs, "S1")
// 	assert.Contains(t, gameStatus.UsableWeaponIDs, "A1")
// 	assert.Contains(t, gameStatus.UsableWeaponIDs, "P1")
// }

// func TestGameStatus_UsableWeapons_Two(t *testing.T) {
// 	k := cards.NewKnight("k1")
// 	a := cards.NewArcher("a1")

// 	cardsInField := []ports.Warrior{k, a}
// 	cardsInHand := []ports.Card{
// 		cards.NewSword("s1", 5),
// 		cards.NewArrow("a1", 3),
// 		cards.NewPoison("p1", 4),
// 	}
// 	p1 := newPlayerWithCards("p1", cardsInHand, cardsInField)
// 	p2 := newPlayerWithCards("p2", nil, nil)
// 	gameStatus := NewGameStatus(p1, p2)

// 	assert.Equal(t, 2, len(gameStatus.UsableWeaponIDs))
// 	assert.Contains(t, gameStatus.UsableWeaponIDs, "S1")
// 	assert.Contains(t, gameStatus.UsableWeaponIDs, "A1")
// 	assert.NotContains(t, gameStatus.UsableWeaponIDs, "P1")
// }

// func TestGameStatus_ConstructionIDs_AsWeapons(t *testing.T) {
// 	cardsInHand := []ports.Card{
// 		cards.NewSword("s1", 1),
// 		cards.NewSword("s2", 5),
// 		cards.NewArrow("a1", 1),
// 		cards.NewArrow("a2", 8),
// 		cards.NewPoison("p1", 1),
// 		cards.NewPoison("p2", 9),
// 	}
// 	p1 := newPlayerWithCards("p1", cardsInHand, nil)
// 	p2 := newPlayerWithCards("p2", nil, nil)
// 	gameStatus := NewGameStatus(p1, p2)

// 	assert.Equal(t, 3, len(gameStatus.ConstructionIDs))
// 	assert.Contains(t, gameStatus.ConstructionIDs, "S1")
// 	assert.Contains(t, gameStatus.ConstructionIDs, "A1")
// 	assert.Contains(t, gameStatus.ConstructionIDs, "P1")
// 	assert.NotContains(t, gameStatus.ConstructionIDs, "S2")
// 	assert.NotContains(t, gameStatus.ConstructionIDs, "A2")
// 	assert.NotContains(t, gameStatus.ConstructionIDs, "P2")
// }

// func TestGameStatus_ConstructionIDs_AsResource(t *testing.T) {

// 	cardsInHand := []ports.Card{
// 		cards.NewGold("g1", 1),
// 		cards.NewGold("g2", 9),
// 	}

// 	p1 := newPlayerWithCards("p1", cardsInHand, nil)
// 	p2 := newPlayerWithCards("p2", nil, nil)
// 	gameStatus := NewGameStatus(p1, p2)

// 	assert.Equal(t, 1, len(gameStatus.ConstructionIDs))
// 	assert.Contains(t, gameStatus.ConstructionIDs, "G1")
// 	assert.NotContains(t, gameStatus.ConstructionIDs, "G2")
// }

// func TestGameStatus_ResourceIDs(t *testing.T) {

// 	cardsInHand := []ports.Card{
// 		cards.NewGold("g1", 1),
// 		cards.NewGold("g2", 9),
// 	}

// 	p1 := newPlayerWithCards("p1", cardsInHand, nil)
// 	p2 := newPlayerWithCards("p2", nil, nil)
// 	gameStatus := NewGameStatus(p1, p2)

// 	assert.Equal(t, 2, len(gameStatus.ResourceIDs))
// 	assert.Contains(t, gameStatus.ResourceIDs, "G1")
// 	assert.Contains(t, gameStatus.ResourceIDs, "G2")
// }

// func TestGameStatus_SpecialPower_CanProtect(t *testing.T) {
// 	sp := cards.NewSpecialPower("sp1")
// 	cardsInHand := []ports.Card{sp}

// 	cardsInField := []ports.Warrior{
// 		cards.NewKnight("m1"),
// 		cards.NewArcher("a1"),
// 		cards.NewDragon("d2"),
// 	}
// 	enemyField := []ports.Warrior{
// 		cards.NewKnight("ek1"),
// 	}

// 	p1 := newPlayerWithCards("p1", cardsInHand, cardsInField)
// 	p2 := newPlayerWithCards("p2", nil, enemyField)
// 	gameStatus := NewGameStatus(p1, p2)

// 	assert.Equal(t, 1, len(gameStatus.SpecialPowerStatus.SpecialPowerIDs))
// 	assert.Equal(t, 2, len(gameStatus.SpecialPowerStatus.CanProtectIDs))
// 	assert.Contains(t, gameStatus.SpecialPowerStatus.SpecialPowerIDs, "SP1")
// 	assert.Contains(t, gameStatus.SpecialPowerStatus.CanProtectIDs, "M1")
// 	assert.Contains(t, gameStatus.SpecialPowerStatus.CanProtectIDs, "A1")
// 	assert.NotContains(t, gameStatus.SpecialPowerStatus.CanProtectIDs, "EK1")
// 	assert.NotContains(t, gameStatus.SpecialPowerStatus.CanProtectIDs, "D2")
// }

// func TestGameStatus_SpecialPower_CanInstantKill(t *testing.T) {
// 	sp := cards.NewSpecialPower("sp1")
// 	cardsInHand := []ports.Card{sp}

// 	// Enemy field: one protected, one unprotected
// 	shield := cards.NewSpecialPower("shield1")
// 	protectedWarrior := cards.NewKnight("ek1")
// 	protectedWarrior.Protect(shield)

// 	unprotectedWarrior := cards.NewArcher("ea1")

// 	enemyField := []ports.Warrior{protectedWarrior, unprotectedWarrior}
// 	myField := []ports.Warrior{cards.NewArcher("a1")}

// 	p1 := newPlayerWithCards("p1", cardsInHand, myField)
// 	p2 := newPlayerWithCards("p2", nil, enemyField)
// 	gameStatus := NewGameStatus(p1, p2)

// 	assert.Equal(t, 1, len(gameStatus.SpecialPowerStatus.SpecialPowerIDs))
// 	assert.Equal(t, 2, len(gameStatus.SpecialPowerStatus.CanInstantKillIDs))
// 	assert.Contains(t, gameStatus.SpecialPowerStatus.SpecialPowerIDs, "SP1")
// 	assert.Contains(t, gameStatus.SpecialPowerStatus.CanInstantKillIDs, "EA1")
// 	assert.Contains(t, gameStatus.SpecialPowerStatus.CanInstantKillIDs, "SHIELD1")
// 	assert.NotContains(t, gameStatus.SpecialPowerStatus.CanInstantKillIDs, "EK1")
// }

// func TestGameStatus_SpecialPower_CanHeal(t *testing.T) {
// 	sp := cards.NewSpecialPower("sp1")
// 	cardsInHand := []ports.Card{sp}

// 	arrow := cards.NewArrow("a1", 4)
// 	damagedWarrior := cards.NewKnight("ek1")
// 	damagedWarrior.ReceiveDamage(arrow, 1)

// 	myField := []ports.Warrior{damagedWarrior,
// 		cards.NewMage("m1")}

// 	p1 := newPlayerWithCards("p1", cardsInHand, myField)
// 	p2 := newPlayerWithCards("p2", nil, nil)
// 	gameStatus := NewGameStatus(p1, p2)

// 	assert.Equal(t, 1, len(gameStatus.SpecialPowerStatus.SpecialPowerIDs))
// 	assert.Equal(t, 1, len(gameStatus.SpecialPowerStatus.CanHealIDs))
// 	assert.Contains(t, gameStatus.SpecialPowerStatus.SpecialPowerIDs, "SP1")
// 	assert.Contains(t, gameStatus.SpecialPowerStatus.CanHealIDs, "EK1")
// 	assert.NotContains(t, gameStatus.SpecialPowerStatus.CanHealIDs, "M1")
// }
// */
