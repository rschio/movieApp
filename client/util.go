package client

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"strconv"
)

func decodeResponse(dst interface{}, r io.Reader) error {
	err := json.NewDecoder(r).Decode(dst)
	if err != nil {
		return fmt.Errorf("failed to decode response: %v", err)
	}
	return err
}

func marshalItems(items []int) ([]byte, error) {
	body := new(toChangeList)
	body.Items = make([]listItem, len(items))
	for i, item := range items {
		body.Items[i] = listItem{MediaType: "movie", MediaID: item}
	}
	return json.Marshal(body)
}

// intsToString convert []int to a string comma separeted.
func intsToString(slice []int) string {
	buf := new(bytes.Buffer)
	for _, val := range slice {
		buf.WriteString(strconv.Itoa(val))
		buf.WriteByte(',')
	}
	s := buf.String()
	// remove last comma.
	if len(slice) > 0 {
		s = s[:len(s)-1]
	}
	return s
}
