package main

import (
	"encoding/json"
	"io"
)

func countJSONList(r io.Reader) (int, error) {
	var arr []json.RawMessage
	if err := json.NewDecoder(r).Decode(&arr); err != nil {
		return 0, err
	}
	return len(arr), nil
}

type findResponse struct {
	Results []json.RawMessage `json:"results"`
}

func countResults(r io.Reader) (int, error) {
	var resp findResponse
	if err := json.NewDecoder(r).Decode(&resp); err != nil {
		return 0, err
	}
	return len(resp.Results), nil
}
