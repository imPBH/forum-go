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
	"strings"
	"time"
)

type Post struct {
	Id         int
	Username   string
	Title      string
	Categories []string
	Content    string
	CreatedAt  string
	UpVotes    int
	DownVotes  int
	Comments   []Comment
}

type Comment struct {
	Id        int
	PostId    int
	Username  string
	Content   string
	CreatedAt string
}

// Database
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

	createUsersTable(database)
	createPostTable(database)
	createCommentTable(database)
	createVoteTable(database)
	createCategoriesTable(database)
	createCategories(database)

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
	router.HandleFunc("/posts", getPostsApi)
	router.HandleFunc("/api/vote", voteApi)
	router.HandleFunc("/filter", getPostsByApi)
	router.HandleFunc("/api/logout", logout)

	router.Handle("/templates/", http.StripPrefix("/templates/", fs))
	http.ListenAndServe(":8000", router)
}

// index displays the index page
func index(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		http.NotFound(w, r)
		return
	}
	if isLoggedIn(r) {
		cookie, _ := r.Cookie("SESSION")
		t, _ := template.ParseGlob("templates/*.html")
		t.ExecuteTemplate(w, "createpost.html", getUser(cookie.Value))
		return
	}
	t, _ := template.ParseGlob("templates/*.html")
	t.ExecuteTemplate(w, "index.html", "")
	return
}

// hashPassword hashes the password
func hashPassword(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), 14)
	return string(bytes), err
}

// addUser adds a user to the database
func addUser(database *sql.DB, username string, email string, password string, cookie string, expires string) {
	password, _ = hashPassword(password)
	statement, _ := database.Prepare("INSERT INTO users (username, email, password, cookie, expires) VALUES (?, ?, ?, ?, ?)")
	statement.Exec(username, email, password, cookie, expires)
	now := time.Now().Format("2006-01-02 15:04:05")
	fmt.Println("Added user: " + username + " with email: " + email + " at " + now)
}

// registerApi handles the register api
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
	if !emailNotTaken(email) {
		http.Redirect(w, r, "/register?err=email_taken", http.StatusFound)
		return
	}
	if !usernameNotTaken(username) {
		http.Redirect(w, r, "/register?err=username_taken", http.StatusFound)
		return
	}
	addUser(database, username, email, password, value, expiration.Format("2006-01-02 15:04:05"))
	cookie := http.Cookie{Name: "SESSION", Value: value, Expires: expiration, Path: "/"}
	http.SetCookie(w, &cookie)
	http.Redirect(w, r, "/", http.StatusFound)
	return

}

//loginApi handles the login api
func loginApi(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		fmt.Fprintf(w, "ParseForm() err: %v", err)
		return
	}
	submittedEmail := r.FormValue("email")
	submittedPassword := r.FormValue("password")
	rows, _ := database.Query("SELECT username, email, password FROM users WHERE email = ?", submittedEmail)
	var username string
	var email string
	var password string
	for rows.Next() {
		rows.Scan(&username, &email, &password)
	}
	now := time.Now().Format("2006-01-02 15:04:05")
	if username == "" && email == "" && password == "" {
		fmt.Println("Login failed (email not found) for " + submittedEmail + " at " + now)
		http.Redirect(w, r, "/login?err=invalid_email", http.StatusFound)
		return
	}
	if err := bcrypt.CompareHashAndPassword([]byte(password), []byte(submittedPassword)); err != nil {
		fmt.Println("Login failed (wrong password) for " + submittedEmail + " at " + now)
		http.Redirect(w, r, "/login?err=invalid_password", http.StatusFound)
		return
	}
	expiration := time.Now().Add(31 * 24 * time.Hour)
	value := uuid.NewV4().String()
	cookie := http.Cookie{Name: "SESSION", Value: value, Expires: expiration, Path: "/"}
	http.SetCookie(w, &cookie)
	// update cookie in DB
	statement, _ := database.Prepare("UPDATE users SET cookie = ?, expires = ? WHERE email = ?")
	statement.Exec(value, expiration.Format("2006-01-02 15:04:05"), email)
	fmt.Println("Logged in user: " + username + " with email: " + email + " at " + now)
	http.Redirect(w, r, "/", http.StatusFound)
	return
}

// register displays the register page
func register(w http.ResponseWriter, r *http.Request) {
	t, _ := template.ParseGlob("templates/*.html")
	t.ExecuteTemplate(w, "register.html", "")
}

// login displays template for the login page
func login(w http.ResponseWriter, r *http.Request) {
	t, _ := template.ParseGlob("templates/*.html")
	t.ExecuteTemplate(w, "login.html", "")
}

// isExpired returns true if the cookie has expired
func isExpired(expires string) bool {
	expiresTime, _ := time.Parse("2006-01-02 15:04:05", expires)
	return time.Now().After(expiresTime)
}

// emailNotTaken returns true if the email is not taken
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

// usernameNotTaken returns true if the username is not taken
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

// createPostTable create post table
func createPostTable(database *sql.DB) {
	statement, _ := database.Prepare("CREATE TABLE IF NOT EXISTS posts (id INTEGER PRIMARY KEY AUTOINCREMENT, username TEXT, title TEXT, categories TEXT, content TEXT, created_at TEXT, upvotes INTEGER, downvotes INTEGER)")
	statement.Exec()
}

// getUser get user by cookie
func getUser(cookie string) string {
	rows, _ := database.Query("SELECT username FROM users WHERE cookie = ?", cookie)
	var username string
	for rows.Next() {
		rows.Scan(&username)
	}
	return username
}

// createPostApi creates a post
func createPostApi(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	if err := r.ParseForm(); err != nil {
		fmt.Fprintf(w, "ParseForm() err: %v", err)
		return
	}
	if !isLoggedIn(r) {
		http.Redirect(w, r, "/login", http.StatusFound)
		return
	}
	cookie, _ := r.Cookie("SESSION")
	username := getUser(cookie.Value)
	title := r.FormValue("title")
	content := r.FormValue("content")
	categories := r.Form["categories[]"]
	validCategories := getCategories(database)
	for _, category := range categories {
		// if string not in array, return error
		if !inArray(category, validCategories) {
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte("Invalid category : " + category))
			return
		}
	}
	stringCategories := strings.Join(categories, ",")
	createdAt := time.Now().Format("2006-01-02 15:04:05")
	statement, _ := database.Prepare("INSERT INTO posts (username, title, categories, content, created_at, upvotes, downvotes) VALUES (?, ?, ?, ?, ?, ?, ?)")
	statement.Exec(username, title, stringCategories, content, createdAt, 0, 0)
	fmt.Println("Post created by " + username + " with title " + title + " at " + createdAt)
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Post created"))
	return
}

// createCommentTable creates a comment table
func createCommentTable(database *sql.DB) {
	statement, _ := database.Prepare("CREATE TABLE IF NOT EXISTS comments (id INTEGER PRIMARY KEY AUTOINCREMENT, username TEXT, post_id INTEGER, content TEXT, created_at TEXT)")
	statement.Exec()
}

// commentsApi creates a comment
func commentsApi(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
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
	fmt.Println("Comment created by " + username + " on post " + postId + " at " + createdAt)
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Comment created"))
}

// getPost by id returns a Post struct with the post data
func getPost(id string) Post {
	rows, _ := database.Query("SELECT username, title, categories, content, created_at FROM posts WHERE id = ?", id)
	var post Post
	post.Id, _ = strconv.Atoi(id)
	for rows.Next() {
		catString := ""
		rows.Scan(&post.Username, &post.Title, &catString, &post.Content, &post.CreatedAt)
		categoriesArray := strings.Split(catString, ",")
		post.Categories = categoriesArray
	}
	return post
}

// displayPost displays a post on a template
func displayPost(w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	id := r.URL.Query().Get("id")
	post := getPost(id)
	post.Comments = getComments(id)
	t, _ := template.ParseGlob("templates/*.html")
	t.ExecuteTemplate(w, "postTemplate.html", post)
}

// getComments get comments by post id
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

// getPosts get all posts
func getPosts() []Post {
	rows, _ := database.Query("SELECT id, username, title, content, created_at FROM posts")
	var posts []Post
	for rows.Next() {
		var post Post
		rows.Scan(&post.Id, &post.Username, &post.Title, &post.Content, &post.CreatedAt)
		posts = append(posts, post)
	}
	return posts
}

// getPostsApi display all posts on a template
func getPostsApi(w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	posts := getPosts()
	t, _ := template.ParseGlob("templates/*.html")
	t.ExecuteTemplate(w, "posts.html", posts)
}

// createVoteTable create the vote table into given database
func createVoteTable(database *sql.DB) {
	statement, _ := database.Prepare("CREATE TABLE IF NOT EXISTS votes (id INTEGER PRIMARY KEY AUTOINCREMENT, username TEXT, post_id INTEGER, vote INTEGER)")
	statement.Exec()
}

// voteApi api to vote on a post
func voteApi(w http.ResponseWriter, r *http.Request) {
	if r.Method == "POST" {
		if err := r.ParseForm(); err != nil {
			fmt.Fprintf(w, "ParseForm() err: %v", err)
			return
		}
		cookie, _ := r.Cookie("SESSION")
		username := getUser(cookie.Value)
		postId := r.FormValue("postId")
		postIdInt, _ := strconv.Atoi(postId)
		vote := r.FormValue("vote")
		voteInt, _ := strconv.Atoi(vote)
		now := time.Now().Format("2006-01-02 15:04:05")
		if voteInt == 1 {
			if hasUpvoted(username, postIdInt) {
				removeVote(database, postIdInt, username)
				decreaseUpvotes(database, postIdInt)
				fmt.Println("Removed upvote from " + username + " on post " + postId + " at " + now)
				w.WriteHeader(http.StatusOK)
				w.Write([]byte("Vote removed"))
				return
			}
			if hasDownvoted(username, postIdInt) {
				decreaseDownvotes(database, postIdInt)
				increaseUpvotes(database, postIdInt)
				updateVote(database, postIdInt, username, 1)
				fmt.Println(username + " upvoted" + " on post " + postId + " at " + now)
				w.WriteHeader(http.StatusOK)
				w.Write([]byte("Upvote added"))
				return
			}
			increaseUpvotes(database, postIdInt)
			addVote(database, postIdInt, username, 1)
			fmt.Println(username + " upvoted" + " on post " + postId + " at " + now)
			w.WriteHeader(http.StatusOK)
			w.Write([]byte("Upvote added"))
			return
		}
		if voteInt == -1 {
			if hasDownvoted(username, postIdInt) {
				removeVote(database, postIdInt, username)
				decreaseDownvotes(database, postIdInt)
				fmt.Println("Removed downvote from " + username + " on post " + postId + " at " + now)
				w.WriteHeader(http.StatusOK)
				w.Write([]byte("Vote removed"))
				return
			}
			if hasUpvoted(username, postIdInt) {
				decreaseUpvotes(database, postIdInt)
				increaseDownvotes(database, postIdInt)
				updateVote(database, postIdInt, username, -1)
				fmt.Println(username + " downvoted" + " on post " + postId + " at " + now)
				w.WriteHeader(http.StatusOK)
				w.Write([]byte("Downvote added"))
				return
			}
			increaseDownvotes(database, postIdInt)
			addVote(database, postIdInt, username, -1)
			fmt.Println(username + " downvoted" + " on post " + postId + " at " + now)
			w.WriteHeader(http.StatusOK)
			w.Write([]byte("Downvote added"))
			return
		}
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("Invalid vote"))
		return
	}
	w.WriteHeader(http.StatusMethodNotAllowed)
	return
}

// hasUpvoted check if user has upvoted a post
func hasUpvoted(username string, postId int) bool {
	rows, _ := database.Query("SELECT vote FROM votes WHERE username = ? AND post_id = ? AND vote = 1", username, postId)
	vote := 0
	for rows.Next() {
		rows.Scan(&vote)
	}
	if vote != 0 {
		return true
	}
	return false
}

// hasDownvoted check if user has downvoted a post
func hasDownvoted(username string, postId int) bool {
	rows, _ := database.Query("SELECT vote FROM votes WHERE username = ? AND post_id = ? AND vote = -1", username, postId)
	vote := 0
	for rows.Next() {
		rows.Scan(&vote)
	}
	if vote != 0 {
		return true
	}
	return false
}

// inArray check if a string is in an array
func inArray(input string, array []string) bool {
	for _, v := range array {
		if v == input {
			return true
		}
	}
	return false
}

// createCategoriesTable create the categories' table into given database
func createCategoriesTable(database *sql.DB) {
	statement, _ := database.Prepare("CREATE TABLE IF NOT EXISTS categories (id INTEGER PRIMARY KEY, name TEXT)")
	statement.Exec()
}

// createCategories creates categories in the database
func createCategories(database *sql.DB) {
	statement, _ := database.Prepare("INSERT INTO categories (name) SELECT ? WHERE NOT EXISTS (SELECT 1 FROM categories WHERE name = ?)")
	statement.Exec("General", "General")
	statement.Exec("Technology", "Technology")
	statement.Exec("Science", "Science")
	statement.Exec("Sports", "Sports")
	statement.Exec("Gaming", "Gaming")
	statement.Exec("Music", "Music")
	statement.Exec("Books", "Books")
	statement.Exec("Movies", "Movies")
	statement.Exec("TV", "TV")
	statement.Exec("Food", "Food")
	statement.Exec("Travel", "Travel")
	statement.Exec("Photography", "Photography")
	statement.Exec("Art", "Art")
	statement.Exec("Writing", "Writing")
	statement.Exec("Programming", "Programming")
	statement.Exec("Other", "Other")
}

// getCategories returns all categories
func getCategories(database *sql.DB) []string {
	rows, _ := database.Query("SELECT name FROM categories")
	var categories []string
	for rows.Next() {
		var name string
		rows.Scan(&name)
		categories = append(categories, name)
	}
	return categories
}

// getPostsByCategory returns all posts in a given category
func getPostsByCategory(category string) []Post {
	rows, _ := database.Query("SELECT id, username, title, categories, content, created_at, upvotes, downvotes  FROM posts WHERE categories LIKE ?", "%"+category+"%")
	var posts []Post
	for rows.Next() {
		var post Post
		var catString string
		rows.Scan(&post.Id, &post.Username, &post.Title, &catString, &post.Content, &post.CreatedAt, &post.UpVotes, &post.DownVotes)
		post.Categories = strings.Split(catString, ",")
		posts = append(posts, post)
	}
	return posts
}

// getPostsByUser returns all posts by a user
func getPostsByUser(username string) []Post {
	rows, _ := database.Query("SELECT id, username, title, categories, content, created_at, upvotes, downvotes  FROM posts WHERE username = ?", username)
	var posts []Post
	for rows.Next() {
		var post Post
		var catString string
		rows.Scan(&post.Id, &post.Username, &post.Title, &catString, &post.Content, &post.CreatedAt, &post.UpVotes, &post.DownVotes)
		post.Categories = strings.Split(catString, ",")
		posts = append(posts, post)
	}
	return posts
}

// getPostByApi gets all post filtered by the given parameters
func getPostsByApi(w http.ResponseWriter, r *http.Request) {
	method := r.URL.Query().Get("by")
	if method == "category" {
		category := r.URL.Query().Get("category")
		posts := getPostsByCategory(category)
		t, _ := template.ParseGlob("templates/*.html")
		t.ExecuteTemplate(w, "posts.html", posts)
		return
	}
	if method == "myposts" {
		if isLoggedIn(r) {
			cookie, _ := r.Cookie("SESSION")
			username := getUser(cookie.Value)
			posts := getPostsByUser(username)
			t, _ := template.ParseGlob("templates/*.html")
			t.ExecuteTemplate(w, "posts.html", posts)
			return
		}
		w.WriteHeader(http.StatusUnauthorized)
		w.Write([]byte("You must be logged in to view your posts"))
	}
	if method == "liked" {
		if isLoggedIn(r) {
			cookie, _ := r.Cookie("SESSION")
			username := getUser(cookie.Value)
			posts := getLikedPosts(username)
			t, _ := template.ParseGlob("templates/*.html")
			t.ExecuteTemplate(w, "posts.html", posts)
			return
		}
		w.WriteHeader(http.StatusUnauthorized)
		w.Write([]byte("You must be logged in to view your liked posts"))
	}
	w.WriteHeader(http.StatusBadRequest)
	w.Write([]byte("Invalid request"))
}

// getLikedPosts gets posts that user has liked
func getLikedPosts(username string) []Post {
	rows, _ := database.Query("SELECT id, username, title, categories, content, created_at, upvotes, downvotes  FROM posts WHERE id IN (SELECT post_id FROM votes WHERE username = ? AND vote = 1)", username)
	var posts []Post
	for rows.Next() {
		var post Post
		var catString string
		rows.Scan(&post.Id, &post.Username, &post.Title, &catString, &post.Content, &post.CreatedAt, &post.UpVotes, &post.DownVotes)
		post.Categories = strings.Split(catString, ",")
		posts = append(posts, post)
	}
	return posts
}

// isLoggedIn checks if the user is logged in
func isLoggedIn(r *http.Request) bool {
	cookie, err := r.Cookie("SESSION")
	if err != nil {
		return false
	}
	var cookieExists bool
	err = database.QueryRow("SELECT IIF(COUNT(*), 'true', 'false') FROM users WHERE cookie = ?", cookie.Value).Scan(&cookieExists)
	if err != nil {
		return false
	}
	if !cookieExists {
		return false
	}
	rows, _ := database.Query("SELECT expires FROM users WHERE cookie = ?", cookie.Value)
	var expires string
	for rows.Next() {
		rows.Scan(&expires)
	}
	if isExpired(expires) {
		return false
	}
	return true
}

// logout deletes the session cookie from the database
func logout(w http.ResponseWriter, r *http.Request) {
	cookie, _ := r.Cookie("SESSION")
	username := getUser(cookie.Value)
	now := time.Now().Format("2006-01-02 15:04:05")
	if cookie != nil {
		username := getUser(cookie.Value)
		statement, _ := database.Prepare("UPDATE users SET cookie = '', expires = '' WHERE username = ?")
		statement.Exec(username)
	}
	fmt.Println("User " + username + " logged out at " + now)
	http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
	return
}

// removeVote removes a vote from a post
func removeVote(database *sql.DB, postId int, username string) {
	statement, _ := database.Prepare("DELETE FROM votes WHERE post_id = ? AND username = ?")
	statement.Exec(postId, username)
}

// decreaseUpvotes decreases the upvotes of a post by 1
func decreaseUpvotes(database *sql.DB, postId int) {
	statement, _ := database.Prepare("UPDATE posts SET upvotes = upvotes - 1 WHERE id = ?")
	statement.Exec(postId)
}

// decreaseDownvotes decreases the downvotes of a post by 1
func decreaseDownvotes(database *sql.DB, postId int) {
	statement, _ := database.Prepare("UPDATE posts SET downvotes = downvotes - 1 WHERE id = ?")
	statement.Exec(postId)
}

// increaseUpvotes increases the upvotes of a post by 1
func increaseUpvotes(database *sql.DB, postId int) {
	statement, _ := database.Prepare("UPDATE posts SET upvotes = upvotes + 1 WHERE id = ?")
	statement.Exec(postId)
}

// increaseDownvotes increases the downvotes of a post by 1
func increaseDownvotes(database *sql.DB, postId int) {
	statement, _ := database.Prepare("UPDATE posts SET downvotes = downvotes + 1 WHERE id = ?")
	statement.Exec(postId)
}

// addVote adds a vote to the database
func addVote(database *sql.DB, postId int, username string, vote int) {
	statement, _ := database.Prepare("INSERT INTO votes (username, post_id, vote) VALUES (?, ?, ?)")
	statement.Exec(username, postId, vote)
}

// updateVote updates the vote of a user for a post
func updateVote(database *sql.DB, postId int, username string, vote int) {
	statement, _ := database.Prepare("UPDATE votes SET vote = ? WHERE post_id = ? AND username = ?")
	statement.Exec(vote, postId, username)
}

// createUsersTable creates the users table
func createUsersTable(database *sql.DB) {
	statement, _ := database.Prepare("CREATE TABLE IF NOT EXISTS users (id INTEGER PRIMARY KEY, username TEXT, email TEXT, password TEXT, cookie TEXT, expires TEXT)")
	statement.Exec()
}
