package main

import (
	"FORUM-GO/databaseAPI"
	"FORUM-GO/webAPI"
	"database/sql"
	"fmt"
	_ "github.com/mattn/go-sqlite3"
	"net/http"
	"os"
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

	databaseAPI.CreateUsersTable(database)
	databaseAPI.CreatePostTable(database)
	databaseAPI.CreateCommentTable(database)
	databaseAPI.CreateVoteTable(database)
	databaseAPI.CreateCategoriesTable(database)
	databaseAPI.CreateCategories(database)

	webAPI.SetDatabase(database)

	fs := http.FileServer(http.Dir("templates"))
	router := http.NewServeMux()
	fmt.Println("Starting server on port 8000")

	router.HandleFunc("/", webAPI.Index)
	router.HandleFunc("/register", webAPI.Register)
	router.HandleFunc("/login", webAPI.Login)
	router.HandleFunc("/api/register", webAPI.RegisterApi)
	router.HandleFunc("/api/login", webAPI.LoginApi)
	router.HandleFunc("/api/createpost", webAPI.CreatePostApi)
	router.HandleFunc("/api/comments", webAPI.CommentsApi)
	router.HandleFunc("/post", webAPI.DisplayPost)
	router.HandleFunc("/posts", webAPI.GetPostsApi)
	router.HandleFunc("/api/vote", webAPI.VoteApi)
	router.HandleFunc("/filter", webAPI.GetPostsByApi)
	router.HandleFunc("/api/logout", webAPI.LogoutAPI)

	router.Handle("/templates/", http.StripPrefix("/templates/", fs))
	http.ListenAndServe(":8000", router)
}
