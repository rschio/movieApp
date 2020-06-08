package main

import (
	"log"
	"net/http"
	"time"

	"github.com/rschio/movieApp/account"
)

type accountHandler func(http.ResponseWriter, *http.Request, *account.Account)

// Authorize authenticates the user and get account.
// Every protected handler should use Authorize as middleware.
func (s *server) Authorize(fn accountHandler) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		token, err := s.authenticate(r)
		if err != nil {
			http.Redirect(w, r, "/login", http.StatusFound)
			return
		}
		// Get account from firebase token.
		acc, err := account.FromUserToken(token)
		if err != nil {
			http.Redirect(w, r, "/login", http.StatusFound)
			return
		}
		// Execute the protected handler.
		fn(w, r, acc)
	}
}

func (s *server) signup(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	var (
		email    = r.FormValue("email-signup")
		password = r.FormValue("password-signup")
		name     = r.FormValue("name-signup")
		birthday = r.FormValue("birthday-signup")
	)
	// Parse date string to time.Time.
	date, err := time.Parse("2006-01-02", birthday)
	if err != nil {
		log.Printf("failed to parse date: %v", err)
		http.Error(w, "Invalid date", http.StatusBadRequest)
		return
	}
	// Create a new account, set it in firebase
	// and send verification email.
	acc := account.New(email, password, name, date, s.client)
	err = s.createAccount(r.Context(), acc)
	if err != nil {
		log.Printf("failed to create account: %v", err)
		http.Error(w, "account already exists or invalid parameters", http.StatusBadRequest)
		return
	}
	http.Redirect(w, r, "/login", http.StatusFound)
}

func (s *server) logout(w http.ResponseWriter, r *http.Request) {
	// Destroy session cookie.
	http.SetCookie(w, &http.Cookie{
		Name:     "session",
		Value:    "",
		MaxAge:   -1,
		HttpOnly: true,
		Secure:   true,
	})
	http.Redirect(w, r, "/login", http.StatusFound)
}

func (s *server) login(w http.ResponseWriter, r *http.Request) {
	if r.Method == "POST" {
		// Get firebase ID token.
		t, err := getIDTokenFromBody(r)
		if err != nil {
			log.Printf("failed to get ID token: %v", err)
			http.Error(w, "failed to get ID token", http.StatusUnauthorized)
			return
		}
		// Verify if token is valid.
		decoded, err := s.auther.VerifyIDToken(r.Context(), t)
		if err != nil {
			log.Println(err)
			http.Error(w, "invalid ID token", http.StatusUnauthorized)
			return
		}
		// Return error if the sign-in is older than 5 minutes.
		if time.Now().Unix()-decoded.AuthTime > 5*60 {
			log.Println("recent sign-in required")
			http.Error(w, "recent sign-in required", http.StatusUnauthorized)
			return
		}
		// Get a firebase session from IDToken.
		expiresIn := 6 * time.Hour
		cookie, err := s.auther.SessionCookie(r.Context(), t, expiresIn)
		if err != nil {
			log.Println(err)
			http.Error(w, "failed to create a session cookie", http.StatusInternalServerError)
			return
		}
		// Set cookie session with firebase sesssion.
		http.SetCookie(w, &http.Cookie{
			Name:     "session",
			Value:    cookie,
			MaxAge:   int(expiresIn.Seconds()),
			HttpOnly: true,
			Secure:   true,
		})
		// Redirect to choose profile.
		http.Redirect(w, r, "/", http.StatusFound)
		return
	}
	// Show login page with login fields.
	s.tmpl.ExecuteTemplate(w, "login.html", nil)
}
