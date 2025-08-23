package logger

// Logger provides unified logging interface for all packages
type Logger interface {
	// Error logs error with structured fields
	Error(msg string, keysAndValues ...interface{})

	// Warn logs warning with structured fields
	Warn(msg string, keysAndValues ...interface{})

	// Info logs info with structured fields
	Info(msg string, keysAndValues ...interface{})

	// Debug logs debug with structured fields
	Debug(msg string, keysAndValues ...interface{})
}
