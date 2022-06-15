package databaseAPI

import (
	"database/sql"
	_ "github.com/mattn/go-sqlite3"
	"strconv"
	"strings"
	"time"
)

// GetPost by id returns a Post struct with the post data
func GetPost(database *sql.DB, id string) Post {
	rows, _ := database.Query("SELECT username, title, categories, content, created_at, upvotes, downvotes FROM posts WHERE id = ?", id)
	var post Post
	post.Id, _ = strconv.Atoi(id)
	for rows.Next() {
		catString := ""
		rows.Scan(&post.Username, &post.Title, &catString, &post.Content, &post.CreatedAt, &post.UpVotes, &post.DownVotes)
		categoriesArray := strings.Split(catString, ",")
		post.Categories = categoriesArray
	}
	return post
}

// GetComments get comments by post id
func GetComments(database *sql.DB, id string) []Comment {
	rows, _ := database.Query("SELECT id, username, content, created_at FROM comments WHERE post_id = ?", id)
	var comments []Comment
	for rows.Next() {
		var comment Comment
		rows.Scan(&comment.Id, &comment.Username, &comment.Content, &comment.CreatedAt)
		comments = append(comments, comment)
	}
	return comments
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

// GetPostsByCategories returns all posts for all categories
func GetPostsByCategories(database *sql.DB) [][]Post {
	categories := GetCategories(database)
	var posts [][]Post
	for _, category := range categories {
		posts = append(posts, GetPostsByCategory(database, category))
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

// GetCategoriesIcons returns all categories' icons
func GetCategoriesIcons(database *sql.DB) []string {
	rows, _ := database.Query("SELECT icon FROM categories")
	var icons []string
	for rows.Next() {
		var icon string
		rows.Scan(&icon)
		icons = append(icons, icon)
	}
	return icons
}

// GetCategoryIcon returns the icon for a category
func GetCategoryIcon(database *sql.DB, category string) string {
	rows, _ := database.Query("SELECT icon FROM categories WHERE name = ?", category)
	var icon string
	for rows.Next() {
		rows.Scan(&icon)
	}
	return icon
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
