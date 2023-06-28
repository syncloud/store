package model

type ErrorListEntry struct {
	Code    string `json:"code"`
	Message string `json:"message"`
	// for assertions
	Type string `json:"type"`
	// either primary-key or sequence-key is expected (but not both)
	PrimaryKey  []string `json:"primary-key,omitempty"`
	SequenceKey []string `json:"sequence-key,omitempty"`
}
