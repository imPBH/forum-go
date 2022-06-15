package databaseAPI

import (
	"database/sql"
	_ "github.com/mattn/go-sqlite3"
)

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
