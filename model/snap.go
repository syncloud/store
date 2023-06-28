package model

type Snap struct {
	SnapID        string            `json:"snap-id"`
	Name          string            `json:"name"`
	Summary       string            `json:"summary"`
	Version       string            `json:"version"`
	Type          string            `json:"type"`
	Architectures []string          `json:"architectures"`
	Revision      int               `json:"revision"` // store revisions are ints starting at 1
	Download      StoreSnapDownload `json:"download"`
	Media         []StoreSnapMedia  `json:"media"`
}
