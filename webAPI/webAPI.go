package webAPI

import (
	"FORUM-GO/databaseAPI"
	"database/sql"
	"fmt"
	_ "github.com/mattn/go-sqlite3"
	uuid "github.com/satori/go.uuid"
	"golang.org/x/crypto/bcrypt"
	"html/template"
	"net/http"
	"strconv"
	"strings"
	"time"
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

// RegisterApi handles the Register api
func RegisterApi(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		fmt.Fprintf(w, "ParseForm() err: %v", err)
		return
	}
	username := r.FormValue("username")
	email := r.FormValue("email")
	password := r.FormValue("password")
	value := uuid.NewV4().String()
	expiration := time.Now().Add(31 * 24 * time.Hour)
	if !databaseAPI.EmailNotTaken(database, email) {
		http.Redirect(w, r, "/register?err=email_taken", http.StatusFound)
		return
	}
	if !databaseAPI.UsernameNotTaken(database, username) {
		http.Redirect(w, r, "/register?err=username_taken", http.StatusFound)
		return
	}
	databaseAPI.AddUser(database, username, email, password, value, expiration.Format("2006-01-02 15:04:05"))
	cookie := http.Cookie{Name: "SESSION", Value: value, Expires: expiration, Path: "/"}
	http.SetCookie(w, &cookie)
	http.Redirect(w, r, "/", http.StatusFound)
	return
}

//LoginApi handles the Login api
func LoginApi(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		fmt.Fprintf(w, "ParseForm() err: %v", err)
		return
	}
	submittedEmail := r.FormValue("email")
	submittedPassword := r.FormValue("password")

	username, email, password := databaseAPI.GetUserInfo(database, submittedEmail)
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
	databaseAPI.UpdateCookie(database, value, expiration, email)
	fmt.Println("Logged in user: " + username + " with email: " + email + " at " + now)
	http.Redirect(w, r, "/", http.StatusFound)
	return
}

// Register displays the Register page
func Register(w http.ResponseWriter, r *http.Request) {
	t, _ := template.ParseGlob("templates/*.html")
	t.ExecuteTemplate(w, "register.html", "")
}

// Login displays template for the Login page
func Login(w http.ResponseWriter, r *http.Request) {
	t, _ := template.ParseGlob("templates/*.html")
	t.ExecuteTemplate(w, "login.html", "")
}

// isExpired returns true if the cookie has expired
func isExpired(expires string) bool {
	expiresTime, _ := time.Parse("2006-01-02 15:04:05", expires)
	return time.Now().After(expiresTime)
}

// CreatePostApi creates a post
func CreatePostApi(w http.ResponseWriter, r *http.Request) {
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
	username := databaseAPI.GetUser(database, cookie.Value)
	title := r.FormValue("title")
	content := r.FormValue("content")
	categories := r.Form["categories[]"]
	validCategories := databaseAPI.GetCategories(database)
	for _, category := range categories {
		// if string not in array, return error
		if !inArray(category, validCategories) {
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte("Invalid category : " + category))
			return
		}
	}
	stringCategories := strings.Join(categories, ",")
	now := time.Now()
	databaseAPI.CreatePost(database, username, title, stringCategories, content, now)
	fmt.Println("Post created by " + username + " with title " + title + " at " + now.Format("2006-01-02 15:04:05"))
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Post created"))
	return
}

// CommentsApi creates a comment
func CommentsApi(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	if err := r.ParseForm(); err != nil {
		fmt.Fprintf(w, "ParseForm() err: %v", err)
		return
	}
	cookie, _ := r.Cookie("SESSION")
	username := databaseAPI.GetUser(database, cookie.Value)
	postId := r.FormValue("postId")
	content := r.FormValue("content")
	now := time.Now()
	postIdInt, _ := strconv.Atoi(postId)
	databaseAPI.AddComment(database, username, postIdInt, content, now)
	fmt.Println("Comment created by " + username + " on post " + postId + " at " + now.Format("2006-01-02 15:04:05"))
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Comment created"))
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

// VoteApi api to vote on a post
func VoteApi(w http.ResponseWriter, r *http.Request) {
	if r.Method == "POST" {
		if err := r.ParseForm(); err != nil {
			fmt.Fprintf(w, "ParseForm() err: %v", err)
			return
		}
		cookie, _ := r.Cookie("SESSION")
		username := databaseAPI.GetUser(database, cookie.Value)
		postId := r.FormValue("postId")
		postIdInt, _ := strconv.Atoi(postId)
		vote := r.FormValue("vote")
		voteInt, _ := strconv.Atoi(vote)
		now := time.Now().Format("2006-01-02 15:04:05")
		if voteInt == 1 {
			if databaseAPI.HasUpvoted(database, username, postIdInt) {
				databaseAPI.RemoveVote(database, postIdInt, username)
				databaseAPI.DecreaseUpvotes(database, postIdInt)
				fmt.Println("Removed upvote from " + username + " on post " + postId + " at " + now)
				w.WriteHeader(http.StatusOK)
				w.Write([]byte("Vote removed"))
				return
			}
			if databaseAPI.HasDownvoted(database, username, postIdInt) {
				databaseAPI.DecreaseDownvotes(database, postIdInt)
				databaseAPI.IncreaseUpvotes(database, postIdInt)
				databaseAPI.UpdateVote(database, postIdInt, username, 1)
				fmt.Println(username + " upvoted" + " on post " + postId + " at " + now)
				w.WriteHeader(http.StatusOK)
				w.Write([]byte("Upvote added"))
				return
			}
			databaseAPI.IncreaseUpvotes(database, postIdInt)
			databaseAPI.AddVote(database, postIdInt, username, 1)
			fmt.Println(username + " upvoted" + " on post " + postId + " at " + now)
			w.WriteHeader(http.StatusOK)
			w.Write([]byte("Upvote added"))
			return
		}
		if voteInt == -1 {
			if databaseAPI.HasDownvoted(database, username, postIdInt) {
				databaseAPI.RemoveVote(database, postIdInt, username)
				databaseAPI.DecreaseDownvotes(database, postIdInt)
				fmt.Println("Removed downvote from " + username + " on post " + postId + " at " + now)
				w.WriteHeader(http.StatusOK)
				w.Write([]byte("Vote removed"))
				return
			}
			if databaseAPI.HasUpvoted(database, username, postIdInt) {
				databaseAPI.DecreaseUpvotes(database, postIdInt)
				databaseAPI.IncreaseDownvotes(database, postIdInt)
				databaseAPI.UpdateVote(database, postIdInt, username, -1)
				fmt.Println(username + " downvoted" + " on post " + postId + " at " + now)
				w.WriteHeader(http.StatusOK)
				w.Write([]byte("Downvote added"))
				return
			}
			databaseAPI.IncreaseDownvotes(database, postIdInt)
			databaseAPI.AddVote(database, postIdInt, username, -1)
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

// isLoggedIn checks if the user is logged in
func isLoggedIn(r *http.Request) bool {
	cookie, err := r.Cookie("SESSION")
	if err != nil {
		return false
	}
	cookieExists := databaseAPI.CheckCookie(database, cookie.Value)
	if !cookieExists {
		return false
	}
	expires := databaseAPI.GetExpires(database, cookie.Value)

	if isExpired(expires) {
		return false
	}
	return true
}

// LogoutAPI deletes the session cookie from the database
func LogoutAPI(w http.ResponseWriter, r *http.Request) {
	cookie, _ := r.Cookie("SESSION")
	username := databaseAPI.GetUser(database, cookie.Value)
	now := time.Now().Format("2006-01-02 15:04:05")
	if cookie != nil {
		username := databaseAPI.GetUser(database, cookie.Value)
		databaseAPI.Logout(database, username)
	}
	fmt.Println("User " + username + " logged out at " + now)
	http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
	return
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
