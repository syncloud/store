package model

type StoreInfoChannelSnap struct {
	Snap
	Channel StoreInfoChannel `json:"channel"`
}
