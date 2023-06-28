package model

import "strings"

type SnapId string

func (s SnapId) Id() string {
	return string(s)
}

func (s SnapId) IsEmpty() bool {
	return string(s) == ""
}

func (s SnapId) Name() string {
	if strings.Contains(s.Id(), ".") {
		parts := strings.Split(s.Id(), ".")
		return parts[0]
	} else {
		return s.Id()
	}
}

func (s SnapId) Version() string {
	if strings.Contains(s.Id(), ".") {
		parts := strings.Split(s.Id(), ".")
		return parts[1]
	} else {
		return ""
	}
}
