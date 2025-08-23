package redisclient

import "go.uber.org/zap"

// SugaredLoggerAdapter adapts *zap.SugaredLogger to our Logger interface
type SugaredLoggerAdapter struct {
	log *zap.SugaredLogger
}

// NewSugaredLoggerAdapter creates adapter for existing sugared logger
func NewSugaredLoggerAdapter(log *zap.SugaredLogger) Logger {
	return &SugaredLoggerAdapter{log: log}
}

// Error logs error with structured fields
func (l *SugaredLoggerAdapter) Error(msg string, keysAndValues ...interface{}) {
	l.log.Errorw(msg, keysAndValues...)
}

// Warn logs warning with structured fields
func (l *SugaredLoggerAdapter) Warn(msg string, keysAndValues ...interface{}) {
	l.log.Warnw(msg, keysAndValues...)
}
