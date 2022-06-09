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
	"strconv"
	"time"
)

type Post struct {
	Id        int
	Username  string
	Title     string
	Content   string
	CreatedAt string
	UpVotes   int
	DownVotes int
	Comments  []Comment
}

type Comment struct {
	Id        int
	PostId    int
	Username  string
	Content   string
	CreatedAt string
}

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

	createPostTable(database)
	createCommentTable(database)

	fs := http.FileServer(http.Dir("templates"))
	router := http.NewServeMux()
	fmt.Println("Starting server on port 8000")

	router.HandleFunc("/", index)
	router.HandleFunc("/register", register)
	router.HandleFunc("/login", login)
	router.HandleFunc("/api/register", registerApi)
	router.HandleFunc("/api/login", loginApi)
	router.HandleFunc("/api/createpost", createPostApi)
	router.HandleFunc("/api/comments", commentsApi)
	router.HandleFunc("/post", displayPost)

	router.Handle("/templates/", http.StripPrefix("/templates/", fs))
	http.ListenAndServe(":8000", router)
}

func index(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		http.NotFound(w, r)
		return
	}

	cookie, err := r.Cookie("SESSION")
	if err != nil {
		fmt.Println(err)
		t, _ := template.ParseGlob("templates/*.html")
		t.ExecuteTemplate(w, "index.html", "")
		return
	}

	var cookieExists bool
	err = database.QueryRow("SELECT IIF(COUNT(*), 'true', 'false') FROM users WHERE cookie = ?", cookie.Value).Scan(&cookieExists)
	if err != nil {
		fmt.Println(err)
		t, _ := template.ParseGlob("templates/*.html")
		t.ExecuteTemplate(w, "index.html", "")
		return
	}

	fmt.Println(cookie.Value)
	if cookieExists {
		rows, _ := database.Query("SELECT expires FROM users WHERE cookie = ?", cookie.Value)
		var expires string
		for rows.Next() {
			rows.Scan(&expires)
		}

		if isExpired(expires) {
			fmt.Println("Expired")
			t, _ := template.ParseGlob("templates/*.html")
			t.ExecuteTemplate(w, "index.html", "")
			return
		}

		fmt.Println("Not expired")
		rows, _ = database.Query("SELECT username FROM users WHERE cookie = ?", cookie.Value)
		var user string
		for rows.Next() {
			rows.Scan(&user)
			t, _ := template.ParseGlob("templates/*.html")
			t.ExecuteTemplate(w, "createpost.html", user)
		}
	} else {
		t, _ := template.ParseGlob("templates/*.html")
		t.ExecuteTemplate(w, "index.html", "")
	}
}

func HashPassword(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), 14)
	return string(bytes), err
}

func addUser(database *sql.DB, username string, email string, password string, cookie string, expires string) {
	password, _ = HashPassword(password)
	statement, _ := database.Prepare("INSERT INTO users (username, email, password, cookie, expires) VALUES (?, ?, ?, ?, ?)")
	statement.Exec(username, email, password, cookie, expires)
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
	value := uuid.NewV4().String()
	expiration := time.Now().Add(31 * 24 * time.Hour)
	if emailNotTaken(email) && usernameNotTaken(username) {
		addUser(database, username, email, password, value, expiration.Format("2006-01-02 15:04:05"))
		cookie := http.Cookie{Name: "SESSION", Value: value, Expires: expiration, Path: "/"}
		http.SetCookie(w, &cookie)
		http.Redirect(w, r, "/home", http.StatusFound)
	} else {
		if !emailNotTaken(email) {
			http.Redirect(w, r, "/register?err=email_taken", http.StatusFound)
		} else if !usernameNotTaken(username) {
			http.Redirect(w, r, "/register?err=username_taken", http.StatusFound)
		}
	}
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
	if username == "" && email == "" && password == "" {
		http.Redirect(w, r, "/login?err=invalid_email", http.StatusFound)
	}
	if err := bcrypt.CompareHashAndPassword([]byte(password), []byte(submittedPassword)); err != nil {
		http.Redirect(w, r, "/login?err=invalid_password", http.StatusFound)
	} else {
		expiration := time.Now().Add(31 * 24 * time.Hour)
		value := uuid.NewV4().String()
		cookie := http.Cookie{Name: "SESSION", Value: value, Expires: expiration, Path: "/"}
		http.SetCookie(w, &cookie)

		// update cookie in DB
		statement, _ := database.Prepare("UPDATE users SET cookie = ?, expires = ? WHERE email = ?")
		statement.Exec(value, expiration.Format("2006-01-02 15:04:05"), email)
		http.Redirect(w, r, "/home", http.StatusFound)
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

func isExpired(expires string) bool {
	expiresTime, _ := time.Parse("2006-01-02 15:04:05", expires)
	return time.Now().After(expiresTime)
}

func emailNotTaken(email string) bool {
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

func usernameNotTaken(username string) bool {
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

// create post table
func createPostTable(database *sql.DB) {
	statement, _ := database.Prepare("CREATE TABLE IF NOT EXISTS posts (id INTEGER PRIMARY KEY AUTOINCREMENT, username TEXT, title TEXT, content TEXT, created_at TEXT, upvotes INTEGER, downvotes INTEGER)")
	statement.Exec()
}

// get user by cookie
func getUser(cookie string) string {
	rows, _ := database.Query("SELECT username FROM users WHERE cookie = ?", cookie)
	var username string
	for rows.Next() {
		rows.Scan(&username)
	}
	return username
}

// create a post
func createPostApi(w http.ResponseWriter, r *http.Request) {
	if r.Method == "POST" {
		if err := r.ParseForm(); err != nil {
			fmt.Fprintf(w, "ParseForm() err: %v", err)
			return
		}

		cookie, _ := r.Cookie("SESSION")
		username := getUser(cookie.Value)

		fmt.Println("cookie: " + cookie.Value)
		fmt.Println("username: " + username)

		title := r.FormValue("title")
		content := r.FormValue("content")
		createdAt := time.Now().Format("2006-01-02 15:04:05")
		statement, _ := database.Prepare("INSERT INTO posts (username, title, content, created_at, upvotes, downvotes) VALUES (?, ?, ?, ?, ?, ?)")
		statement.Exec(username, title, content, createdAt, 0, 0)
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("Post created"))
	}
}

// create a comment table
func createCommentTable(database *sql.DB) {
	statement, _ := database.Prepare("CREATE TABLE IF NOT EXISTS comments (id INTEGER PRIMARY KEY AUTOINCREMENT, username TEXT, post_id INTEGER, content TEXT, created_at TEXT)")
	statement.Exec()
}

func commentsApi(w http.ResponseWriter, r *http.Request) {
	if r.Method == "POST" {
		if err := r.ParseForm(); err != nil {
			fmt.Fprintf(w, "ParseForm() err: %v", err)
			return
		}
		cookie, _ := r.Cookie("SESSION")
		username := getUser(cookie.Value)
		postId := r.FormValue("postId")
		content := r.FormValue("content")
		createdAt := time.Now().Format("2006-01-02 15:04:05")
		statement, _ := database.Prepare("INSERT INTO comments (username, post_id, content, created_at) VALUES (?, ?, ?, ?)")
		statement.Exec(username, postId, content, createdAt)
		w.WriteHeader(http.StatusOK)
	}
}

// get post by id return a Post struct with the post data
func getPost(id string) Post {
	rows, _ := database.Query("SELECT username, title, content FROM posts WHERE id = ?", id)
	var post Post
	post.Id, _ = strconv.Atoi(id)
	for rows.Next() {
		rows.Scan(&post.Username, &post.Title, &post.Content)
	}
	return post
}

func displayPost(w http.ResponseWriter, r *http.Request) {
	if r.Method == "GET" {
		id := r.URL.Query().Get("id")
		post := getPost(id)
		post.Comments = getComments(id)
		t, _ := template.ParseGlob("templates/*.html")
		t.ExecuteTemplate(w, "postTemplate.html", post)
	}
}

// get comments by post id
func getComments(id string) []Comment {
	rows, _ := database.Query("SELECT id, username, content FROM comments WHERE post_id = ?", id)
	var comments []Comment
	for rows.Next() {
		var comment Comment
		rows.Scan(&comment.Id, &comment.Username, &comment.Content)
		comments = append(comments, comment)
	}
	return comments
}
