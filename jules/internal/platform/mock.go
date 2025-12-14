package platform

import (
	"fmt"
	"os"
)

// MockPlatform implements the Platform interface for testing.
type MockPlatform struct {
	Screenshots []string
	KeyPresses  []string
}

// NewMockPlatform creates a new instance of MockPlatform.
func NewMockPlatform() *MockPlatform {
	return &MockPlatform{
		Screenshots: []string{},
		KeyPresses:  []string{},
	}
}

// Screenshot simulates capturing a screen by creating an empty file.
func (p *MockPlatform) Screenshot(filename string) error {
	p.Screenshots = append(p.Screenshots, filename)
	// Create a dummy file so file system checks pass
	f, err := os.Create(filename)
	if err != nil {
		return err
	}
	f.Close()
	return nil
}

// PressKey records the key press.
func (p *MockPlatform) PressKey(direction string) error {
	p.KeyPresses = append(p.KeyPresses, direction)
	if direction != "left" && direction != "right" {
		return fmt.Errorf("unknown direction: %s", direction)
	}
	return nil
}
