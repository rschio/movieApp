package main

import (
	"net/http"
	"net/url"
	"strconv"

	"github.com/rschio/movieApp/client"
)

// ListPage stores a list with
// Prev and Next pages.
type ListPage struct {
	List *client.List
	Prev int
	Next int
}

// Paginate check if list at it's page has
// a next page and prev page and set ListPage.
func paginate(list *client.List) *ListPage {
	lp := &ListPage{List: list}
	if list.Page > 1 {
		lp.Prev = list.Page - 1
	}
	if list.Page < list.TotalPages {
		lp.Next = list.Page + 1
	}
	return lp
}

// pageParam return the page params from url with name name.
func pageParam(params url.Values, name string) int {
	page, err := strconv.Atoi(params.Get(name))
	if err != nil {
		page = 1
	}
	return page
}

// idFromPath return the id from path.
func idFromPath(path string, r *http.Request) (int, error) {
	strID := r.URL.Path[len(path):]
	return strconv.Atoi(strID)
}

// preferredGenre returns the ID that is most frequent
// in the watch and watched lists.
func preferredGenre(watch, watched []client.Result) int {
	genres := make(map[int]int)
	for _, r := range watch {
		for _, id := range r.GenreIds {
			genres[id]++
		}
	}
	for _, r := range watched {
		for _, id := range r.GenreIds {
			genres[id]++
		}
	}
	mostWatchedKey := 0
	mostWatchedVal := 0
	for key, val := range genres {
		if val > mostWatchedVal {
			mostWatchedKey = key
		}
	}
	return mostWatchedKey
}
