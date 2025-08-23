package kafka

// SugaredLogger is a minimal interface compatible with zap.SugaredLogger
// Use this in shared publisher/consumer to avoid direct dependency on zap.
type SugaredLogger interface {
	Infow(msg string, keysAndValues ...any)
	Warnw(msg string, keysAndValues ...any)
	Errorw(msg string, keysAndValues ...any)
}
