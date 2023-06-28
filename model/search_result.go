package model

type SearchResult struct {
	Revision SearchRevision `json:"revision"`
	Snap     Snap           `json:"snap"`
	Name     string         `json:"name"`
	SnapID   string         `json:"snap-id"`
}
