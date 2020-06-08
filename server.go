package main

import (
	"context"
	"fmt"
	"html/template"
	"log"
	"math/rand"
	"path/filepath"
	"sync"

	firebase "firebase.google.com/go"
	"firebase.google.com/go/auth"
	"github.com/rschio/movieApp/client"
	"github.com/rschio/movieApp/mail"
	"google.golang.org/api/option"
)

type server struct {
	// tmpl is used to render web pages.
	tmpl *template.Template
	// auther is used to authenticate and
	// make login with firebase.
	auther *auth.Client
	// client make requests to TMDB API to
	// get movies and lists.
	client *client.Client
	// mailer send emails with sendgrid API.
	mailer *mail.Mailer
	// mu is a mutex for scheduleList.
	mu sync.Mutex
	// scheduleList schedules the movies to
	// send to email on determined time.
	scheduleList *ScheduleList
}

type serverConfig struct {
	templatePath    string
	autherCredsPath string
	clientAPIToken  string
	mailerAPIKey    string
	mailerName      string
	mailerAddr      string
}

func NewServer(cfg *serverConfig) *server {
	s := new(server)
	tmpls := filepath.Join(cfg.templatePath, "*")
	list := make(ScheduleList, 0, 10)
	s.tmpl = template.Must(template.ParseGlob(tmpls))
	s.auther = NewAuther(cfg.autherCredsPath)
	s.client = client.New(client.DefaultURL, cfg.clientAPIToken, nil)
	s.mailer = mail.NewMailer(cfg.mailerName, cfg.mailerAddr, cfg.mailerAPIKey)
	s.scheduleList = &list
	return s
}

func NewAuther(credsFile string) *auth.Client {
	opt := option.WithCredentialsFile(credsFile)
	app, err := firebase.NewApp(context.Background(), nil, opt)
	if err != nil {
		log.Fatalf("error initializing app: %v", err)
	}
	client, err := app.Auth(context.Background())
	if err != nil {
		log.Fatalf("error getting Auth client: %v\n", err)
	}
	return client
}

func (s *server) suggestMovie(listID int, genres []int) error {
	movies, err := s.client.DiscoverMovie(genres)
	if err != nil {
		log.Println(err)
		return err
	}
	n := len(movies)
	if n <= 0 {
		log.Println("failed to suggest movie")
		return fmt.Errorf("failed to suggest movie")
	}
	i := rand.Intn(n)
	_, err = s.client.AddItems(listID, movies[i].ID)
	return err
}
