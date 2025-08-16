package idgen

type IDGenerator interface {
	NewID(prefix string) string
}
