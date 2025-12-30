package sound

import (
	"os/exec"
)

// Player defines the interface for playing sounds
type Player interface {
	PlaySuccess() error
	PlayError() error
}

// DefaultPlayer plays sounds using the system's afplay command (macOS)
type DefaultPlayer struct{}

// NewPlayer creates a new sound player
func NewPlayer() Player {
	return &DefaultPlayer{}
}

// PlaySuccess plays the success sound
func (p *DefaultPlayer) PlaySuccess() error {
	return p.play("/System/Library/Sounds/Glass.aiff")
}

// PlayError plays the error sound
func (p *DefaultPlayer) PlayError() error {
	return p.play("/System/Library/Sounds/Basso.aiff")
}

func (p *DefaultPlayer) play(soundPath string) error {
	// We use Start() to play in background and ignore errors/output
	// because sound playback is non-critical and shouldn't block
	return exec.Command("afplay", soundPath).Start()
}

// NoOpPlayer is a sound player that does nothing (for testing)
type NoOpPlayer struct{}

// NewNoOpPlayer creates a new no-op sound player
func NewNoOpPlayer() Player {
	return &NoOpPlayer{}
}

// PlaySuccess does nothing
func (p *NoOpPlayer) PlaySuccess() error {
	return nil
}

// PlayError does nothing
func (p *NoOpPlayer) PlayError() error {
	return nil
}
