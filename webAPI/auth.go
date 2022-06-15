package webAPI

import (
	"FORUM-GO/databaseAPI"
	"fmt"
	uuid "github.com/satori/go.uuid"
	"golang.org/x/crypto/bcrypt"
	"html/template"
	"net/http"
	"time"
)

type Error struct {
	Message string
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

	if username == "" || email == "" || password == "" {
		http.Redirect(w, r, "/register?err=invalid_informations", http.StatusFound)
		return
	}
	if !databaseAPI.UsernameNotTaken(database, username) {
		http.Redirect(w, r, "/register?err=username_taken", http.StatusFound)
		return
	}
	if !databaseAPI.EmailNotTaken(database, email) {
		http.Redirect(w, r, "/register?err=email_taken", http.StatusFound)
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

// isExpired returns true if the cookie has expired
func isExpired(expires string) bool {
	expiresTime, _ := time.Parse("2006-01-02 15:04:05", expires)
	return time.Now().After(expiresTime)
}

// Register displays the Register page
func Register(w http.ResponseWriter, r *http.Request) {
	error := r.URL.Query().Get("err")
	payload := Error{Message: ""}
	if error == "invalid_informations" {
		payload = Error{Message: "Invalid informations"}
	}
	if error == "email_taken" {
		payload = Error{Message: "Email already taken"}
	}
	if error == "username_taken" {
		payload = Error{Message: "Username already taken"}
	}
	t, _ := template.ParseGlob("public/HTML/*.html")
	t.ExecuteTemplate(w, "registerForm.html", payload)
}

// Login displays template for the Login page
func Login(w http.ResponseWriter, r *http.Request) {
	error := r.URL.Query().Get("err")
	payload := Error{Message: ""}
	if error == "invalid_email" {
		payload = Error{Message: "Invalid email"}
	}
	if error == "invalid_password" {
		payload = Error{Message: "Invalid password"}
	}
	t, _ := template.ParseGlob("public/HTML/*.html")
	t.ExecuteTemplate(w, "signinForm.html", payload)
}
