package config

import (
	"reflect"
	"testing"
	"time"

	"github.com/leanovate/gopter"
	"github.com/leanovate/gopter/gen"
	"github.com/leanovate/gopter/prop"
)

// Property 29: Default Configuration
// For any conversion when no configuration is specified, sensible default settings must be used.
func TestProperty29_DefaultConfiguration(t *testing.T) {
	parameters := gopter.DefaultTestParameters()
	parameters.MinSuccessfulTests = 100
	properties := gopter.NewProperties(parameters)

	properties.Property("ApplyDefaults ensures all fields have valid values", prop.ForAll(
		func(input *ConversionOptions) bool {
			// Apply defaults
			result := ApplyDefaults(input)

			// Check all required fields have valid values
			if result.ScreenshotQuality <= 0 || result.ScreenshotQuality > 100 {
				return false
			}
			if result.PageDelay <= 0 {
				return false
			}
			if result.StartupDelay < 0 {
				return false
			}
			if result.Mode != "generate" && result.Mode != "detect" {
				return false
			}
			if result.PageTurnKey != "right" && result.PageTurnKey != "left" {
				return false
			}
			if len(result.PDFQuality) == 0 {
				return false
			}

			return true
		},
		gen.PtrOf(gen.Struct(reflect.TypeOf(ConversionOptions{}), map[string]gopter.Gen{
			"OutputDir":         gen.AnyString(),
			"ScreenshotQuality": gen.IntRange(0, 100),
			"PageDelay":         gen.Int64Range(0, 10000000000).Map(func(i int64) time.Duration { return time.Duration(i) }),
			"StartupDelay":      gen.Int64Range(0, 10000000000).Map(func(i int64) time.Duration { return time.Duration(i) }),
			"ShowCountdown":     gen.Bool(),
			"PDFQuality":        gen.OneConstOf("", "low", "medium", "high"),
			"Verbose":           gen.Bool(),
			"AutoConfirm":       gen.Bool(),
			"Mode":              gen.OneConstOf("", "detect", "generate"),
			"TrimTop":           gen.Int(),
			"TrimBottom":        gen.Int(),
			"TrimLeft":          gen.Int(),
			"TrimRight":         gen.Int(),
			"PageTurnKey":       gen.OneConstOf("", "left", "right"),
		})),
	))

	properties.TestingRun(t)
}
