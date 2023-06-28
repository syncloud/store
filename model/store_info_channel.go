package model

import "time"

type StoreInfoChannel struct {
	Architecture string    `json:"architecture"`
	Name         string    `json:"name"`
	Risk         string    `json:"risk"`
	Track        string    `json:"track"`
	ReleasedAt   time.Time `json:"released-at"`
}
