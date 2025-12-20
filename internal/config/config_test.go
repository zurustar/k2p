package config

import (
	"testing"
	"time"
)

func TestApplyDefaults(t *testing.T) {
	t.Run("Nil input should return defaults", func(t *testing.T) {
		defaults := ApplyDefaults(nil)

		if defaults.ScreenshotQuality != 95 {
			t.Errorf("Expected default quality 95, got %d", defaults.ScreenshotQuality)
		}
		if defaults.PageDelay != 500*time.Millisecond {
			t.Errorf("Expected default page delay 500ms, got %v", defaults.PageDelay)
		}
		if defaults.Mode != "generate" {
			t.Errorf("Expected default mode 'generate', got %s", defaults.Mode)
		}
		if defaults.PageTurnKey != "right" {
			t.Errorf("Expected default page turn key 'right', got %s", defaults.PageTurnKey)
		}
		if !defaults.ShowCountdown {
			t.Error("Expected default ShowCountdown=true")
		}
	})

	t.Run("Override defaults with provided values", func(t *testing.T) {
		input := &ConversionOptions{
			ScreenshotQuality: 80,
			PageDelay:         1 * time.Second,
			Verbose:           true,
			Mode:              "detect",
			TrimTop:           50,
			PageTurnKey:       "left",
		}

		merged := ApplyDefaults(input)

		if merged.ScreenshotQuality != 80 {
			t.Errorf("Expected overridden quality 80, got %d", merged.ScreenshotQuality)
		}
		if merged.PageDelay != 1*time.Second {
			t.Errorf("Expected overridden page delay 1s, got %v", merged.PageDelay)
		}
		if !merged.Verbose {
			t.Error("Expected verbose to be true")
		}
		if merged.Mode != "detect" {
			t.Errorf("Expected mode 'detect', got %s", merged.Mode)
		}
		if merged.TrimTop != 50 {
			t.Errorf("Expected trim top 50, got %d", merged.TrimTop)
		}
		if merged.PageTurnKey != "left" {
			t.Errorf("Expected page turn key 'left', got %s", merged.PageTurnKey)
		}

		// Defaults should still be present for unset fields
		if merged.PDFQuality != "high" {
			t.Errorf("Expected default PDF quality 'high', got %s", merged.PDFQuality)
		}
	})

	t.Run("Partial overrides", func(t *testing.T) {
		input := &ConversionOptions{
			ScreenshotQuality: 100,
		}

		merged := ApplyDefaults(input)

		if merged.ScreenshotQuality != 100 {
			t.Errorf("Expected quality 100, got %d", merged.ScreenshotQuality)
		}
		if merged.PageDelay != 500*time.Millisecond {
			t.Errorf("Expected default page delay, got %v", merged.PageDelay)
		}
	})
}
