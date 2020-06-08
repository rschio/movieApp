package main

import (
	"context"
	"net/http"
	"os"
)

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	srvCfg := &serverConfig{
		templatePath:    "templates",
		autherCredsPath: os.Getenv("AUTHERCREDSPATH"),
		clientAPIToken:  os.Getenv("TMDBTOKEN"),
		mailerAPIKey:    os.Getenv("SENDGRID_API_KEY"),
		mailerName:      "no-reply",
		mailerAddr:      os.Getenv("MAILERADDR"),
	}
	s := NewServer(srvCfg)

	ctx := context.Background()
	go s.schedule(ctx)

	static := http.FileServer(http.Dir("static"))
	http.Handle("/scripts/", static)

	http.HandleFunc("/", s.Authorize(s.index))
	http.HandleFunc("/profile/", s.Authorize(s.chooseProfile))
	http.HandleFunc("/addprofile", s.Authorize(s.addProfile))
	http.HandleFunc("/browse", s.Authorize(s.browse))
	http.HandleFunc("/searchmovie", s.Authorize(s.searchMovie))
	http.HandleFunc("/addmovie/", s.Authorize(s.addMovie))
	http.HandleFunc("/watchmovie/", s.Authorize(s.watchMovie))
	http.HandleFunc("/showscheduler/", s.Authorize(s.showScheduler))
	http.HandleFunc("/schedulemovie", s.Authorize(s.scheduleMovie))
	http.HandleFunc("/login", s.login)
	http.HandleFunc("/logout", s.logout)
	http.HandleFunc("/signup", s.signup)
	http.ListenAndServe(":"+port, nil)
}
