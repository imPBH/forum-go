package webAPI

import (
	"FORUM-GO/databaseAPI"
	"database/sql"
	_ "github.com/mattn/go-sqlite3"
	"html/template"
	"net/http"
)

type User struct {
	IsLoggedIn bool
	Username   string
}

type HomePage struct {
	User              User
	Categories        []string
	Icons             []string
	PostsByCategories [][]databaseAPI.Post
}

type PostsPage struct {
	User  User
	Title string
	Posts []databaseAPI.Post
	Icon  string
}

type PostPage struct {
	User User
	Post databaseAPI.Post
}

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
		payload := HomePage{
			User:              User{IsLoggedIn: true, Username: databaseAPI.GetUser(database, cookie.Value)},
			Categories:        databaseAPI.GetCategories(database),
			Icons:             databaseAPI.GetCategoriesIcons(database),
			PostsByCategories: databaseAPI.GetPostsByCategories(database),
		}
		t, _ := template.ParseGlob("public/HTML/*.html")
		t.ExecuteTemplate(w, "forum.html", payload)
		return
	}
	payload := HomePage{
		User:              User{IsLoggedIn: false},
		Categories:        databaseAPI.GetCategories(database),
		Icons:             databaseAPI.GetCategoriesIcons(database),
		PostsByCategories: databaseAPI.GetPostsByCategories(database),
	}
	t, _ := template.ParseGlob("public/HTML/*.html")
	t.ExecuteTemplate(w, "forum.html", payload)
	return
}

// DisplayPost displays a post on a template
func DisplayPost(w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	id := r.URL.Query().Get("id")
	payload := PostPage{
		Post: databaseAPI.GetPost(database, id),
	}
	if !isLoggedIn(r) {
		payload.User = User{IsLoggedIn: false}
	} else {
		payload.User = User{IsLoggedIn: true}
	}
	payload.Post.Comments = databaseAPI.GetComments(database, id)
	t, _ := template.ParseGlob("public/HTML/*.html")
	t.ExecuteTemplate(w, "detail.html", payload)
}

// GetPostsByApi GetPostByApi gets all post filtered by the given parameters
func GetPostsByApi(w http.ResponseWriter, r *http.Request) {
	method := r.URL.Query().Get("by")
	if method == "category" {
		category := r.URL.Query().Get("category")
		posts := databaseAPI.GetPostsByCategory(database, category)
		payload := PostsPage{
			Title: "Posts in category " + category,
			Posts: posts,
			Icon:  databaseAPI.GetCategoryIcon(database, category),
		}
		if isLoggedIn(r) {
			payload.User = User{IsLoggedIn: true}
		}
		t, _ := template.ParseGlob("public/HTML/*.html")
		t.ExecuteTemplate(w, "posts.html", payload)
		return
	}
	if method == "myposts" {
		if isLoggedIn(r) {
			cookie, _ := r.Cookie("SESSION")
			username := databaseAPI.GetUser(database, cookie.Value)
			posts := databaseAPI.GetPostsByUser(database, username)
			payload := PostsPage{
				User:  User{IsLoggedIn: true},
				Title: "My posts",
				Posts: posts,
				Icon:  "fa-user",
			}
			t, _ := template.ParseGlob("public/HTML/*.html")
			t.ExecuteTemplate(w, "posts.html", payload)
			return
		}
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}
	if method == "liked" {
		if isLoggedIn(r) {
			cookie, _ := r.Cookie("SESSION")
			username := databaseAPI.GetUser(database, cookie.Value)
			posts := databaseAPI.GetLikedPosts(database, username)
			payload := PostsPage{
				User:  User{IsLoggedIn: true},
				Title: "Posts liked by me",
				Posts: posts,
				Icon:  "fa-heart",
			}
			t, _ := template.ParseGlob("public/HTML/*.html")
			t.ExecuteTemplate(w, "posts.html", payload)
			return
		}
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}
	http.Redirect(w, r, "/", http.StatusSeeOther)
}

// NewPost displays the NewPost page
func NewPost(w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	if !isLoggedIn(r) {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}
	t, _ := template.ParseGlob("public/HTML/*.html")
	t.ExecuteTemplate(w, "createThread.html", nil)
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
