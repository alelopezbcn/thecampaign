package main

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/alelopezbcn/thecampaign/internal/domain"
	_ "github.com/mattn/go-sqlite3"
)

func main() {
	reader := bufio.NewReader(os.Stdin)

	fmt.Println("Insert the name of the player 1:")
	p1, err := reader.ReadString('\n')
	if err != nil {
		fmt.Println("Error reading player1:", err)
		os.Exit(-1)
	}

	fmt.Println("Insert the name of the player 2:")
	p2, err := reader.ReadString('\n')
	if err != nil {
		fmt.Println("Error reading player2:", err)
		os.Exit(-1)
	}

	g := domain.NewGame(strings.TrimSpace(p1), strings.TrimSpace(p2))

	showCurrentPlayerHand(g)

	println(g.WhoIsNext().Name + " Insert comma separated the Initial warriors:")
	w1, err := reader.ReadString('\n')
	if err != nil {
		fmt.Println("Error reading warriors for player1:", err)
		os.Exit(-1)
	}
	warriors1 := strings.Split(strings.TrimSpace(w1), ",")
	err = g.SetInitialWarriors(g.Players[g.CurrentTurn].ID, warriors1)
	if err != nil {
		fmt.Println("Error setting initial warriors for player1:", err)
		os.Exit(-1)
	}

	showCurrentPlayerHand(g)
	println(g.WhoIsNext().Name + " Insert comma separated the Initial warriors:")
	w2, err := reader.ReadString('\n')
	if err != nil {
		fmt.Println("Error reading warriors for player1:", err)
		os.Exit(-1)
	}
	warriors2 := strings.Split(strings.TrimSpace(w2), ",")
	err = g.SetInitialWarriors(g.Players[g.CurrentTurn].ID, warriors2)
	if err != nil {
		fmt.Println("Error setting initial warriors for player1:", err)
		os.Exit(-1)
	}

	showCurrentPlayerHand(g)
	println(g.WhoIsNext().Name + " Draw a card")

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

func showCurrentPlayerHand(g *domain.Game) {
	next := g.WhoIsNext()
	println("****************************************")
	println(next.Name + "'s Hand:")
	for _, c := range next.ShowHand() {
		println(fmt.Sprintf("- %s", c.String()))
	}
	println("****************************************")
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
