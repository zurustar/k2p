package main

import (
	"bytes"
	"context"
	"strings"
	"testing"

	"github.com/leanovate/gopter"
	"github.com/leanovate/gopter/gen"
	"github.com/leanovate/gopter/prop"
	"github.com/oumi/k2p/internal/config"
	"github.com/oumi/k2p/internal/orchestrator"
)

// MockOrchestrator for testing
type MockOrchestrator struct {
	ShouldFail bool
}

func (m *MockOrchestrator) ConvertCurrentBook(ctx context.Context, options *config.ConversionOptions) (*orchestrator.ConversionResult, error) {
	if m.ShouldFail {
		return nil, context.DeadlineExceeded // some error
	}
	return &orchestrator.ConversionResult{
		OutputPath: "/tmp/output.pdf",
		PageCount:  10,
	}, nil
}

// Property 6: Usage Display
// Property 7: Help Flag Response
// Property 10: Version Display
// Property 9: Invalid Argument Handling
func TestCLIProperties(t *testing.T) {
	parameters := gopter.DefaultTestParameters()
	parameters.MinSuccessfulTests = 100
	properties := gopter.NewProperties(parameters)

	// Property: Help flag always succeeds and prints usage
	properties.Property("Help flag prints usage", prop.ForAll(
		func() bool {
			stdout := &bytes.Buffer{}
			stderr := &bytes.Buffer{}
			orch := &MockOrchestrator{}

			code := run([]string{"--help"}, stdout, stderr, orch)

			return code == 0 && strings.Contains(stdout.String(), "USAGE")
		},
	))

	// Property: Version flag always succeeds and prints version
	properties.Property("Version flag prints version", prop.ForAll(
		func() bool {
			stdout := &bytes.Buffer{}
			stderr := &bytes.Buffer{}
			orch := &MockOrchestrator{}

			code := run([]string{"--version"}, stdout, stderr, orch)

			return code == 0 && strings.Contains(stdout.String(), "version")
		},
	))

	// Property: Invalid flags return non-zero exit code
	properties.Property("Invalid flags return error", prop.ForAll(
		func(invalidFlag string) bool {
			stdout := &bytes.Buffer{}
			stderr := &bytes.Buffer{}
			orch := &MockOrchestrator{}

			code := run([]string{"--" + invalidFlag}, stdout, stderr, orch)

			// Should return non-zero
			return code != 0
		},
		gen.Identifier().SuchThat(func(s interface{}) bool {
			// Filter out valid flags
			valid := map[string]bool{
				"output": true, "quality": true, "page-delay": true,
				"startup-delay": true, "pdf-quality": true, "mode": true,
				"trim-top": true, "trim-bottom": true, "trim-left": true, "trim-right": true,
				"page-turn-key": true, "verbose": true, "auto-confirm": true,
				"version": true, "help": true,
			}
			return !valid[s.(string)]
		}),
	))

	properties.TestingRun(t)
}
