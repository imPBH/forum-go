package webAPI

import (
	"FORUM-GO/databaseAPI"
	"database/sql"
	_ "github.com/mattn/go-sqlite3"
	"html/template"
	"net/http"
)

var database *sql.DB

func SetDatabase(db *sql.DB) {
	database = db
}

// Index displays the Index page
func Index(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		http.NotFound(w, r)
		return
	}
	if isLoggedIn(r) {
		cookie, _ := r.Cookie("SESSION")
		t, _ := template.ParseGlob("templates/*.html")
		t.ExecuteTemplate(w, "createpost.html", databaseAPI.GetUser(database, cookie.Value))
		return
	}
	t, _ := template.ParseGlob("templates/*.html")
	t.ExecuteTemplate(w, "index.html", "")
	return
}

// DisplayPost displays a post on a template
func DisplayPost(w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	id := r.URL.Query().Get("id")
	post := databaseAPI.GetPost(database, id)
	post.Comments = databaseAPI.GetComments(database, id)
	t, _ := template.ParseGlob("templates/*.html")
	t.ExecuteTemplate(w, "postTemplate.html", post)
}

// GetPostsApi display all posts on a template
func GetPostsApi(w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	posts := databaseAPI.GetPosts(database)
	t, _ := template.ParseGlob("templates/*.html")
	t.ExecuteTemplate(w, "posts.html", posts)
}

// GetPostsByApi GetPostByApi gets all post filtered by the given parameters
func GetPostsByApi(w http.ResponseWriter, r *http.Request) {
	method := r.URL.Query().Get("by")
	if method == "category" {
		category := r.URL.Query().Get("category")
		posts := databaseAPI.GetPostsByCategory(database, category)
		t, _ := template.ParseGlob("templates/*.html")
		t.ExecuteTemplate(w, "posts.html", posts)
		return
	}
	if method == "myposts" {
		if isLoggedIn(r) {
			cookie, _ := r.Cookie("SESSION")
			username := databaseAPI.GetUser(database, cookie.Value)
			posts := databaseAPI.GetPostsByUser(database, username)
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
			username := databaseAPI.GetUser(database, cookie.Value)
			posts := databaseAPI.GetLikedPosts(database, username)
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

// inArray check if a string is in an array
func inArray(input string, array []string) bool {
	for _, v := range array {
		if v == input {
			return true
		}
	}
	return false
}
