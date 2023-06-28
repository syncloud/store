package model

type SearchResults struct {
	Results   []*SearchResult `json:"results"`
	ErrorList []SearchError   `json:"error-list"`
}
