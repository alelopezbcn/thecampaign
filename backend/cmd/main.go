package main

import (
	"database/sql"
	"log"

	_ "github.com/mattn/go-sqlite3"
)

// func main() {
// 	reader := bufio.NewReader(os.Stdin)
// 	fmt.Println("Insert the name of the player 1:")
// 	player1, err := reader.ReadString('\n')
// 	if err != nil {
// 		fmt.Println("Error reading player1:", err)
// 		os.Exit(-1)
// 	}
//
// 	fmt.Println("Insert the name of the player 2:")
// 	player2, err := reader.ReadString('\n')
// 	if err != nil {
// 		fmt.Println("Error reading player2:", err)
// 		os.Exit(-1)
// 	}
//
// 	for {
// 		input, err := reader.ReadString('\n')
// 		if err != nil {
// 			fmt.Println("Error reading input:", err)
// 			continue
// 		}
// 		input = strings.TrimSpace(input)
// 		if len(input) == 0 {
// 			continue
// 		}
// 		char := input[0]
// 		switch char {
// 		case 'q', 'Q':
// 			fmt.Println("Quitting...")
// 			return
// 		case 'a', 'A':
// 			fmt.Println("You pressed A!")
// 		case 'b', 'B':
// 			fmt.Println("You pressed B!")
// 		default:
// 			fmt.Printf("You pressed: %c\n", char)
// 		}
// 	}
// }

func main() {
	db, err := sql.Open("sqlite3", "./test.db")
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()
	sqlStmt := `
    CREATE TABLE IF NOT EXISTS users (
        id INTEGER NOT NULL PRIMARY KEY AUTOINCREMENT,
        name TEXT
    );
    `
	_, err = db.Exec(sqlStmt)
	if err != nil {
		log.Fatal(err)
	}
	log.Println("Table 'users' created successfully")
}
