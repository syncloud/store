package model

import (
	"fmt"
	"strconv"
)

type App struct {
	Name        string `json:"id"`
	Summary     string `json:"name"`
	Description string `json:"description,omitempty"`
	Type        string `json:"type,omitempty"`
	Enabled     bool   `json:"enabled,omitempty"`
}

func (a *App) ToInfo(version string, downloadSize int64, downloadSha384 string, downloadUrl string, arch string, iconUrl string) (*Snap, error) {
	appType := a.Type
	if appType == "" {
		appType = "app"
	}
	revision, err := strconv.Atoi(version)
	if err != nil {
		return nil, fmt.Errorf("unable to get revision: %s", err)
	}
	snapId := NewSnapId(a.Name, version)
	return &Snap{
		SnapID:        snapId.Id(),
		Name:          a.Name,
		Summary:       a.Summary,
		Version:       version,
		Type:          appType,
		Architectures: []string{arch},
		Revision:      revision,
		Download: StoreSnapDownload{
			URL:      downloadUrl,
			Sha3_384: downloadSha384,
			Size:     downloadSize,
		},
		Media: []StoreSnapMedia{
			{Type: "icon", URL: iconUrl},
		},
	}, nil
}
