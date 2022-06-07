package main

import (
	"database/sql"
	"fmt"
	_ "github.com/mattn/go-sqlite3"
	uuid "github.com/satori/go.uuid"
	"golang.org/x/crypto/bcrypt"
	"html/template"
	"net/http"
	"os"
	"time"
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
	}

	database, _ = sql.Open("sqlite3", "./database.db")
	statement, _ := database.Prepare("CREATE TABLE IF NOT EXISTS users (id INTEGER PRIMARY KEY, username TEXT, email TEXT, password TEXT, cookie TEXT, expires TEXT)")
	statement.Exec()

	fs := http.FileServer(http.Dir("templates"))
	router := http.NewServeMux()
	fmt.Println("Starting server on port 8000")

	router.HandleFunc("/", index)
	router.HandleFunc("/register", register)
	router.HandleFunc("/login", login)
	router.HandleFunc("/api/register", registerApi)
	router.HandleFunc("/api/login", loginApi)

	router.Handle("/templates/", http.StripPrefix("/templates/", fs))
	http.ListenAndServe(":8000", router)
}

func index(w http.ResponseWriter, r *http.Request) {
	t, _ := template.ParseGlob("templates/*.html")
	t.ExecuteTemplate(w, "index.html", "")
}

func HashPassword(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), 14)
	return string(bytes), err
}

func addUser(database *sql.DB, username string, email string, password string) {
	password, _ = HashPassword(password)
	statement, _ := database.Prepare("INSERT INTO users (username, email, password) VALUES (?, ?, ?)")
	statement.Exec(username, email, password)
	fmt.Println("username: " + username + " email: " + email + " password: " + password)
}

func registerApi(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		fmt.Fprintf(w, "ParseForm() err: %v", err)
		return
	}
	username := r.FormValue("username")
	email := r.FormValue("email")
	password := r.FormValue("password")
	addUser(database, username, email, password)
	fmt.Fprintf(w, "User registered successfully")
}

func loginApi(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		fmt.Fprintf(w, "ParseForm() err: %v", err)
		return
	}
	submittedEmail := r.FormValue("email")
	submittedPassword := r.FormValue("password")
	fmt.Println(submittedEmail)
	fmt.Println(submittedPassword)
	rows, _ := database.Query("SELECT username, email, password FROM users WHERE email = ?", submittedEmail)
	var username string
	var email string
	var password string
	for rows.Next() {
		rows.Scan(&username, &email, &password)
		fmt.Println(username + " : " + email + " " + password)
	}
	if err := bcrypt.CompareHashAndPassword([]byte(password), []byte(submittedPassword)); err != nil {
		fmt.Fprintf(w, "Invalid Password")
	} else {
		expiration := time.Now().Add(365 * 24 * time.Hour)
		value := uuid.NewV4().String()
		cookie := http.Cookie{Name: "SESSION", Value: value, Expires: expiration, Path: "/"}
		http.SetCookie(w, &cookie)
		fmt.Fprintf(w, "Success")

		// update cookie in DB
		statement, _ := database.Prepare("UPDATE users SET cookie = ?, expires = ? WHERE email = ?")
		statement.Exec(value, expiration.String(), email)
	}
}

func register(w http.ResponseWriter, r *http.Request) {
	t, _ := template.ParseGlob("templates/*.html")
	t.ExecuteTemplate(w, "register.html", "")
}

func login(w http.ResponseWriter, r *http.Request) {
	t, _ := template.ParseGlob("templates/*.html")
	t.ExecuteTemplate(w, "login.html", "")
}
