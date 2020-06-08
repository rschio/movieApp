package client

import (
	"io"
	"net/http"
	"net/url"
)

// DefaultURL is the URL to api v4 of TMDB.
const DefaultURL = "https://api.themoviedb.org/4"

// Client is a TMDB API client.
type Client struct {
	client   *http.Client
	apiToken string
	baseURL  string
}

// New creates a new client of TMDB API.
func New(baseURL, apiToken string, client *http.Client) *Client {
	if client == nil {
		client = http.DefaultClient
	}
	c := &Client{
		client:   client,
		apiToken: apiToken,
		baseURL:  baseURL,
	}
	return c
}

// MakeGet make a GET request with specified params to path path.
func (c *Client) MakeGet(path string, params url.Values) (*http.Response, error) {
	encoded := params.Encode()
	r, err := http.NewRequest("GET", c.baseURL+path, nil)
	if err != nil {
		return nil, err
	}
	r.URL.RawQuery = encoded
	r.Header.Set("Authorization", "Bearer "+c.apiToken)
	return c.client.Do(r)
}

// MakePost make a POST request with specified body body to path path.
func (c *Client) MakePost(path string, body io.Reader) (*http.Response, error) {
	r, err := http.NewRequest("POST", c.baseURL+path, body)
	if err != nil {
		return nil, err
	}
	r.Header.Set("Content-Type", "application/json")
	r.Header.Set("Authorization", "Bearer "+c.apiToken)
	return c.client.Do(r)
}

// MakeDelete make a DELETe request with specified body body to path path.
func (c *Client) MakeDelete(path string, body io.Reader) (*http.Response, error) {
	r, err := http.NewRequest("DELETE", c.baseURL+path, body)
	if err != nil {
		return nil, err
	}
	r.Header.Set("Content-Type", "application/json")
	r.Header.Set("Authorization", "Bearer "+c.apiToken)
	return c.client.Do(r)
}
