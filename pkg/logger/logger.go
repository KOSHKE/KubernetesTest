package logger

// Logger provides unified logging interface for all packages
type Logger interface {
	// Error logs error with structured fields
	Error(msg string, keysAndValues ...any)

	// Warn logs warning with structured fields
	Warn(msg string, keysAndValues ...any)

	// Info logs info with structured fields
	Info(msg string, keysAndValues ...any)

	// Debug logs debug with structured fields
	Debug(msg string, keysAndValues ...any)
}
