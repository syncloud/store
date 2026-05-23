package release

import (
	"fmt"
	"path/filepath"
	"regexp"
	"strings"
)

type Info struct {
	Version          string
	StoreSnapPath    string
	StoreVersionPath string
	StoreSizePath    string
	StoreSha384Path  string
	Name             string
}

func Parse(file string, branch string) (*Info, error) {
	name := filepath.Base(file)
	re := regexp.MustCompile(`/?(?P<Name>.*)_(?P<Version>.*)_(?P<Arch>.*).snap`)
	matches := re.FindStringSubmatch(name)
	if len(matches) < 3 {
		return nil, fmt.Errorf("cannot parse the file name, found only these parts: (%s)", strings.Join(matches, ","))
	}
	appName := matches[re.SubexpIndex("Name")]
	appArch := matches[re.SubexpIndex("Arch")]
	appVersion := matches[re.SubexpIndex("Version")]
	channel := branch
	if branch == "stable" {
		channel = "rc"
	}
	return &Info{
		StoreSnapPath:    fmt.Sprintf("apps/%s", name),
		StoreSha384Path:  fmt.Sprintf("apps/%s.sha384", name),
		StoreSizePath:    fmt.Sprintf("apps/%s.size", name),
		StoreVersionPath: fmt.Sprintf("releases/%s/%s.%s.version", channel, appName, appArch),
		Version:          appVersion,
		Name:             appName,
	}, nil
}
