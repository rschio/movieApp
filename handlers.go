package main

import (
	"container/heap"
	"log"
	"net/http"
	"strconv"
	"time"

	"firebase.google.com/go/auth"
	"github.com/rschio/movieApp/account"
)

// index serves a profile choose page.
func (s *server) index(w http.ResponseWriter, r *http.Request, acc *account.Account) {
	s.tmpl.ExecuteTemplate(w, "index.html", acc)
}

// browse is the core handler.
// It displays a field to seach movie by term.
// It displays all the profile's lists with pagination.
// The WatchList displays the actions to send movie to WatchedList or ScheduleMovie.
func (s *server) browse(w http.ResponseWriter, r *http.Request, acc *account.Account) {
	// Get the page param of each list.
	params := r.URL.Query()
	var (
		watchPage      = pageParam(params, "w")
		watchedPage    = pageParam(params, "d")
		sujestionsPage = pageParam(params, "s")
	)
	// Get the user profile.
	id, err := account.ProfileFromRequest(r, acc)
	if err != nil {
		log.Println(err)
		http.Error(w, "Bad Request", http.StatusBadRequest)
		return
	}
	profile := acc.Profiles[id]
	// Request all the 3 lists from TMDB API, concurrently.
	lists, err := s.client.GetLists(
		profile.WatchListID, watchPage,
		profile.WatchedListID, watchedPage,
		profile.SujestionsListID, sujestionsPage,
	)
	if err != nil {
		log.Println(err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
	// Paginate, check if list has a previous or
	// next page and if has set the correct page to
	// prev and next to template can render the actions
	// at end.
	toShow := []*ListPage{
		paginate(lists[0]),
		paginate(lists[1]),
		paginate(lists[2]),
	}
	genreID := preferredGenre(lists[0].Results, lists[1].Results)
	// Try to suggest one movie to next browse.
	go s.suggestMovie(profile.SujestionsListID, []int{genreID})
	// Execute the template with toShow data, this template
	// does a bunch of work.
	s.tmpl.ExecuteTemplate(w, "browse.html", toShow)
}

// searchMovie search a movie with specified query and display the results,
// the query must be sent as a POST request.
func (s *server) searchMovie(w http.ResponseWriter, r *http.Request, acc *account.Account) {
	if r.Method != "POST" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	query := r.FormValue("query")
	if query == "" {
		http.Error(w, "Invalid query", http.StatusBadRequest)
		return
	}
	movies, err := s.client.SearchMovie(query)
	if err != nil {
		log.Println(err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
	// Display the found movies.
	s.tmpl.ExecuteTemplate(w, "searchmovie.html", movies)
}

// addProfile creates a new profile with name profileName and updates the user
// account with new profile.
func (s *server) addProfile(w http.ResponseWriter, r *http.Request, acc *account.Account) {
	if r.Method != "POST" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	name := r.FormValue("profileName")
	if name == "" {
		http.Error(w, "Invalid name", http.StatusBadRequest)
		return
	}
	// Creates a new profile.
	err := acc.NewProfile(name, s.client)
	if err != nil {
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
	record, err := s.auther.GetUserByEmail(r.Context(), acc.Email)
	if err != nil {
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
	// Update the user with new profile to firebase.
	update := new(auth.UserToUpdate)
	update.CustomClaims(map[string]interface{}{
		"birthday": acc.Birthday,
		"profiles": acc.Profiles,
	})
	_, err = s.auther.UpdateUser(r.Context(), record.UID, update)
	if err != nil {
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
	// Revoke tokens (and logout) to get a updated token with new profile.
	if err = s.auther.RevokeRefreshTokens(r.Context(), record.UID); err != nil {
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
	http.Redirect(w, r, "/logout", http.StatusFound)
}

// addMovie add a movie to WatchList.
func (s *server) addMovie(w http.ResponseWriter, r *http.Request, acc *account.Account) {
	const path = "/addmovie/"
	movieID, err := idFromPath(path, r)
	if err != nil {
		log.Println(err)
		http.Error(w, "Bad Request", http.StatusBadRequest)
		return
	}
	id, err := account.ProfileFromRequest(r, acc)
	if err != nil {
		log.Println(err)
		http.Error(w, "Bad Request", http.StatusBadRequest)
		return
	}
	profile := acc.Profiles[id]
	// Add movieID to WatchList.
	_, err = s.client.AddItems(profile.WatchListID, movieID)
	if err != nil {
		log.Println(err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
	http.Redirect(w, r, "/browse", http.StatusFound)
}

// watchMovie deletes a movie from WatchList and add to WatchedList.
func (s *server) watchMovie(w http.ResponseWriter, r *http.Request, acc *account.Account) {
	const path = "/watchmovie/"
	movieID, err := idFromPath(path, r)
	if err != nil {
		log.Println(err)
		http.Error(w, "Bad Request", http.StatusBadRequest)
		return
	}
	id, err := account.ProfileFromRequest(r, acc)
	if err != nil {
		log.Println(err)
		http.Error(w, "Bad Request", http.StatusBadRequest)
		return
	}
	profile := acc.Profiles[id]
	_, err = s.client.DeleteItems(profile.WatchListID, movieID)
	if err != nil {
		log.Println(err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
	_, err = s.client.AddItems(profile.WatchedListID, movieID)
	if err != nil {
		log.Println(err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
	http.Redirect(w, r, "/browse", http.StatusFound)
}

// chooseProfile set a profile to profile cookie.
func (s *server) chooseProfile(w http.ResponseWriter, r *http.Request, acc *account.Account) {
	const path = "/profile/"
	p, err := idFromPath(path, r)
	if err != nil || p >= len(acc.Profiles) {
		http.Error(w, "Bad request", http.StatusBadRequest)
		return
	}
	profile := strconv.Itoa(p)
	// Set profile cookie with path "/".
	http.SetCookie(w, &http.Cookie{
		Name:     "profile",
		Value:    profile,
		Path:     "/",
		HttpOnly: true,
		Secure:   true,
	})
	http.Redirect(w, r, "/browse", http.StatusFound)
}

// showScheduler display the page to schedule a movie.
func (s *server) showScheduler(w http.ResponseWriter, r *http.Request, acc *account.Account) {
	const path = "/showscheduler/"
	movieID, err := idFromPath(path, r)
	if err != nil {
		log.Println(err)
		http.Error(w, "Bad Request", http.StatusBadRequest)
		return
	}
	s.tmpl.ExecuteTemplate(w, "schedulemovie.html", movieID)
}

// scheduleMovie add a movie to scheduleList.
func (s *server) scheduleMovie(w http.ResponseWriter, r *http.Request, acc *account.Account) {
	if r.Method != "POST" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	d := r.FormValue("date-schedule")
	t := r.FormValue("time-schedule")
	idStr := r.FormValue("movie-id")

	date, err := time.Parse("2006-01-02-15:04", d+"-"+t)
	if err != nil {
		log.Println(err)
		http.Error(w, "Bad Request", http.StatusBadRequest)
		return
	}
	// Check if id is a number.
	id, err := strconv.Atoi(idStr)
	if err != nil {
		log.Println(err)
		http.Error(w, "Bad Request", http.StatusBadRequest)
		return
	}
	register := &ScheduledMovie{
		Time:     date,
		MovieID:  id,
		UserName: acc.Name,
		Email:    acc.Email,
	}
	// Add register to scheduleList.
	s.mu.Lock()
	heap.Push(s.scheduleList, register)
	s.mu.Unlock()

	http.Redirect(w, r, "/browse", http.StatusFound)
}
