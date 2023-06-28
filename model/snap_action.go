package model

type SnapAction struct {
	Action string `json:"action"`
	// For snap
	InstanceKey      string `json:"instance-key,omitempty"`
	Name             string `json:"name,omitempty"`
	SnapID           string `json:"snap-id,omitempty"`
	Channel          string `json:"channel,omitempty"`
	Revision         int    `json:"revision,omitempty"`
	CohortKey        string `json:"cohort-key,omitempty"`
	IgnoreValidation *bool  `json:"ignore-validation,omitempty"`
	// For assertions
	Key            string        `json:"key,omitempty"`
	Assertions     []interface{} `json:"assertions,omitempty"`
	ValidationSets [][]string    `json:"validation-sets,omitempty"`
}
