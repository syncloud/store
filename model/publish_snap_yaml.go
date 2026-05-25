package model

type PublishSnapYamlRequest struct {
	Token    string `json:"token"`
	Name     string `json:"name"`
	Channel  string `json:"channel"`
	SnapYaml string `json:"snap_yaml"`
}

type PublishSnapYamlResponse struct {
	Ok bool `json:"ok"`
}
