package domain

import (
	"errors"
	"strings"
	"testing"

	"github.com/alelopezbcn/thecampaign/internal/domain/cards"
	"github.com/alelopezbcn/thecampaign/internal/domain/ports"
	"github.com/alelopezbcn/thecampaign/internal/domain/types"
	"github.com/alelopezbcn/thecampaign/test/mocks"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

func TestAttackAction_PlayerName(t *testing.T) {
	action := NewAttackAction("Player1", "Player2", "t1", "w1")
	assert.Equal(t, "Player1", action.PlayerName())
}

func TestAttackAction_NextPhase(t *testing.T) {
	action := NewAttackAction("Player1", "Player2", "t1", "w1")
	assert.Equal(t, types.PhaseTypeSpySteal, action.NextPhase())
}

func TestAttackAction_Validate(t *testing.T) {
	t.Run("Error when not in Attack phase", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockPlayer1 := mocks.NewMockPlayer(ctrl)
		mockPlayer2 := mocks.NewMockPlayer(ctrl)

		mockPlayer1.EXPECT().Name().Return("Player1").AnyTimes()

		g := &Game{
			Players:       []ports.Player{mockPlayer1, mockPlayer2},
			CurrentTurn:   0,
			currentAction: types.PhaseTypeBuy,
		}

		action := NewAttackAction("Player1", "Player2", "targetID", "weaponID")
		err := action.Validate(g)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "cannot attack in the")
	})

	t.Run("Error when target card not in enemy field", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockPlayer1 := mocks.NewMockPlayer(ctrl)
		mockPlayer2 := mocks.NewMockPlayer(ctrl)

		mockPlayer1.EXPECT().Name().Return("Player1").AnyTimes()
		mockPlayer2.EXPECT().Name().Return("Player2").AnyTimes()
		mockPlayer2.EXPECT().GetCardFromField("targetID").Return(nil, false)

		g := &Game{
			Players:       []ports.Player{mockPlayer1, mockPlayer2},
			CurrentTurn:   0,
			currentAction: types.PhaseTypeAttack,
		}

		action := NewAttackAction("Player1", "Player2", "targetID", "weaponID")
		err := action.Validate(g)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "target card not in enemy field")
	})

	t.Run("Error when weapon card not in hand", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockPlayer1 := mocks.NewMockPlayer(ctrl)
		mockPlayer2 := mocks.NewMockPlayer(ctrl)
		mockWarrior := mocks.NewMockWarrior(ctrl)

		mockPlayer1.EXPECT().Name().Return("Player1").AnyTimes()
		mockPlayer2.EXPECT().Name().Return("Player2").AnyTimes()
		mockPlayer2.EXPECT().GetCardFromField("targetID").Return(mockWarrior, true)
		mockPlayer1.EXPECT().GetCardFromHand("weaponID").Return(nil, false)

		g := &Game{
			Players:       []ports.Player{mockPlayer1, mockPlayer2},
			CurrentTurn:   0,
			currentAction: types.PhaseTypeAttack,
		}

		action := NewAttackAction("Player1", "Player2", "targetID", "weaponID")
		err := action.Validate(g)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "weapon card not in hand")
	})

	t.Run("Error when target is not attackable", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockPlayer1 := mocks.NewMockPlayer(ctrl)
		mockPlayer2 := mocks.NewMockPlayer(ctrl)
		mockCard := mocks.NewMockCard(ctrl) // Not Attackable
		mockWeapon := mocks.NewMockWeapon(ctrl)

		mockPlayer1.EXPECT().Name().Return("Player1").AnyTimes()
		mockPlayer2.EXPECT().Name().Return("Player2").AnyTimes()
		mockPlayer2.EXPECT().GetCardFromField("targetID").Return(mockCard, true)
		mockPlayer1.EXPECT().GetCardFromHand("weaponID").Return(mockWeapon, true)

		g := &Game{
			Players:       []ports.Player{mockPlayer1, mockPlayer2},
			CurrentTurn:   0,
			currentAction: types.PhaseTypeAttack,
		}

		action := NewAttackAction("Player1", "Player2", "targetID", "weaponID")
		err := action.Validate(g)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "the target cardBase cannot be attacked")
	})

	t.Run("Error when card is not a weapon", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockPlayer1 := mocks.NewMockPlayer(ctrl)
		mockPlayer2 := mocks.NewMockPlayer(ctrl)
		mockWarrior := mocks.NewMockWarrior(ctrl)
		mockResource := mocks.NewMockResource(ctrl) // Not a weapon

		mockPlayer1.EXPECT().Name().Return("Player1").AnyTimes()
		mockPlayer2.EXPECT().Name().Return("Player2").AnyTimes()
		mockPlayer2.EXPECT().GetCardFromField("targetID").Return(mockWarrior, true)
		mockPlayer1.EXPECT().GetCardFromHand("weaponID").Return(mockResource, true)

		g := &Game{
			Players:       []ports.Player{mockPlayer1, mockPlayer2},
			CurrentTurn:   0,
			currentAction: types.PhaseTypeAttack,
		}

		action := NewAttackAction("Player1", "Player2", "targetID", "weaponID")
		err := action.Validate(g)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "the card is not a weapon")
	})

	t.Run("Success stores target and weapon", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockPlayer1 := mocks.NewMockPlayer(ctrl)
		mockPlayer2 := mocks.NewMockPlayer(ctrl)
		mockWarrior := mocks.NewMockWarrior(ctrl)
		mockWeapon := mocks.NewMockWeapon(ctrl)

		mockPlayer1.EXPECT().Name().Return("Player1").AnyTimes()
		mockPlayer2.EXPECT().Name().Return("Player2").AnyTimes()
		mockPlayer2.EXPECT().GetCardFromField("targetID").Return(mockWarrior, true)
		mockPlayer1.EXPECT().GetCardFromHand("weaponID").Return(mockWeapon, true)

		g := &Game{
			Players:       []ports.Player{mockPlayer1, mockPlayer2},
			CurrentTurn:   0,
			currentAction: types.PhaseTypeAttack,
		}

		action := NewAttackAction("Player1", "Player2", "targetID", "weaponID")
		err := action.Validate(g)

		assert.NoError(t, err)
		assert.Equal(t, mockWarrior, action.target)
		assert.Equal(t, mockWeapon, action.weapon)
	})
}

func TestAttackAction_Execute(t *testing.T) {
	t.Run("Error when attack fails", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockPlayer1 := mocks.NewMockPlayer(ctrl)
		mockPlayer2 := mocks.NewMockPlayer(ctrl)
		mockWarrior := mocks.NewMockWarrior(ctrl)
		mockWeapon := mocks.NewMockWeapon(ctrl)

		mockPlayer1.EXPECT().Name().Return("Player1").AnyTimes()
		mockPlayer1.EXPECT().Attack(mockWarrior, mockWeapon).Return(errors.New("attack failed"))

		g := &Game{
			Players:       []ports.Player{mockPlayer1, mockPlayer2},
			CurrentTurn:   0,
			currentAction: types.PhaseTypeAttack,
		}

		action := NewAttackAction("Player1", "Player2", "targetID", "weaponID")
		action.target = mockWarrior
		action.weapon = mockWeapon

		result, _, err := action.Execute(g)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "attack action failed")
		assert.NotNil(t, result)
	})

	t.Run("Success returns result with attack details", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockPlayer1 := mocks.NewMockPlayer(ctrl)
		mockPlayer2 := mocks.NewMockPlayer(ctrl)
		mockProvider := NewMockGameStatusProvider(ctrl)
		mockWarrior := mocks.NewMockWarrior(ctrl)
		mockWeapon := mocks.NewMockWeapon(ctrl)

		expectedStatus := GameStatus{CurrentPlayer: "Player1"}

		mockPlayer1.EXPECT().Name().Return("Player1").AnyTimes()
		mockPlayer1.EXPECT().Attack(mockWarrior, mockWeapon).Return(nil)
		mockWarrior.EXPECT().String().Return("Knight (20)")
		mockWeapon.EXPECT().String().Return("Sword (5)")

		g := &Game{
			Players:            []ports.Player{mockPlayer1, mockPlayer2},
			CurrentTurn:        0,
			currentAction:      types.PhaseTypeAttack,
			GameStatusProvider: mockProvider,
			history:            []historyLine{},
		}

		mockProvider.EXPECT().Get(mockPlayer1, g).Return(expectedStatus)

		action := NewAttackAction("Player1", "Player2", "K1", "S1")
		action.target = mockWarrior
		action.weapon = mockWeapon

		result, statusFn, err := action.Execute(g)

		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Equal(t, types.LastActionAttack, result.Action)
		assert.Equal(t, "S1", result.AttackWeaponID)
		assert.Equal(t, "K1", result.AttackTargetID)
		assert.Equal(t, "Player2", result.AttackTargetPlayer)
		assert.Equal(t, expectedStatus, statusFn())
	})

	t.Run("History is updated on successful attack", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockPlayer1 := mocks.NewMockPlayer(ctrl)
		mockPlayer2 := mocks.NewMockPlayer(ctrl)
		mockWarrior := mocks.NewMockWarrior(ctrl)
		mockWeapon := mocks.NewMockWeapon(ctrl)

		mockPlayer1.EXPECT().Name().Return("Player1").AnyTimes()
		mockPlayer1.EXPECT().Attack(mockWarrior, mockWeapon).Return(nil)
		mockWarrior.EXPECT().String().Return("Knight (20)")
		mockWeapon.EXPECT().String().Return("Sword (5)")

		g := &Game{
			Players:       []ports.Player{mockPlayer1, mockPlayer2},
			CurrentTurn:   0,
			currentAction: types.PhaseTypeAttack,
			history:       []historyLine{},
		}

		action := NewAttackAction("Player1", "Player2", "K1", "S1")
		action.target = mockWarrior
		action.weapon = mockWeapon

		_, statusFn, err := action.Execute(g)

		assert.NoError(t, err)
		assert.NotNil(t, statusFn)
		found := false
		for _, h := range g.history {
			if strings.Contains(h.Msg, "Player1") && strings.Contains(h.Msg, "attacked") {
				found = true
				break
			}
		}
		assert.True(t, found, "History should contain attack action")
	})
}

// Integration tests using real card implementations to verify combat damage mechanics.
func TestAttackAction_CombatDamage(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mockProvider := NewMockGameStatusProvider(ctrl)
	mockProvider.EXPECT().Get(gomock.Any(), gomock.Any()).Return(GameStatus{}).AnyTimes()

	newAttackGame := func(
		p1Cards []ports.Card, p1Warriors []ports.Warrior,
		p2Cards []ports.Card, p2Warriors []ports.Warrior,
	) (g *Game, p1, p2 ports.Player) {
		g = &Game{
			currentAction:      types.PhaseTypeAttack,
			history:            []historyLine{},
			discardPile:        newDiscardPile(),
			cemetery:           newCemetery(),
			GameStatusProvider: mockProvider,
		}
		p1 = newPlayerWithCardAndObserver("Player1", p1Cards, p1Warriors, g)
		p2 = newPlayerWithCardAndObserver("Player2", p2Cards, p2Warriors, g)
		g.Players = []ports.Player{p1, p2}
		return g, p1, p2
	}

	executeAttack := func(g *Game, playerName, targetPlayerName, targetID, weaponID string) error {
		_, err := g.ExecuteAction(NewAttackAction(playerName, targetPlayerName, targetID, weaponID))
		return err
	}

	// Knight attacks
	t.Run("Knight attacks Archer causing double damage", func(t *testing.T) {
		dmgAmnt := 4
		k := cards.NewKnight("k1")
		a := cards.NewArcher("a1")
		sword := cards.NewSword("s1", dmgAmnt)
		g, _, _ := newAttackGame([]ports.Card{sword}, []ports.Warrior{k}, []ports.Card{}, []ports.Warrior{a})

		err := executeAttack(g, "Player1", "Player2", a.GetID(), sword.GetID())
		assert.NoError(t, err)
		assert.Equal(t, 20-dmgAmnt*2, a.Health())
	})

	t.Run("Knight attacks Mage causing normal damage", func(t *testing.T) {
		dmgAmnt := 4
		k := cards.NewKnight("k1")
		m := cards.NewMage("m1")
		sword := cards.NewSword("s1", dmgAmnt)
		g, _, _ := newAttackGame([]ports.Card{sword}, []ports.Warrior{k}, []ports.Card{}, []ports.Warrior{m})

		err := executeAttack(g, "Player1", "Player2", m.GetID(), sword.GetID())
		assert.NoError(t, err)
		assert.Equal(t, 20-dmgAmnt*1, m.Health())
	})

	t.Run("Knight attacks Knight causing normal damage", func(t *testing.T) {
		dmgAmnt := 4
		k := cards.NewKnight("k1")
		k2 := cards.NewKnight("k2")
		sword := cards.NewSword("s1", dmgAmnt)
		g, _, _ := newAttackGame([]ports.Card{sword}, []ports.Warrior{k}, []ports.Card{}, []ports.Warrior{k2})

		err := executeAttack(g, "Player1", "Player2", k2.GetID(), sword.GetID())
		assert.NoError(t, err)
		assert.Equal(t, 20-dmgAmnt*1, k2.Health())
	})

	t.Run("Knight attacks Dragon causing normal damage", func(t *testing.T) {
		dmgAmnt := 4
		k := cards.NewKnight("k1")
		d := cards.NewDragon("d1")
		sword := cards.NewSword("s1", dmgAmnt)
		g, _, _ := newAttackGame([]ports.Card{sword}, []ports.Warrior{k}, []ports.Card{}, []ports.Warrior{d})

		err := executeAttack(g, "Player1", "Player2", d.GetID(), sword.GetID())
		assert.NoError(t, err)
		assert.Equal(t, 20-dmgAmnt*1, d.Health())
	})

	t.Run("Knight cant attack with wrong weapon", func(t *testing.T) {
		k := cards.NewKnight("k1")
		a := cards.NewArcher("a1")
		poison := cards.NewPoison("s1", 4)
		g, p1, _ := newAttackGame([]ports.Card{poison}, []ports.Warrior{k}, []ports.Card{}, []ports.Warrior{a})

		err := executeAttack(g, "Player1", "Player2", a.GetID(), poison.GetID())
		assert.Error(t, err)
		assert.Equal(t, 20, a.Health())
		assert.Contains(t, p1.Hand().ShowCards(), poison)
	})

	// Archer attacks
	t.Run("Archer attacks Mage causing double damage", func(t *testing.T) {
		dmgAmnt := 4
		attacker := cards.NewArcher("a1")
		target := cards.NewMage("m1")
		weapon := cards.NewArrow("s1", dmgAmnt)
		g, _, _ := newAttackGame([]ports.Card{weapon}, []ports.Warrior{attacker}, []ports.Card{}, []ports.Warrior{target})

		err := executeAttack(g, "Player1", "Player2", target.GetID(), weapon.GetID())
		assert.NoError(t, err)
		assert.Equal(t, 20-dmgAmnt*2, target.Health())
	})

	t.Run("Archer attacks Knight causing normal damage", func(t *testing.T) {
		dmgAmnt := 4
		attacker := cards.NewArcher("a1")
		target := cards.NewKnight("k1")
		weapon := cards.NewArrow("s1", dmgAmnt)
		g, _, _ := newAttackGame([]ports.Card{weapon}, []ports.Warrior{attacker}, []ports.Card{}, []ports.Warrior{target})

		err := executeAttack(g, "Player1", "Player2", target.GetID(), weapon.GetID())
		assert.NoError(t, err)
		assert.Equal(t, 20-dmgAmnt*1, target.Health())
	})

	t.Run("Archer attacks Archer causing normal damage", func(t *testing.T) {
		dmgAmnt := 4
		attacker := cards.NewArcher("a1")
		target := cards.NewArcher("a2")
		weapon := cards.NewArrow("s1", dmgAmnt)
		g, _, _ := newAttackGame([]ports.Card{weapon}, []ports.Warrior{attacker}, []ports.Card{}, []ports.Warrior{target})

		err := executeAttack(g, "Player1", "Player2", target.GetID(), weapon.GetID())
		assert.NoError(t, err)
		assert.Equal(t, 20-dmgAmnt*1, target.Health())
	})

	t.Run("Archer attacks Dragon causing normal damage", func(t *testing.T) {
		dmgAmnt := 4
		attacker := cards.NewArcher("a1")
		target := cards.NewDragon("d1")
		weapon := cards.NewArrow("s1", dmgAmnt)
		g, _, _ := newAttackGame([]ports.Card{weapon}, []ports.Warrior{attacker}, []ports.Card{}, []ports.Warrior{target})

		err := executeAttack(g, "Player1", "Player2", target.GetID(), weapon.GetID())
		assert.NoError(t, err)
		assert.Equal(t, 20-dmgAmnt*1, target.Health())
	})

	t.Run("Archer cant attack with wrong weapon", func(t *testing.T) {
		attacker := cards.NewArcher("a1")
		target := cards.NewMage("m1")
		weapon := cards.NewSword("s1", 4)
		g, p1, _ := newAttackGame([]ports.Card{weapon}, []ports.Warrior{attacker}, []ports.Card{}, []ports.Warrior{target})

		err := executeAttack(g, "Player1", "Player2", target.GetID(), weapon.GetID())
		assert.Error(t, err)
		assert.Equal(t, 20, target.Health())
		assert.Contains(t, p1.Hand().ShowCards(), weapon)
	})

	// Mage attacks
	t.Run("Mage attacks Knight causing double damage", func(t *testing.T) {
		dmgAmnt := 4
		attacker := cards.NewMage("m1")
		target := cards.NewKnight("k1")
		weapon := cards.NewPoison("s1", dmgAmnt)
		g, _, _ := newAttackGame([]ports.Card{weapon}, []ports.Warrior{attacker}, []ports.Card{}, []ports.Warrior{target})

		err := executeAttack(g, "Player1", "Player2", target.GetID(), weapon.GetID())
		assert.NoError(t, err)
		assert.Equal(t, 20-dmgAmnt*2, target.Health())
	})

	t.Run("Mage attacks Archer causing normal damage", func(t *testing.T) {
		dmgAmnt := 4
		attacker := cards.NewMage("m1")
		target := cards.NewArcher("a1")
		weapon := cards.NewPoison("s1", dmgAmnt)
		g, _, _ := newAttackGame([]ports.Card{weapon}, []ports.Warrior{attacker}, []ports.Card{}, []ports.Warrior{target})

		err := executeAttack(g, "Player1", "Player2", target.GetID(), weapon.GetID())
		assert.NoError(t, err)
		assert.Equal(t, 20-dmgAmnt*1, target.Health())
	})

	t.Run("Mage attacks Mage causing normal damage", func(t *testing.T) {
		dmgAmnt := 4
		attacker := cards.NewMage("m1")
		target := cards.NewMage("m2")
		weapon := cards.NewPoison("s1", dmgAmnt)
		g, _, _ := newAttackGame([]ports.Card{weapon}, []ports.Warrior{attacker}, []ports.Card{}, []ports.Warrior{target})

		err := executeAttack(g, "Player1", "Player2", target.GetID(), weapon.GetID())
		assert.NoError(t, err)
		assert.Equal(t, 20-dmgAmnt*1, target.Health())
	})

	t.Run("Mage attacks Dragon causing normal damage", func(t *testing.T) {
		dmgAmnt := 4
		attacker := cards.NewMage("m1")
		target := cards.NewDragon("d1")
		weapon := cards.NewPoison("s1", dmgAmnt)
		g, _, _ := newAttackGame([]ports.Card{weapon}, []ports.Warrior{attacker}, []ports.Card{}, []ports.Warrior{target})

		err := executeAttack(g, "Player1", "Player2", target.GetID(), weapon.GetID())
		assert.NoError(t, err)
		assert.Equal(t, 20-dmgAmnt*1, target.Health())
	})

	t.Run("Mage cant attack with wrong weapon", func(t *testing.T) {
		attacker := cards.NewMage("m1")
		target := cards.NewKnight("k1")
		weapon := cards.NewArrow("s1", 4)
		g, p1, _ := newAttackGame([]ports.Card{weapon}, []ports.Warrior{attacker}, []ports.Card{}, []ports.Warrior{target})

		err := executeAttack(g, "Player1", "Player2", target.GetID(), weapon.GetID())
		assert.Error(t, err)
		assert.Equal(t, 20, target.Health())
		assert.Contains(t, p1.Hand().ShowCards(), weapon)
	})

	t.Run("Player cant attack with non existing cards", func(t *testing.T) {
		k := cards.NewKnight("k1")
		a := cards.NewArcher("a1")
		sword := cards.NewSword("s1", 4)
		g, _, _ := newAttackGame([]ports.Card{sword}, []ports.Warrior{k}, []ports.Card{}, []ports.Warrior{a})

		err := executeAttack(g, "Player1", "Player2", "non-existent-target", sword.GetID())
		assert.Error(t, err)

		err = executeAttack(g, "Player1", "Player2", a.GetID(), "non-existent-weapon")
		assert.Error(t, err)
	})

	// Dragon attacks
	t.Run("Dragon attacks Knight with Sword causing normal damage", func(t *testing.T) {
		dmgAmnt := 6
		attacker := cards.NewDragon("d1")
		target := cards.NewKnight("k1")
		weapon := cards.NewSword("s1", dmgAmnt)
		g, _, _ := newAttackGame([]ports.Card{weapon}, []ports.Warrior{attacker}, []ports.Card{}, []ports.Warrior{target})

		err := executeAttack(g, "Player1", "Player2", target.GetID(), weapon.GetID())
		assert.NoError(t, err)
		assert.Equal(t, 20-dmgAmnt*1, target.Health())
	})

	t.Run("Dragon attacks Knight with Arrow causing normal damage", func(t *testing.T) {
		dmgAmnt := 6
		attacker := cards.NewDragon("d1")
		target := cards.NewKnight("k1")
		weapon := cards.NewArrow("s1", dmgAmnt)
		g, _, _ := newAttackGame([]ports.Card{weapon}, []ports.Warrior{attacker}, []ports.Card{}, []ports.Warrior{target})

		err := executeAttack(g, "Player1", "Player2", target.GetID(), weapon.GetID())
		assert.NoError(t, err)
		assert.Equal(t, 20-dmgAmnt*1, target.Health())
	})

	t.Run("Dragon attacks Knight with Poison causing double damage", func(t *testing.T) {
		dmgAmnt := 6
		attacker := cards.NewDragon("d1")
		target := cards.NewKnight("k1")
		weapon := cards.NewPoison("s1", dmgAmnt)
		g, _, _ := newAttackGame([]ports.Card{weapon}, []ports.Warrior{attacker}, []ports.Card{}, []ports.Warrior{target})

		err := executeAttack(g, "Player1", "Player2", target.GetID(), weapon.GetID())
		assert.NoError(t, err)
		assert.Equal(t, 20-dmgAmnt*2, target.Health())
	})

	t.Run("Dragon attacks Archer with Sword causing double damage", func(t *testing.T) {
		dmgAmnt := 6
		attacker := cards.NewDragon("d1")
		target := cards.NewArcher("a1")
		weapon := cards.NewSword("s1", dmgAmnt)
		g, _, _ := newAttackGame([]ports.Card{weapon}, []ports.Warrior{attacker}, []ports.Card{}, []ports.Warrior{target})

		err := executeAttack(g, "Player1", "Player2", target.GetID(), weapon.GetID())
		assert.NoError(t, err)
		assert.Equal(t, 20-dmgAmnt*2, target.Health())
	})

	t.Run("Dragon attacks Archer with Arrow causing normal damage", func(t *testing.T) {
		dmgAmnt := 6
		attacker := cards.NewDragon("d1")
		target := cards.NewArcher("a1")
		weapon := cards.NewArrow("s1", dmgAmnt)
		g, _, _ := newAttackGame([]ports.Card{weapon}, []ports.Warrior{attacker}, []ports.Card{}, []ports.Warrior{target})

		err := executeAttack(g, "Player1", "Player2", target.GetID(), weapon.GetID())
		assert.NoError(t, err)
		assert.Equal(t, 20-dmgAmnt*1, target.Health())
	})

	t.Run("Dragon attacks Archer with Poison causing normal damage", func(t *testing.T) {
		dmgAmnt := 6
		attacker := cards.NewDragon("d1")
		target := cards.NewArcher("a1")
		weapon := cards.NewPoison("s1", dmgAmnt)
		g, _, _ := newAttackGame([]ports.Card{weapon}, []ports.Warrior{attacker}, []ports.Card{}, []ports.Warrior{target})

		err := executeAttack(g, "Player1", "Player2", target.GetID(), weapon.GetID())
		assert.NoError(t, err)
		assert.Equal(t, 20-dmgAmnt*1, target.Health())
	})

	t.Run("Dragon attacks Mage with Sword causing normal damage", func(t *testing.T) {
		dmgAmnt := 6
		attacker := cards.NewDragon("d1")
		target := cards.NewMage("m1")
		weapon := cards.NewSword("s1", dmgAmnt)
		g, _, _ := newAttackGame([]ports.Card{weapon}, []ports.Warrior{attacker}, []ports.Card{}, []ports.Warrior{target})

		err := executeAttack(g, "Player1", "Player2", target.GetID(), weapon.GetID())
		assert.NoError(t, err)
		assert.Equal(t, 20-dmgAmnt*1, target.Health())
	})

	t.Run("Dragon attacks Mage with Arrow causing double damage", func(t *testing.T) {
		dmgAmnt := 6
		attacker := cards.NewDragon("d1")
		target := cards.NewMage("m1")
		weapon := cards.NewArrow("s1", dmgAmnt)
		g, _, _ := newAttackGame([]ports.Card{weapon}, []ports.Warrior{attacker}, []ports.Card{}, []ports.Warrior{target})

		err := executeAttack(g, "Player1", "Player2", target.GetID(), weapon.GetID())
		assert.NoError(t, err)
		assert.Equal(t, 20-dmgAmnt*2, target.Health())
	})

	t.Run("Dragon attacks Mage with Poison causing normal damage", func(t *testing.T) {
		dmgAmnt := 6
		attacker := cards.NewDragon("d1")
		target := cards.NewMage("m1")
		weapon := cards.NewPoison("s1", dmgAmnt)
		g, _, _ := newAttackGame([]ports.Card{weapon}, []ports.Warrior{attacker}, []ports.Card{}, []ports.Warrior{target})

		err := executeAttack(g, "Player1", "Player2", target.GetID(), weapon.GetID())
		assert.NoError(t, err)
		assert.Equal(t, 20-dmgAmnt*1, target.Health())
	})

	// Multi-attack scenarios
	t.Run("Warrior dead on second attack", func(t *testing.T) {
		dmgAmnt := 5
		k := cards.NewKnight("k1")
		a := cards.NewArcher("a1")
		a2 := cards.NewArcher("a2")
		sword1 := cards.NewSword("s1", dmgAmnt)
		sword2 := cards.NewSword("s2", dmgAmnt)
		g := &Game{
			currentAction:      types.PhaseTypeAttack,
			history:            []historyLine{},
			discardPile:        newDiscardPile(),
			cemetery:           newCemetery(),
			GameStatusProvider: mockProvider,
		}
		p1 := newPlayerWithCardAndObserver("Player1",
			[]ports.Card{sword1, sword2},
			[]ports.Warrior{k},
			g,
		)
		p2 := newPlayerWithCardAndObserver("Player2",
			[]ports.Card{},
			[]ports.Warrior{a, a2},
			g,
		)

		g.Players = []ports.Player{p1, p2}

		err := executeAttack(g, p1.Name(), p2.Name(), a.GetID(), sword1.GetID())
		assert.NoError(t, err)
		assert.NotContains(t, p1.Hand().ShowCards(), sword1)

		g.currentAction = types.PhaseTypeAttack
		err = executeAttack(g, p1.Name(), p2.Name(), a.GetID(), sword2.GetID())
		assert.NoError(t, err)
		assert.NotContains(t, p1.Hand().ShowCards(), sword2)

		assert.Equal(t, 0, a.Health())
		_, ok := p2.GetCardFromField(a.GetID())
		assert.False(t, ok, "Archer should have been removed from field after death")
		_, ok = p2.GetCardFromField(a2.GetID())
		assert.True(t, ok, "Second Archer should still be on the field")
		assert.True(t, foundInCemetery(g, a), "Cemetery should contain the dead archer")
		assert.True(t, foundInDiscardPile(g, sword1), "Discard pile should contain the used sword")
		assert.True(t, foundInDiscardPile(g, sword2), "Discard pile should contain the used sword")
	})

	t.Run("Dragon dead on multiple attacks", func(t *testing.T) {
		dmgAmnt := 5
		m1 := cards.NewMage("m1")
		k2 := cards.NewKnight("k2")
		a3 := cards.NewArcher("a3")

		target := cards.NewDragon("d1")
		a2 := cards.NewArcher("a2")

		poison1 := cards.NewPoison("p1", dmgAmnt)
		sword2 := cards.NewSword("s2", dmgAmnt)
		arrow3 := cards.NewArrow("ar3", dmgAmnt)
		sword4 := cards.NewSword("s4", dmgAmnt)

		g := &Game{
			currentAction:      types.PhaseTypeAttack,
			history:            []historyLine{},
			discardPile:        newDiscardPile(),
			cemetery:           newCemetery(),
			GameStatusProvider: mockProvider,
		}
		p1 := newPlayerWithCardAndObserver("Player1",
			[]ports.Card{poison1, sword2, arrow3, sword4},
			[]ports.Warrior{m1, k2, a3},
			g,
		)
		p2 := newPlayerWithCardAndObserver("Player2",
			[]ports.Card{},
			[]ports.Warrior{target, a2},
			g,
		)

		g.Players = []ports.Player{p1, p2}

		err := executeAttack(g, p1.Name(), p2.Name(), target.GetID(), poison1.GetID())
		assert.NoError(t, err)
		g.currentAction = types.PhaseTypeAttack
		err = executeAttack(g, p1.Name(), p2.Name(), target.GetID(), sword2.GetID())
		assert.NoError(t, err)
		g.currentAction = types.PhaseTypeAttack
		err = executeAttack(g, p1.Name(), p2.Name(), target.GetID(), arrow3.GetID())
		assert.NoError(t, err)
		g.currentAction = types.PhaseTypeAttack
		err = executeAttack(g, p1.Name(), p2.Name(), target.GetID(), sword4.GetID())
		assert.NoError(t, err)

		assert.Equal(t, 0, target.Health())
		_, ok := p1.GetCardFromField(m1.GetID())
		assert.True(t, ok, "Mage should still be on the field")
		_, ok = p1.GetCardFromField(k2.GetID())
		assert.True(t, ok, "Knight should still be on the field")
		_, ok = p1.GetCardFromField(a3.GetID())
		assert.True(t, ok, "Archer should still be on the field")

		_, ok = p2.GetCardFromField(target.GetID())
		assert.False(t, ok, "Dragon should have been removed from field after death")
		_, ok = p2.GetCardFromField(a2.GetID())
		assert.True(t, ok, "Archer should still be on the field")

		_, ok = p1.GetCardFromHand(poison1.GetID())
		assert.False(t, ok, "Poison should have been discarded after attack")
		_, ok = p1.GetCardFromHand(sword2.GetID())
		assert.False(t, ok, "Sword should have been discarded after attack")
		_, ok = p1.GetCardFromHand(arrow3.GetID())
		assert.False(t, ok, "Arrow should have been discarded after attack")
		_, ok = p1.GetCardFromHand(sword4.GetID())
		assert.False(t, ok, "Sword should have been discarded after attack")

		assert.True(t, foundInCemetery(g, target), "Cemetery should contain the dead dragon")
		assert.True(t, foundInDiscardPile(g, poison1), "Discard pile should contain the used poison")
		assert.True(t, foundInDiscardPile(g, sword2), "Discard pile should contain the used sword")
		assert.True(t, foundInDiscardPile(g, arrow3), "Discard pile should contain the used arrow")
		assert.True(t, foundInDiscardPile(g, sword4), "Discard pile should contain the used sword")
	})
}

func foundInCemetery(g *Game, a ports.Warrior) bool {
	c := g.cemetery.(*cemetery)
	for _, w := range c.corps {
		if w == a || (w != nil && w.GetID() == a.GetID()) {
			return true
		}
	}
	return false
}

func foundInDiscardPile(g *Game, a ports.Card) bool {
	dp := g.discardPile.(*discardPile)
	for _, card := range dp.cards {
		if card == a || (card != nil && card.GetID() == a.GetID()) {
			return true
		}
	}
	return false
}

func newPlayerWithCardAndObserver(name string, cardsInHand []ports.Card,
	cardsInField []ports.Warrior, game *Game) ports.Player {
	p := &player{
		name:                           name,
		cardMovedToPileObserver:        game,
		warriorMovedToCemeteryObserver: game,
		hand: &hand{
			cards: cardsInHand,
		},
		field: &field{
			playerName:        name,
			cards:             cardsInField,
			gameEndedObserver: game,
		},
		castle: &castle{},
	}

	for _, card := range cardsInField {
		card.AddCardMovedToPileObserver(p)
		targ, ok := card.(ports.Warrior)
		if ok {
			targ.AddWarriorDeadObserver(p)
		}
	}
	for _, card := range cardsInHand {
		card.AddCardMovedToPileObserver(p)
	}

	return p
}
