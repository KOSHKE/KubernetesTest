package idgen

// IDGenerator abstracts ID creation logic
type IDGenerator interface {
	NewID(prefix string) string
}
