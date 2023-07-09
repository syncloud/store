package model

import (
	"fmt"
	"strconv"
)

type App struct {
	Name     string `json:"id"`
	Summary  string `json:"name"`
	Icon     string `json:"icon,omitempty"`
	Enabled  bool   `json:"enabled,omitempty"`
	Required bool   `json:"required"`
}

func (a *App) ToInfo(version string, downloadSize int64, downloadSha384 string, downloadUrl string, arch string) (*Snap, error) {
	appType := "app"
	if a.Required {
		appType = "base"
	}

	revision, err := strconv.Atoi(version)
	if err != nil {
		return nil, fmt.Errorf("unable to get revision: %s", err)
	}
	snapId := NewSnapId(a.Name, version)

	result := &Snap{
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
			{
				Type: "icon",
				URL:  a.Icon,
			},
		},
	}

	return result, nil
}
