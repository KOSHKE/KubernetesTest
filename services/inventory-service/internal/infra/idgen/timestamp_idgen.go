package idgen

import "time"

type TimestampIDGenerator struct{}

func NewTimestampIDGenerator() *TimestampIDGenerator { return &TimestampIDGenerator{} }
func (TimestampIDGenerator) NewID(prefix string) string {
	return prefix + time.Now().Format("20060102150405")
}
