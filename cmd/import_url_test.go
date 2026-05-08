package cmd

import "testing"

func TestRootHasImportURLCommand(t *testing.T) {
	command, _, err := rootCmd.Find([]string{"import-url"})
	if err != nil {
		t.Fatalf("Find import-url: %v", err)
	}
	if command == rootCmd || command.Use != "import-url <url>" {
		t.Fatalf("command = %q, want import-url <url>", command.Use)
	}
}
