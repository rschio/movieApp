package account

import (
	"fmt"
	"net/http"
	"strconv"
	"time"

	"firebase.google.com/go/auth"
	"github.com/rschio/movieApp/client"
)

// Account stores user informartion and user profiles.
// Account has at most 4 Profiles.
type Account struct {
	Email    string
	Name     string
	Password string
	Birthday time.Time
	Profiles []Profile
}

// FromUserToken gets the user information from a firebase token
// and returns a Account type with Email, Name, Birthday and Profiles
// or error.
func FromUserToken(token *auth.Token) (*Account, error) {
	acc := new(Account)
	// get email and name.
	var ok bool
	if acc.Email, ok = token.Claims["email"].(string); !ok {
		return nil, fmt.Errorf("failed to get email")
	}
	if acc.Name, ok = token.Claims["name"].(string); !ok {
		return nil, fmt.Errorf("failed to get name")
	}
	// get birthday.
	birthday := token.Claims["birthday"]
	datestr, ok := birthday.(string)
	if !ok {
		return nil, fmt.Errorf("failed to get birthday")
	}
	// parse birthday date from string to time.Time.
	var err error
	acc.Birthday, err = time.Parse(time.RFC3339, datestr)
	if err != nil {
		return nil, err
	}
	// Here we get profiles and we have to make some asserts.
	// First assert a interface{} -> []interface{} then
	// range this slice and get a interface{} and assert
	// interface{} -> map[string]interface{} then
	// get the values from map asserting to float64 and
	// cast to int or just assert to string.
	profiles := token.Claims["profiles"]
	slice, ok := profiles.([]interface{})
	if !ok {
		return nil, fmt.Errorf("failed to get profiles")
	}
	for _, item := range slice {
		p := item.(map[string]interface{})
		if !ok {
			return nil, fmt.Errorf("failed to get profiles")
		}
		prof := Profile{
			Name:             p["Name"].(string),
			WatchListID:      int(p["WatchListID"].(float64)),
			WatchedListID:    int(p["WatchedListID"].(float64)),
			SujestionsListID: int(p["SujestionsListID"].(float64)),
		}
		acc.Profiles = append(acc.Profiles, prof)
	}
	return acc, nil
}

// ProfileFromRequest returns the profile id from a request or error.
func ProfileFromRequest(r *http.Request, acc *Account) (int, error) {
	profile, err := r.Cookie("profile")
	if err != nil {
		return -1, err
	}
	p, err := strconv.Atoi(profile.Value)
	if err != nil {
		return -1, err
	}
	if p >= len(acc.Profiles) {
		return -1, fmt.Errorf("invalid profile")
	}
	return p, nil
}

// Profile stores the name the profile and 3 list IDs.
type Profile struct {
	// Name is the name of profile.
	Name string
	// WatchListID is the ID of list of movies user want to watch.
	WatchListID int
	// WatchedListID it the ID of list of movies user has already watched.
	WatchedListID int
	// SujestionsListID is the ID of list of movies user may want to watch.
	SujestionsListID int
}

// New creates a new account with email, password, name, birthday and one profile.
// The profile created by New has the name of account and is already populated with
// listIDs.
func New(email, password, name string, birthday time.Time, c *client.Client) *Account {
	acc := &Account{
		Email:    email,
		Name:     name,
		Password: password,
		Birthday: birthday,
		Profiles: make([]Profile, 0, 4),
	}
	acc.NewProfile(name, c)
	return acc
}

// NewProfile creates a new profile to a Account with name name
// if account has less than 4 profiles.
// The profile created is populated with listIDs.
func (a *Account) NewProfile(name string, c *client.Client) error {
	if len(a.Profiles) >= 4 {
		return fmt.Errorf("limit of profiles reached")
	}
	p := Profile{
		Name: name,
	}
	// Set the name of list as "emailnameListType".
	baseName := a.Email + p.Name
	ids, err := createListIDs(baseName, c)
	if err != nil {
		return err
	}
	p.WatchListID = ids[0]
	p.WatchedListID = ids[1]
	p.SujestionsListID = ids[2]

	a.Profiles = append(a.Profiles, p)
	return nil
}

// createListIDs create the 3 profile's list concurrently.
func createListIDs(baseName string, c *client.Client) ([]int, error) {
	lists := []string{"WatchList", "WatchedList", "SujestionsList"}
	ids := make([]int, 3)
	errs := make(chan error, 1)
	// Range lists and try to create each one.
	for i, list := range lists {
		// Make requests concurrent to not wait each request.
		go func(i int, list string) {
			var err error
			// create list and send error to  errs channel.
			ids[i], err = c.CreateList(baseName + list)
			errs <- err
		}(i, list)
	}
	// Check errors.
	var outputErr error
	for i := 0; i < len(lists); i++ {
		// If some o errs is not nil store it into
		// outputErr, only the last err is returned.
		// Drain the channel, do not stop in first err.
		err := <-errs
		if err != nil {
			outputErr = err
		}
	}
	if outputErr != nil {
		return nil, outputErr
	}
	return ids, nil
}
