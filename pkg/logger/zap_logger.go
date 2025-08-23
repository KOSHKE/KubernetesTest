package logger

import "go.uber.org/zap"

// ZapLogger implements Logger interface using zap.SugaredLogger
type ZapLogger struct {
	log *zap.SugaredLogger
}

// NewZapLogger creates new logger using zap.SugaredLogger
func NewZapLogger(log *zap.SugaredLogger) Logger {
	return &ZapLogger{log: log}
}

// Error logs error with structured fields
func (z *ZapLogger) Error(msg string, keysAndValues ...interface{}) {
	z.log.Errorw(msg, keysAndValues...)
}

// Warn logs warning with structured fields
func (z *ZapLogger) Warn(msg string, keysAndValues ...interface{}) {
	z.log.Warnw(msg, keysAndValues...)
}

// Info logs info with structured fields
func (z *ZapLogger) Info(msg string, keysAndValues ...interface{}) {
	z.log.Infow(msg, keysAndValues...)
}

// Debug logs debug with structured fields
func (z *ZapLogger) Debug(msg string, keysAndValues ...interface{}) {
	z.log.Debugw(msg, keysAndValues...)
}
