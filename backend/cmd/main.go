package main

import (
	"bufio"
	"errors"
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/alelopezbcn/thecampaign/internal/domain"
	_ "github.com/mattn/go-sqlite3"
)

const (
	currentHandHeader = "****************************************"
	newPhaseHeader    = "----------------------------------------------------------"
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
		println("Error starting game:", err)
		os.Exit(-1)
	}

	if err = setInitialWarriors(); err != nil {
		println("Error setting initial warriors:", err)
		os.Exit(-1)
	}

	gameEnded := false
	for !gameEnded {
		if err = drawACard(); err != nil {
			println("Error drawing a card:", err)
			os.Exit(-1)
		}

		if err = playTurn(); err != nil {
			println("Error performing action:", err)
			os.Exit(-1)
		}
		gameEnded = g.IsGameEnded()
	}
	println("HASTA AQUI LLEGUE")

	// for {
	// 	input, err := reader.ReadString('\n')
	// 	if err != nil {
	// 		fmt.Println("Error reading input:", err)
	// 		continue
	// 	}
	// 	input = strings.TrimSpace(input)
	// 	if len(input) == 0 {
	// 		continue
	// 	}
	// 	char := input[0]
	// 	switch char {
	// 	case 'q', 'Q':
	// 		fmt.Println("Quitting...")
	// 		return
	// 	case 'a', 'A':
	// 		fmt.Println("You pressed A!")
	// 	case 'b', 'B':
	// 		fmt.Println("You pressed B!")
	// 	default:
	// 		fmt.Printf("You pressed: %c\n", char)
	// 	}
	// }
}

func startGame() (*domain.Game, error) {
	// fmt.Print("Insert the name of the player 1: ")
	// p1, err := reader.ReadString('\n')
	// if err != nil {
	// 	return nil, fmt.Errorf("error reading player1: %w", err)
	// }
	// p1 = strings.TrimSpace(p1)
	// var p2 string
	//
	// ok := false
	// for !ok {
	// 	fmt.Print("Insert the name of the player 2: ")
	// 	p2, err = reader.ReadString('\n')
	// 	if err != nil {
	// 		return nil, fmt.Errorf("error reading player2: %w", err)
	// 	}
	// 	p2 = strings.TrimSpace(p2)
	// 	if p2 == p1 {
	// 		fmt.Println("Player 2 name must be different from Player 1 name.")
	// 		continue
	// 	}
	// 	ok = true
	// }

	p1 := "Alelo"
	p2 := "Matuelo"
	return domain.NewGame(p1, p2), nil
}

func setInitialWarriors() error {
	for i := 0; i < 2; i++ {
		ok := false
		for !ok {
			current, _ := g.WhoIsCurrent()
			printTurnHeader(current.Name, "SET INITIAL WARRIORS")
			showCurrentPlayerHand(current)

			print(current.Name + " Insert the Initial warriors: ")
			w, err := reader.ReadString('\n')
			if err != nil {
				return fmt.Errorf("error reading warriors for player %s: %w",
					current.Name, err)
			}
			warriors := strings.Split(strings.TrimSpace(w), ",")
			err = g.SetInitialWarriors(current.Name, warriors)
			if err != nil {
				println(fmt.Sprintf("error setting initial warriors for player %s: %s",
					current.Name, err.Error()))
				continue
			}
			ok = true
		}
	}

	return nil
}

func showCurrentPlayerHand(player *domain.Player) {
	println(currentHandHeader)
	println(player.Name + "'s Hand:")
	for _, c := range player.ShowHand() {
		println(fmt.Sprintf("  - %s", c.String()))
	}
	println(currentHandHeader)
	println()
}

func drawACard() error {
	current, _ := g.WhoIsCurrent()
	card, err := g.DrawCard(current.Name)
	if err != nil {
		if errors.Is(err, domain.ErrHandLimitExceeded) {
			msg := "DRAW A CARD: player can't take more cards"
			printTurnHeader(current.Name, msg)
			return nil
		}

		return fmt.Errorf("error drawing a card for player %s: %w",
			current.Name, err)
	}

	msg := fmt.Sprintf("DRAW A CARD: %s", card.String())
	printTurnHeader(current.Name, msg)

	return nil
}

func playTurn() error {
	actionsPerformed := map[int]bool{}
	actionsPending := 8

	var status domain.BoardStatus

	for actionsPending > 0 {
		status = g.GetStatusForNextPlayer()
		println("\nBoard Status for", status.Player)
		println()
		println(status.String())

		println("Available Actions:")
		hasAttacked, ok := actionsPerformed[1]
		if !ok || !hasAttacked {
			println("  1. Attack")
		}
		hasSpied, ok := actionsPerformed[2]
		if !ok || !hasSpied {
			println("  2. Spy")
		}
		hasStolen, ok := actionsPerformed[3]
		if !ok || !hasStolen {
			println("  3. Steal")
		}
		hasBought, ok := actionsPerformed[4]
		if !ok || !hasBought {
			println("  4. Buy")
		}
		hasConstructed, ok := actionsPerformed[5]
		if !ok || !hasConstructed {
			println("  5. Construct")
		}
		hasTraded, ok := actionsPerformed[6]
		if !ok || !hasTraded {
			println("  6. Trade")
		}
		hasPlayedWarrior, ok := actionsPerformed[7]
		if !ok || !hasPlayedWarrior {
			println("  7. Play warrior")
		}
		println("  8. Pass")
		print("Select an action: ")

		okOpt := false
		opt := 0
		for !okOpt {
			actionInput, err := reader.ReadString('\n')
			if err != nil {
				return fmt.Errorf("error reading action input: %w", err)
			}

			opt, err = strconv.Atoi(strings.TrimSpace(actionInput))
			if err != nil || opt < 1 || opt > 8 {
				println("Invalid action. Please select a valid option.")
				continue
			}

			_, alreadyDone := actionsPerformed[opt]
			if alreadyDone && opt != 8 {
				println("Action already performed this turn. Please select another action.")
				continue
			}

			okOpt = true
		}

		switch opt {
		case 1:
			ok = false
			for !ok {
				print("Select the warrior, the target and the weapon: ")
				w, err := reader.ReadString('\n')
				if err != nil {
					return fmt.Errorf("error reading attack: %w", err)
				}
				cards := strings.Split(strings.TrimSpace(w), ",")
				if len(cards) != 3 {
					println("Invalid input. Please provide attack, warrior and target.")
					continue
				}

				err = g.Attack(status.Player, cards[0], cards[1], cards[2])
				if err != nil {
					println("Error performing attack:", err)
					continue
				}
				ok = true
			}
			actionsPerformed[1] = true
			actionsPending--
		case 8:
			actionsPending = 0
		default:
			println("Action not yet implemented.")
		}
	}

	_ = g.EndTurn(status.Player)

	return nil
}

func printTurnHeader(player string, action string) {
	println()
	println(newPhaseHeader)
	println(fmt.Sprintf("%s's TURN - %s", player, action))
	println(newPhaseHeader)
	println()
}

// func main() {
// 	db, err := sql.Open("sqlite3", "./test.db")
// 	if err != nil {
// 		log.Fatal(err)
// 	}
// 	defer db.Close()
// 	sqlStmt := `
//     CREATE TABLE IF NOT EXISTS users (
//         id INTEGER NOT NULL PRIMARY KEY AUTOINCREMENT,
//         name TEXT
//     );
//     `
// 	_, err = db.Exec(sqlStmt)
// 	if err != nil {
// 		log.Fatal(err)
// 	}
// 	log.Println("Table 'users' created successfully")
// }
