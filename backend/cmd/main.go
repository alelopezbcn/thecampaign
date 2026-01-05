package main

import (
	"bufio"
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

	if err = drawACard(); err != nil {
		println("Error drawing a card:", err)
		os.Exit(-1)
	}
	if err = performAction(); err != nil {
		println("Error performing action:", err)
		os.Exit(-1)
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
	fmt.Print("Insert the name of the player 1: ")
	p1, err := reader.ReadString('\n')
	if err != nil {
		return nil, fmt.Errorf("error reading player1: %w", err)
	}
	p1 = strings.TrimSpace(p1)
	var p2 string

	ok := false
	for !ok {
		fmt.Print("Insert the name of the player 2: ")
		p2, err = reader.ReadString('\n')
		if err != nil {
			return nil, fmt.Errorf("error reading player2: %w", err)
		}
		p2 = strings.TrimSpace(p2)
		if p2 == p1 {
			fmt.Println("Player 2 name must be different from Player 1 name.")
			continue
		}
		ok = true
	}

	return domain.NewGame(p1, p2), nil
}

func setInitialWarriors() error {
	for i := 0; i < 2; i++ {
		next, _ := g.WhoIsNext()
		printTurnHeader(next.Name, "SET INITIAL WARRIORS")
		showCurrentPlayerHand(next)

		print(next.Name + " Insert the Initial warriors: ")
		w, err := reader.ReadString('\n')
		if err != nil {
			return fmt.Errorf("error reading warriors for player %s: %w",
				next.Name, err)
		}
		warriors := strings.Split(strings.TrimSpace(w), ",")
		err = g.SetInitialWarriors(next.Name, warriors)
		if err != nil {
			return fmt.Errorf("error setting initial warriors for player %s: %w",
				next.Name, err)
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
	next, _ := g.WhoIsNext()
	card, err := g.DrawCard(next.Name)
	if err != nil {
		return fmt.Errorf("error drawing a card for player %s: %w",
			next.Name, err)
	}

	msg := fmt.Sprintf("DRAW A CARD: %s", card.String())
	printTurnHeader(next.Name, msg)

	return nil
}

func performAction() error {
	status := g.GetStatusForNextPlayer()
	println("Board Status for", status.Player)
	println(status.String())
	println()

	println("Available Actions:")
	println("  1. Attack")
	println("  2. Spy / Steal")
	println("  3. Buy")
	println("  4. Construct")
	println("  5. Trade")
	println("  6. Play Warrior")
	print("Select an action: ")

	ok := false
	opt := 0
	for !ok {
		actionInput, err := reader.ReadString('\n')
		if err != nil {
			return fmt.Errorf("error reading action input: %w", err)
		}

		opt, err = strconv.Atoi(strings.TrimSpace(actionInput))
		if err != nil || opt < 1 || opt > 6 {
			println("Invalid action. Please select a valid option.")
			continue
		}
		ok = true
	}

	println("You selected action:", opt)
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
	}
	// Here you would call the appropriate method on g to perform the action
	// For example:
	// err = g.PerformAction(status.Player, opt, ...)
	// if err != nil {
	//     println("Error performing action:", err)
	//     continue
	// }

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
