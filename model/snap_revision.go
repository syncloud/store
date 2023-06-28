package model

type SnapRevision struct {
	Revision string `json:"snap-revision"`
	Id       string `json:"snap-id"`
	Size     string `json:"snap-size"`
	Sha384   string `json:"snap-sha3-385"`
}
