package version

import "fmt"

var (
	Version = "dev"
	Commit  = "unknown"
	Date    = "unknown"
)

type BuildInfo struct {
	Version string
	Commit  string
	Date    string
}

func Current() BuildInfo {
	return BuildInfo{
		Version: Version,
		Commit:  Commit,
		Date:    Date,
	}
}

func (i BuildInfo) String() string {
	return fmt.Sprintf("ReaderX %s\nCommit: %s\nBuilt: %s", i.Version, i.Commit, i.Date)
}
