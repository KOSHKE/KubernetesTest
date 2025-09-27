//go:build integration

package helpers

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// AssertStatusCode asserts HTTP status code equals expected
func AssertStatusCode(t *testing.T, got, expected int, body string) {
	assert.Equalf(t, expected, got, "unexpected status code, body: %s", body)
}
