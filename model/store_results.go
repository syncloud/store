package model

type StoreResults struct {
	Results   []*StoreResult   `json:"results"`
	ErrorList []ErrorListEntry `json:"error-list,omitempty"`
}
