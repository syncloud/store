package model

type StoreInfo struct {
	ChannelMap []*StoreInfoChannelSnap `json:"channel-map"`
	Snap       Snap                    `json:"snap"`
	Name       string                  `json:"name"`
	SnapID     string                  `json:"snap-id"`
}
