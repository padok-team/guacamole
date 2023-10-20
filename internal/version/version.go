package version

import (
	"fmt"
)

var (
	Version        string
	CommitHash     string
	BuildTimestamp string
)

func BuildVersion() string {
	return fmt.Sprintf("%s-%s (%s)", Version, CommitHash, BuildTimestamp)
}
