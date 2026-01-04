package main

import (
	"bufio"
	"fmt"
	"os"
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
	fmt.Println("Insert the name of the player 1:")
	p1, err := reader.ReadString('\n')
	if err != nil {
		return nil, fmt.Errorf("error reading player1: %w", err)
	}

	fmt.Println("Insert the name of the player 2:")
	p2, err := reader.ReadString('\n')
	if err != nil {
		return nil, fmt.Errorf("error reading player2: %w", err)
	}

	p1 = strings.TrimSpace(p1)
	p2 = strings.TrimSpace(p2)

	g := domain.NewGame(p1, p2)

	return g, nil
}

func setInitialWarriors() error {
	for i := 0; i < 2; i++ {
		next := g.WhoIsNext()
		printTurnHeader(next.Name, "SET INITIAL WARRIORS")
		showCurrentPlayerHand(next)

		println(next.Name + " Insert comma separated the Initial warriors:")
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

	next := g.WhoIsNext()
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
