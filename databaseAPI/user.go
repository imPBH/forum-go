package databaseAPI

import (
	"database/sql"
	_ "github.com/mattn/go-sqlite3"
)

type User struct {
	isLoggedIn bool
	username   string
}

// GetUser get user by cookie
func GetUser(database *sql.DB, cookie string) string {
	rows, _ := database.Query("SELECT username FROM users WHERE cookie = ?", cookie)
	var username string
	for rows.Next() {
		rows.Scan(&username)
	}
	return username
}

// GetUserInfo returns the username, email and hashed password of a user
func GetUserInfo(database *sql.DB, submittedEmail string) (string, string, string) {
	var user string
	var email string
	var password string
	rows, _ := database.Query("SELECT username, email, password FROM users WHERE email = ?", submittedEmail)
	for rows.Next() {
		rows.Scan(&user, &email, &password)
	}
	return user, email, password
}
