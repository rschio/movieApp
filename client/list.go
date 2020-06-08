package client

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
)

type listToCreate struct {
	Name   string `json:"name"`
	ISO    string `json:"iso_639_1"`
	Public bool   `json:"public"`
}

type createListResponse struct {
	StatusMessage string `json:"status_message"`
	ID            int    `json:"id"`
	Success       bool   `json:"success"`
	StatusCode    int    `json:"status_code"`
}

// CreateList create a list in TMDB with name name
// and returns the list ID or error.
func (c *Client) CreateList(name string) (int, error) {
	const path = "/list"
	l := listToCreate{Name: name, ISO: "en"}
	payload, err := json.Marshal(l)
	if err != nil {
		return 0, err
	}
	resp, err := c.MakePost(path, bytes.NewReader(payload))
	if err != nil {
		return 0, err
	}
	defer resp.Body.Close()

	clResp := new(createListResponse)
	err = decodeResponse(clResp, resp.Body)
	if err != nil {
		return 0, err
	}
	if clResp.Success == false {
		return 0, fmt.Errorf("failed to create list")
	}

	return clResp.ID, nil
}

// List stores the content of a TMDB list.
type List struct {
	ID            int      `json:"id"`
	PosterPath    string   `json:"poster_path"`
	BackdropPath  string   `json:"backdrop_path"`
	TotalResults  int      `json:"total_results"`
	Public        bool     `json:"public"`
	Page          int      `json:"page"`
	Results       []Result `json:"results"`
	TotalPages    int      `json:"total_pages"`
	Description   string   `json:"description"`
	AverageRating float64  `json:"average_rating"`
	Name          string   `json:"name"`
}

// GetList get a list by id and page and returns the
// list *List or error.
func (c *Client) GetList(id, page int) (*List, error) {
	path := "/list/" + strconv.Itoa(id)
	if page < 1 {
		page = 1
	}
	params := make(url.Values)
	params.Set("page", strconv.Itoa(page))
	resp, err := c.MakeGet(path, params)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	lResp := new(List)
	err = decodeResponse(lResp, resp.Body)
	if err != nil {
		return nil, err
	}
	return lResp, nil
}

// Get lists get lists concurrently, based on id and page of list.
func (c *Client) GetLists(idPage ...int) ([]*List, error) {
	if len(idPage)%2 != 0 {
		return nil, fmt.Errorf("invalid length of idPage")
	}
	n := len(idPage) / 2
	lists := make([]*List, n)
	errs := make(chan error, 1)

	j := 0
	for i := 0; i < len(idPage); i += 2 {
		// Get the id and page of list.
		id, page := idPage[i], idPage[i+1]
		// Fetch list concurrently.
		go func(j, id, page int) {
			var err error
			lists[j], err = c.GetList(id, page)
			errs <- err
		}(j, id, page)
		j++
	}
	// Drain the errs channel and check error.
	// If an error occurs save it in err and
	// continue until the channel drained.
	var err error
	for i := 0; i < n; i++ {
		if e := <-errs; e != nil {
			err = e
		}
	}
	if err != nil {
		return nil, err
	}

	return lists, nil
}

type toChangeList struct {
	Items []listItem `json:"items"`
}

type listItem struct {
	MediaType string `json:"media_type"`
	MediaID   int    `json:"media_id"`
}

type changeListResponse struct {
	StatusMessage string `json:"status_message"`
	Results       []struct {
		MediaType string `json:"media_type"`
		MediaID   int    `json:"media_id"`
		Success   bool   `json:"success"`
	} `json:"results"`
	Success    bool `json:"success"`
	StatusCode int  `json:"status_code"`
}

type reqWithBodyFunc func(path string, body io.Reader) (*http.Response, error)

// list change change list (add or remove items).
func listChange(listID int, fn reqWithBodyFunc, items []int) (*changeListResponse, error) {
	path := "/list/" + strconv.Itoa(listID) + "/items"
	payload, err := marshalItems(items)
	if err != nil {
		return nil, err
	}
	resp, err := fn(path, bytes.NewReader(payload))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	clResp := new(changeListResponse)
	err = decodeResponse(clResp, resp.Body)
	if err != nil {
		return nil, err
	}
	// TODO: Check status message.
	return clResp, nil
}

// AddItems add items to list with ID listID.
func (c *Client) AddItems(listID int, items ...int) (*changeListResponse, error) {
	if len(items) == 0 {
		return nil, fmt.Errorf("need at least 1 item to add")
	}
	return listChange(listID, c.MakePost, items)
}

// DeleteItems delete items from list with ID listID.
func (c *Client) DeleteItems(listID int, items ...int) (*changeListResponse, error) {
	if len(items) == 0 {
		return nil, fmt.Errorf("need at least 1 item to remove")
	}
	return listChange(listID, c.MakeDelete, items)
}
