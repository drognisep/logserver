package protocol

import "fmt"

type VersionSpecifier struct {
	MajorVersion int `json:"major"`
	MinorVersion int `json:"minor"`
}

func (v VersionSpecifier) String() string {
	return fmt.Sprintf("%d.%d", v.MajorVersion, v.MinorVersion)
}

func ValidateVersion(version VersionSpecifier) error {
	if version.MajorVersion != 1 || version.MinorVersion != 0 {
		return fmt.Errorf("unrecognized protocol version %s", version)
	}
	return nil
}
