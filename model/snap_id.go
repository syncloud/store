package model

import (
	"fmt"
	"strings"
)

type SnapId string

func NewSnapId(name string, version string) SnapId {
	return SnapId(fmt.Sprintf("%s.%s", name, version))
}

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

//func (s SnapId) Arch() string {
//	if strings.Contains(s.Id(), ".") {
//		parts := strings.Split(s.Id(), ".")
//		if len(parts) > 2 {
//			return parts[2]
//		}
//		return "amd64"
//	} else {
//		return ""
//	}
//}
