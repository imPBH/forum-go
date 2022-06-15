package databaseAPI

import (
	"database/sql"
	"fmt"
	_ "github.com/mattn/go-sqlite3"
	"golang.org/x/crypto/bcrypt"
	"time"
)

// AddUser adds a user to the database
func AddUser(database *sql.DB, username string, email string, password string, cookie string, expires string) {
	password, _ = hashPassword(password)
	statement, _ := database.Prepare("INSERT INTO users (username, email, password, cookie, expires) VALUES (?, ?, ?, ?, ?)")
	statement.Exec(username, email, password, cookie, expires)
	now := time.Now().Format("2006-01-02 15:04:05")
	fmt.Println("Added user: " + username + " with email: " + email + " at " + now)
}

// EmailNotTaken returns true if the email is not taken
func EmailNotTaken(database *sql.DB, email string) bool {
	rows, _ := database.Query("SELECT email FROM users WHERE email = ?", email)
	var emailExists string
	for rows.Next() {
		rows.Scan(&emailExists)
	}
	if emailExists == "" {
		return true
	}
	return false
}

// UsernameNotTaken returns true if the username is not taken
func UsernameNotTaken(database *sql.DB, username string) bool {
	rows, _ := database.Query("SELECT username FROM users WHERE username = ?", username)
	var usernameExists string
	for rows.Next() {
		rows.Scan(&usernameExists)
	}
	if usernameExists == "" {
		return true
	}
	return false
}

// CheckCookie checks if a cookie is valid
func CheckCookie(database *sql.DB, cookie string) bool {
	var result bool
	err := database.QueryRow("SELECT IIF(COUNT(*), 'true', 'false') FROM users WHERE cookie = ?", cookie).Scan(&result)
	if err != nil {
		return false
	}
	return result
}

// GetExpires returns the expiration date of a cookie
func GetExpires(database *sql.DB, cookie string) string {
	var expires string
	rows, _ := database.Query("SELECT expires FROM users WHERE cookie = ?", cookie)
	for rows.Next() {
		rows.Scan(&expires)
	}
	return expires
}

// Logout logs a user out
func Logout(database *sql.DB, username string) {
	statement, _ := database.Prepare("UPDATE users SET cookie = '', expires = '' WHERE username = ?")
	statement.Exec(username)
}

// UpdateCookie updates the cookie of a user
func UpdateCookie(database *sql.DB, token string, expiration time.Time, email string) {
	statement, _ := database.Prepare("UPDATE users SET cookie = ?, expires = ? WHERE email = ?")
	statement.Exec(token, expiration.Format("2006-01-02 15:04:05"), email)
}

// hashPassword hashes the password
func hashPassword(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), 14)
	return string(bytes), err
}
