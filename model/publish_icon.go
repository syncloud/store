package model

type PublishIconRequest struct {
	Token      string `json:"token"`
	Name       string `json:"name"`
	Channel    string `json:"channel"`
	IconPngB64 string `json:"icon_png_b64"`
}

type PublishIconResponse struct {
	Ok bool `json:"ok"`
}
