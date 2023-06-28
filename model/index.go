package model

import "encoding/json"

type Index struct {
	Apps []json.RawMessage `json:"apps"`
}
