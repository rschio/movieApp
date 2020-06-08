package client

import (
	"fmt"
	"net/url"
)

// Result is a movie search result.
type Result struct {
	PosterPath       string  `json:"poster_path"`
	Adult            bool    `json:"adult"`
	Overview         string  `json:"overview"`
	ReleaseDate      string  `json:"release_date"`
	GenreIds         []int   `json:"genre_ids"`
	ID               int     `json:"id"`
	OriginalTitle    string  `json:"original_title"`
	OriginalLanguage string  `json:"original_language"`
	Title            string  `json:"title"`
	BackdropPath     string  `json:"backdrop_path"`
	Popularity       float64 `json:"popularity"`
	VoteCount        int     `json:"vote_count"`
	Video            bool    `json:"video"`
	VoteAverage      float64 `json:"vote_average"`
}

type SearchMovieResp struct {
	Page         int      `json:"page"`
	Results      []Result `json:"results"`
	TotalResults int      `json:"total_results"`
	TotalPages   int      `json:"total_pages"`
}

// SearchMovie seachs a movie by a term an return the results.
func (c *Client) SearchMovie(query string) ([]Result, error) {
	const path = "/search/movie"
	params := make(url.Values)
	params.Set("query", query)
	resp, err := c.MakeGet(path, params)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	movieResp := new(SearchMovieResp)
	err = decodeResponse(movieResp, resp.Body)
	if err != nil {
		return nil, err
	}

	results := movieResp.Results
	if results == nil {
		return nil, fmt.Errorf("invalid results")
	}
	return results, nil
}

// DiscoverMovie searchs for movies based in genres.
func (c *Client) DiscoverMovie(genres []int) ([]Result, error) {
	const path = "/discover/movie"
	params := make(url.Values)
	params.Set("with_genres", intsToString(genres))
	resp, err := c.MakeGet(path, params)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	movieResp := new(SearchMovieResp)
	err = decodeResponse(movieResp, resp.Body)
	if err != nil {
		return nil, err
	}
	results := movieResp.Results
	if results == nil {
		return nil, fmt.Errorf("invalid results")
	}
	return results, nil
}
