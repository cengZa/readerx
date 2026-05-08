package cmd

import "testing"

func TestRootHasSourceCommands(t *testing.T) {
	for _, args := range [][]string{
		{"source"},
		{"source", "list"},
		{"source", "search"},
	} {
		command, _, err := rootCmd.Find(args)
		if err != nil {
			t.Fatalf("Find %v: %v", args, err)
		}
		if command == rootCmd {
			t.Fatalf("Find %v returned root command", args)
		}
	}
}
