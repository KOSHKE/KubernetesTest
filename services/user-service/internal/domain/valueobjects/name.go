package valueobjects

import (
	"strings"
)

type Name struct {
	value string
}

func NewName(name string) Name {
	name = strings.TrimSpace(name)
	return Name{value: name}
}

func (n Name) Value() string {
	return n.value
}

func (n Name) String() string {
	return n.value
}

func (n Name) Equals(other Name) bool {
	return n.value == other.value
}
