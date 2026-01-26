package main

import (
	"bufio"
	"errors"
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/alelopezbcn/thecampaign/internal/domain"
	"github.com/alelopezbcn/thecampaign/internal/domain/cards"
	"github.com/alelopezbcn/thecampaign/internal/domain/gamestatus"
	"github.com/alelopezbcn/thecampaign/internal/domain/ports"
	"github.com/alelopezbcn/thecampaign/internal/domain/types"
	_ "github.com/mattn/go-sqlite3"
)

const (
	currentHandHeader = "****************************************"
	newPhaseHeader    = "----------------------------------------------------------"
	attackAction      = 1
	spyAction         = 2
	stealAction       = 3
	buyAction         = 4
	constructAction   = 5
	tradeAction       = 6
	moveWarriorAction = 7
	catapultAction    = 8
	skipPhaseAction   = 9
	endTurnAction     = 10
)

var (
	reader *bufio.Reader
	g      *domain.Game
)

func main() {
	reader = bufio.NewReader(os.Stdin)
	var err error
	g, err = startGame()
	if err != nil {
		println("Error starting game:", err.Error())
		os.Exit(-1)
	}

	if err = setInitialWarriors(); err != nil {
		println("Error setting initial warriors:", err.Error())
		os.Exit(-1)
	}

	gameEnded := false
	for !gameEnded {
		status, err := drawACard()
		if err != nil {
			println("Error drawing a card:", err.Error())
			os.Exit(-1)
		}

		if err = playTurn(status); err != nil {
			println("Error performing action:", err.Error())
			os.Exit(-1)
		}
		gameEnded = g.IsGameEnded()
	}
	println("GAME ENDED!")
}

func startGame() (*domain.Game, error) {
	p1 := "Alelo"
	p2 := "Matuelo"
	return domain.NewGame(p1, p2, cards.NewDealer()), nil
}

func setInitialWarriors() error {
	for i := 0; i < 2; i++ {
		ok := false
		for !ok {
			current, _ := g.WhoIsCurrent()
			printTurnHeader(current.Name(), "SET INITIAL WARRIORS")
			showCurrentPlayerHand(current)

			print(current.Name() + " Insert the Initial warriors: ")
			w, err := reader.ReadString('\n')
			if err != nil {
				return fmt.Errorf("error reading warriors for player %s: %w",
					current.Name(), err)
			}
			warriors := strings.Split(strings.TrimSpace(w), ",")
			err = g.SetInitialWarriors(current.Name(), warriors)
			if err != nil {
				println(fmt.Sprintf("error setting initial warriors for player %s: %s",
					current.Name(), err.Error()))
				continue
			}
			ok = true
		}
	}

	return nil
}

func showCurrentPlayerHand(player ports.Player) {
	println(currentHandHeader)
	println(player.Name() + "'s Hand:")
	for _, c := range player.Hand().ShowCards() {
		println(fmt.Sprintf("  - %s", c.String()))
	}
	println(currentHandHeader)
	println()
}

func drawACard() (gamestatus.GameStatus, error) {
	current, _ := g.WhoIsCurrent()
	status, err := g.DrawCard(current.Name())
	if err != nil {
		if errors.Is(err, domain.ErrHandLimitExceeded) {
			msg := "DRAW A CARD: player can't take more cards"
			printTurnHeader(current.Name(), msg)
			// Return a status even on error
			enemy, _ := g.WhoIsCurrent()
			return gamestatus.NewGameStatus(current, enemy, g.CurrentAction()), nil
		}

		return status, fmt.Errorf("error drawing a card for player %s: %w",
			current.Name(), err)
	}

	printTurnHeader(current.Name(), "DRAW A CARD")

	return status, nil
}

func playTurn(status gamestatus.GameStatus) error {
	for status.CurrentAction != string(types.ActionTypeEndTurn) {
		println("\nBoard Status for", status.CurrentPlayer)
		println("Current Phase:", status.CurrentAction)
		println()
		println(status.ShowBoard())

		println("Available Actions:")
		printAvailableActions(status)
		print("Select an action: ")

		okOpt := false
		opt := 0
		for !okOpt {
			actionInput, err := reader.ReadString('\n')
			if err != nil {
				return fmt.Errorf("error reading action input: %w", err)
			}

			opt, err = strconv.Atoi(strings.TrimSpace(actionInput))
			if err != nil || opt < 1 || opt > 10 {
				println("Invalid action. Please select a valid option.")
				continue
			}

			okOpt = true
		}

		var err error
		switch opt {
		case attackAction:
			status, err = attack(status.CurrentPlayer)
		case spyAction:
			status, err = spy(status.CurrentPlayer)
		case stealAction:
			status, err = steal(status)
		case buyAction:
			status, err = buy(status.CurrentPlayer)
		case constructAction:
			status, err = construct(status.CurrentPlayer)
		case tradeAction:
			status, err = trade(status.CurrentPlayer)
		case moveWarriorAction:
			status, err = moveWarrior(status.CurrentPlayer)
		case catapultAction:
			status, err = catapult(status)
		case skipPhaseAction:
			status, err = g.SkipPhase(status.CurrentPlayer)
		case endTurnAction:
			status, err = g.EndTurn(status.CurrentPlayer)
		default:
			println("Action not yet implemented.")
		}

		if err != nil {
			println("Error:", err.Error())
		}
	}

	// End the turn
	_, err := g.EndTurn(status.CurrentPlayer)
	return err
}

func printAvailableActions(status gamestatus.GameStatus) {
	currentAction := status.CurrentAction

	// Always available actions
	if status.CanMoveWarrior {
		println(fmt.Sprintf("  %d. Move warrior to Field", moveWarriorAction))
	}
	if len(status.CurrentPlayerHand) >= 3 {
		println(fmt.Sprintf("  %d. Trade (3 cards)", tradeAction))
	}

	// Phase-specific actions
	switch currentAction {
	case string(types.ActionTypeAttack):
		println(fmt.Sprintf("  %d. Attack with Weapon", attackAction))
		println(fmt.Sprintf("  %d. Attack with Catapult", catapultAction))
	case string(types.ActionTypeSpySteal):
		println(fmt.Sprintf("  %d. Spy", spyAction))
		println(fmt.Sprintf("  %d. Steal", stealAction))
	case string(types.ActionTypeBuy):
		println(fmt.Sprintf("  %d. Buy", buyAction))
	case string(types.ActionTypeConstruct):
		println(fmt.Sprintf("  %d. Construct", constructAction))
	}

	println(fmt.Sprintf("  %d. Skip Phase", skipPhaseAction))
	println(fmt.Sprintf("  %d. End Turn", endTurnAction))
}

func steal(status gamestatus.GameStatus) (gamestatus.GameStatus, error) {
	print(fmt.Sprintf("The enemy has %d cards in hand. Choose one: ",
		status.CardsInEnemyHand))

	for {
		w, err := reader.ReadString('\n')
		if err != nil {
			return status, fmt.Errorf("error reading resource: %w", err)
		}
		position, err := strconv.Atoi(strings.TrimSpace(w))
		if err != nil || position < 1 || position > status.CardsInEnemyHand {
			println("Invalid position. Please select a valid option.")
			continue
		}

		return g.Steal(status.CurrentPlayer, position)
	}
}

func buy(player string) (gamestatus.GameStatus, error) {
	print("Select the resource for buying: ")
	for {
		w, err := reader.ReadString('\n')
		if err != nil {
			return gamestatus.GameStatus{}, fmt.Errorf("error reading resource: %w", err)
		}
		resourceID := strings.TrimSpace(w)

		status, err := g.Buy(player, resourceID)
		if err != nil {
			println("Error buying:", err.Error())
			continue
		}
		return status, nil
	}
}

func construct(player string) (gamestatus.GameStatus, error) {
	print("Select the resource for constructing: ")
	for {
		w, err := reader.ReadString('\n')
		if err != nil {
			return gamestatus.GameStatus{}, fmt.Errorf("error reading resource: %w", err)
		}
		resourceID := strings.TrimSpace(w)

		status, err := g.Construct(player, resourceID)
		if err != nil {
			println("Error constructing:", err.Error())
			continue
		}
		return status, nil
	}
}

func spy(player string) (gamestatus.GameStatus, error) {
	println("1- Spy next 5 cards from the deck")
	println("2- Spy enemy's cards")
	for {
		w, err := reader.ReadString('\n')
		if err != nil {
			return gamestatus.GameStatus{}, fmt.Errorf("error reading resource: %w", err)
		}
		opt, err := strconv.Atoi(strings.TrimSpace(w))
		if err != nil || opt < 1 || opt > 2 {
			println("Please select a valid option.")
			continue
		}

		cards, status, err := g.Spy(player, opt)
		if err != nil {
			println("Error spying:", err.Error())
			continue
		}

		println("Spied Cards:")
		for _, c := range cards {
			println(fmt.Sprintf("  - %s", c.String()))
		}
		return status, nil
	}
}

func moveWarrior(player string) (gamestatus.GameStatus, error) {
	for {
		print("Select the warrior to move to field: ")
		w, err := reader.ReadString('\n')
		if err != nil {
			return gamestatus.GameStatus{}, fmt.Errorf("error reading move warrior: %w", err)
		}
		warriorID := strings.TrimSpace(w)

		status, err := g.MoveWarriorToField(player, warriorID)
		if err != nil {
			println("Error moving warrior to field:", err.Error())
			continue
		}
		return status, nil
	}
}

func catapult(status gamestatus.GameStatus) (gamestatus.GameStatus, error) {
	print(fmt.Sprintf("The enemy's castle has %d resource cards. Choose one: ",
		status.ResourceCardsInEnemyCastle))

	for {
		w, err := reader.ReadString('\n')
		if err != nil {
			return status, fmt.Errorf("error reading resource: %w", err)
		}
		position, err := strconv.Atoi(strings.TrimSpace(w))
		if err != nil || position < 1 || position > status.ResourceCardsInEnemyCastle {
			println("Invalid position. Please select a valid option.")
			continue
		}

		return g.Catapult(status.CurrentPlayer, position)
	}
}

func trade(player string) (gamestatus.GameStatus, error) {
	for {
		print("Select the cards (3) to trade (comma separated): ")
		w, err := reader.ReadString('\n')
		if err != nil {
			return gamestatus.GameStatus{}, fmt.Errorf("error reading trade cards: %w", err)
		}
		cardIDs := strings.Split(strings.TrimSpace(w), ",")

		status, err := g.Trade(player, cardIDs)
		if err != nil {
			println("Error performing trade:", err.Error())
			continue
		}
		return status, nil
	}
}

func attack(playerName string) (gamestatus.GameStatus, error) {
	for {
		print("Select the target and the weapon (target,weapon): ")
		w, err := reader.ReadString('\n')
		if err != nil {
			return gamestatus.GameStatus{}, fmt.Errorf("error reading attack: %w", err)
		}
		cards := strings.Split(strings.TrimSpace(w), ",")
		if len(cards) != 2 {
			println("Invalid input. Please provide target and weapon.")
			continue
		}

		status, err := g.Attack(playerName, cards[0], cards[1])
		if err != nil {
			println("Error performing attack:", err.Error())
			continue
		}
		return status, nil
	}
}

func printTurnHeader(player string, action string) {
	println()
	println(newPhaseHeader)
	println(fmt.Sprintf("%s's TURN - %s", player, action))
	println(newPhaseHeader)
	println()
}
