package databaseAPI

import (
	"database/sql"
	"fmt"
	_ "github.com/mattn/go-sqlite3"
	"golang.org/x/crypto/bcrypt"
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

// CreateUsersTable creates the users table
func CreateUsersTable(database *sql.DB) {
	statement, _ := database.Prepare("CREATE TABLE IF NOT EXISTS users (id INTEGER PRIMARY KEY, username TEXT, email TEXT, password TEXT, cookie TEXT, expires TEXT)")
	statement.Exec()
}

// CreatePostTable create post table
func CreatePostTable(database *sql.DB) {
	statement, _ := database.Prepare("CREATE TABLE IF NOT EXISTS posts (id INTEGER PRIMARY KEY AUTOINCREMENT, username TEXT, title TEXT, categories TEXT, content TEXT, created_at TEXT, upvotes INTEGER, downvotes INTEGER)")
	statement.Exec()
}

// CreateCommentTable creates a comment table
func CreateCommentTable(database *sql.DB) {
	statement, _ := database.Prepare("CREATE TABLE IF NOT EXISTS comments (id INTEGER PRIMARY KEY AUTOINCREMENT, username TEXT, post_id INTEGER, content TEXT, created_at TEXT)")
	statement.Exec()
}

// CreateVoteTable create the vote table into given database
func CreateVoteTable(database *sql.DB) {
	statement, _ := database.Prepare("CREATE TABLE IF NOT EXISTS votes (id INTEGER PRIMARY KEY AUTOINCREMENT, username TEXT, post_id INTEGER, vote INTEGER)")
	statement.Exec()
}

// CreateCategoriesTable create the categories' table into given database
func CreateCategoriesTable(database *sql.DB) {
	statement, _ := database.Prepare("CREATE TABLE IF NOT EXISTS categories (id INTEGER PRIMARY KEY, name TEXT)")
	statement.Exec()
}

// CreateCategories creates categories in the database
func CreateCategories(database *sql.DB) {
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

// GetUser get user by cookie
func GetUser(database *sql.DB, cookie string) string {
	rows, _ := database.Query("SELECT username FROM users WHERE cookie = ?", cookie)
	var username string
	for rows.Next() {
		rows.Scan(&username)
	}
	return username
}

// GetPost by id returns a Post struct with the post data
func GetPost(database *sql.DB, id string) Post {
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

// GetComments get comments by post id
func GetComments(database *sql.DB, id string) []Comment {
	rows, _ := database.Query("SELECT id, username, content FROM comments WHERE post_id = ?", id)
	var comments []Comment
	for rows.Next() {
		var comment Comment
		rows.Scan(&comment.Id, &comment.Username, &comment.Content)
		comments = append(comments, comment)
	}
	return comments
}

// GetPosts get all posts
func GetPosts(database *sql.DB) []Post {
	rows, _ := database.Query("SELECT id, username, title, content, created_at FROM posts")
	var posts []Post
	for rows.Next() {
		var post Post
		rows.Scan(&post.Id, &post.Username, &post.Title, &post.Content, &post.CreatedAt)
		posts = append(posts, post)
	}
	return posts
}

// HasUpvoted check if user has upvoted a post
func HasUpvoted(database *sql.DB, username string, postId int) bool {
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

// HasDownvoted check if user has downvoted a post
func HasDownvoted(database *sql.DB, username string, postId int) bool {
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

// GetCategories returns all categories
func GetCategories(database *sql.DB) []string {
	rows, _ := database.Query("SELECT name FROM categories")
	var categories []string
	for rows.Next() {
		var name string
		rows.Scan(&name)
		categories = append(categories, name)
	}
	return categories
}

// GetPostsByCategory returns all posts in a given category
func GetPostsByCategory(database *sql.DB, category string) []Post {
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

// GetPostsByUser returns all posts by a user
func GetPostsByUser(database *sql.DB, username string) []Post {
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

// GetLikedPosts gets posts that user has liked
func GetLikedPosts(database *sql.DB, username string) []Post {
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

// RemoveVote removes a vote from a post
func RemoveVote(database *sql.DB, postId int, username string) {
	statement, _ := database.Prepare("DELETE FROM votes WHERE post_id = ? AND username = ?")
	statement.Exec(postId, username)
}

// DecreaseUpvotes decreases the upvotes of a post by 1
func DecreaseUpvotes(database *sql.DB, postId int) {
	statement, _ := database.Prepare("UPDATE posts SET upvotes = upvotes - 1 WHERE id = ?")
	statement.Exec(postId)
}

// DecreaseDownvotes decreases the downvotes of a post by 1
func DecreaseDownvotes(database *sql.DB, postId int) {
	statement, _ := database.Prepare("UPDATE posts SET downvotes = downvotes - 1 WHERE id = ?")
	statement.Exec(postId)
}

// IncreaseUpvotes increases the upvotes of a post by 1
func IncreaseUpvotes(database *sql.DB, postId int) {
	statement, _ := database.Prepare("UPDATE posts SET upvotes = upvotes + 1 WHERE id = ?")
	statement.Exec(postId)
}

// IncreaseDownvotes increases the downvotes of a post by 1
func IncreaseDownvotes(database *sql.DB, postId int) {
	statement, _ := database.Prepare("UPDATE posts SET downvotes = downvotes + 1 WHERE id = ?")
	statement.Exec(postId)
}

// AddVote adds a vote to the database
func AddVote(database *sql.DB, postId int, username string, vote int) {
	statement, _ := database.Prepare("INSERT INTO votes (username, post_id, vote) VALUES (?, ?, ?)")
	statement.Exec(username, postId, vote)
}

// UpdateVote updates the vote of a user for a post
func UpdateVote(database *sql.DB, postId int, username string, vote int) {
	statement, _ := database.Prepare("UPDATE votes SET vote = ? WHERE post_id = ? AND username = ?")
	statement.Exec(vote, postId, username)
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

// CreatePost
func CreatePost(database *sql.DB, username string, title string, categories string, content string, createdAt time.Time) {
	createdAtString := createdAt.Format("2006-01-02 15:04:05")
	statement, _ := database.Prepare("INSERT INTO posts (username, title, categories, content, created_at, upvotes, downvotes) VALUES (?, ?, ?, ?, ?, ?, ?)")
	statement.Exec(username, title, categories, content, createdAtString, 0, 0)
}

// AddComment adds a comment to a post
func AddComment(database *sql.DB, username string, postId int, content string, createdAt time.Time) {
	createdAtString := createdAt.Format("2006-01-02 15:04:05")
	statement, _ := database.Prepare("INSERT INTO comments (username, post_id, content, created_at) VALUES (?, ?, ?, ?)")
	statement.Exec(username, postId, content, createdAtString)
}

// hashPassword hashes the password
func hashPassword(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), 14)
	return string(bytes), err
}
