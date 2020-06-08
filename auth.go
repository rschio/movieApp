package main

import (
	"context"
	"io/ioutil"
	"net/http"
	"strings"

	"firebase.google.com/go/auth"
	"github.com/rschio/movieApp/account"
)

// authenticate verify if user if user is logged with a valid session.
func (s *server) authenticate(r *http.Request) (*auth.Token, error) {
	ctx := r.Context()
	session, err := r.Cookie("session")
	if err != nil {
		return nil, err
	}
	return s.auther.VerifySessionCookie(ctx, session.Value)
}

// createAccount creates a account on firebase with account info, store profiles
// as firebase claims and send a email to a.Email with verification link.
func (s *server) createAccount(ctx context.Context, a *account.Account) error {
	u := new(auth.UserToCreate)
	u.Email(a.Email)
	u.Password(a.Password)
	u.DisplayName(a.Name)
	u.EmailVerified(false)
	user, err := s.auther.CreateUser(ctx, u)
	if err != nil {
		return err
	}
	// Set birthday and profiles as claims of user token.
	// This avoids to create a storage only for that and
	// avoid a bunch of requests to firebase API.
	update := new(auth.UserToUpdate)
	claims := map[string]interface{}{
		"birthday": a.Birthday,
		"profiles": a.Profiles,
	}
	update.CustomClaims(claims)
	_, err = s.auther.UpdateUser(ctx, user.UID, update)
	if err != nil {
		return err
	}
	link, err := s.auther.EmailVerificationLink(ctx, a.Email)
	if err != nil {
		return err
	}
	// Send verification link to confirm that emails is owned
	// by user.
	return s.mailer.SendVerificationLink(a.Name, a.Email, link)
}

func getIDTokenFromBody(r *http.Request) (string, error) {
	b, err := ioutil.ReadAll(r.Body)
	if err != nil {
		return "", err
	}
	data := string(b)
	data = strings.TrimPrefix(data, "idToken=")
	data = strings.TrimSuffix(data, "&csrfToken=")
	return data, nil
}
