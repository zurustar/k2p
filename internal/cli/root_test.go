package cli

import (
	"bytes"
	"strings"
	"testing"
)

func TestRootCommand(t *testing.T) {
	// Test help output
	cmd := rootCmd
	cmd.SetArgs([]string{"--help"})
	
	buf := new(bytes.Buffer)
	cmd.SetOut(buf)
	cmd.SetErr(buf)
	
	err := cmd.Execute()
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
	
	output := buf.String()
	if !strings.Contains(output, "kindle-to-pdf") {
		t.Errorf("Help output should contain command name, got: %s", output)
	}
}

func TestVersionCommand(t *testing.T) {
	cmd := rootCmd
	cmd.SetArgs([]string{"--version"})
	
	buf := new(bytes.Buffer)
	cmd.SetOut(buf)
	cmd.SetErr(buf)
	
	err := cmd.Execute()
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
	
	output := buf.String()
	if !strings.Contains(output, "version") {
		t.Errorf("Version output should contain version information, got: %s", output)
	}
}