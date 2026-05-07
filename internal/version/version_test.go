package version

import "testing"

func TestBuildInfoStringIncludesVersionCommitAndDate(t *testing.T) {
	info := BuildInfo{
		Version: "v1.2.3",
		Commit:  "abcdef1",
		Date:    "2026-05-07T06:30:00Z",
	}

	got := info.String()
	want := "ReaderX v1.2.3\nCommit: abcdef1\nBuilt: 2026-05-07T06:30:00Z"
	if got != want {
		t.Fatalf("String() = %q, want %q", got, want)
	}
}

func TestCurrentUsesPackageVariables(t *testing.T) {
	oldVersion, oldCommit, oldDate := Version, Commit, Date
	t.Cleanup(func() {
		Version, Commit, Date = oldVersion, oldCommit, oldDate
	})
	Version = "v9.9.9"
	Commit = "1234567"
	Date = "2026-05-07T06:31:00Z"

	info := Current()
	if info.Version != Version || info.Commit != Commit || info.Date != Date {
		t.Fatalf("Current() = %#v, want package variables", info)
	}
}
