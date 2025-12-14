package platform

// Platform defines the interface for OS-specific interactions.
type Platform interface {
	// Screenshot captures the screen and saves it to the specified filename.
	Screenshot(filename string) error
	// PressKey simulates a key press. Direction can be "left" or "right".
	PressKey(direction string) error
}
