package main

import (
	"database/sql"
	_ "github.com/mattn/go-sqlite3"
	"os"
)

var database *sql.DB

func main() {
	// check if DB exists
	var _, err = os.Stat("database.db")

	// create DB if not exists
	if os.IsNotExist(err) {
		var file, err = os.Create("database.db")
		if err != nil {
			return
		}
		defer file.Close()
		database, _ = sql.Open("sqlite3", "./database.db")
		statement, _ := database.Prepare("CREATE TABLE IF NOT EXISTS users (id INTEGER PRIMARY KEY, username TEXT, email TEXT, password TEXT)")
		statement.Exec()
	}
}
