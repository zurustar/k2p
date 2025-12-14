package config

// Config holds the application configuration
type Config struct {
	Output      string
	TempDir     string
	PageCount   int
	Direction   string
	MaxSizeStr  string
	MaxSize     int64
	CountDown   int
}
