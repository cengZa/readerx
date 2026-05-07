package cmd

import "testing"

func TestRemoveConfirmationMatchesBookID(t *testing.T) {
	if !removeConfirmationMatches(" 42\n", 42) {
		t.Fatalf("expected matching book id to confirm removal")
	}
	if removeConfirmationMatches("41", 42) {
		t.Fatalf("different book id should not confirm removal")
	}
	if removeConfirmationMatches("yes", 42) {
		t.Fatalf("non-numeric input should not confirm removal")
	}
}
