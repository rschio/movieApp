package client

import (
	"os"
	"testing"
)

var (
	testID      int
	godfatherID = 238
)

func newClient() *Client {
	token := os.Getenv("TMDBTOKEN")
	return New(DefaultURL, token, nil)
}

func TestSearchMovie(t *testing.T) {
	c := newClient()
	query := "GodFather"
	res, err := c.SearchMovie(query)
	if err != nil {
		t.Errorf("failed to search movie: %v", err)
		return
	}
	if len(res) <= 0 {
		t.Errorf("no results for GodFather")
	}
}

func TestDiscoverMovie(t *testing.T) {
	c := newClient()
	genres := []int{18, 80}
	res, err := c.DiscoverMovie(genres)
	if err != nil {
		t.Errorf("failed to discover movie: %v", err)
		return
	}
	if len(res) <= 0 {
		t.Errorf("no results")
	}
}

func TestCreateList(t *testing.T) {
	c := newClient()
	listname := "testlist"
	var err error
	testID, err = c.CreateList(listname)
	if err != nil {
		t.Errorf("failed to create list: %v", err)
		return
	}
	if testID < 0 {
		t.Errorf("negative id")
	}
}

func TestGetList(t *testing.T) {
	c := newClient()
	page := 1
	listResp, err := c.GetList(testID, page)
	if err != nil {
		t.Errorf("failed to get list: %v", err)
		return
	}
	if listResp.ID != testID {
		t.Error("got a list with different id")
		return
	}

}

func TestAddItem(t *testing.T) {
	c := newClient()
	resp, err := c.AddItems(testID, godfatherID)
	if err != nil {
		t.Errorf("failed to add movie: %v", err)
		return
	}
	movie := resp.Results[0]
	if movie.MediaID != godfatherID || movie.Success == false {
		t.Errorf("failed to add godfather movie")
	}
}

func TestDeleteItem(t *testing.T) {
	c := newClient()
	resp, err := c.DeleteItems(testID, godfatherID)
	if err != nil {
		t.Errorf("failed to delete movie: %v", err)
		return
	}
	movie := resp.Results[0]
	if movie.MediaID != godfatherID || movie.Success == false {
		t.Errorf("failed to remove godfather movie")
	}
}
