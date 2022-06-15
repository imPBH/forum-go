package webAPI

import (
	"FORUM-GO/databaseAPI"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"
)

type Vote struct {
	PostId int
	Vote   int
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
	http.Redirect(w, r, "/filter?by=myposts", http.StatusFound)
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
	if !isLoggedIn(r) {
		http.Redirect(w, r, "/login", http.StatusFound)
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
	http.Redirect(w, r, "/post?id="+postId, http.StatusFound)
}

// VoteApi api to vote on a post
func VoteApi(w http.ResponseWriter, r *http.Request) {
	if r.Method == "POST" {
		if !isLoggedIn(r) {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}
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
