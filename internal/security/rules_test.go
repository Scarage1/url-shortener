package security

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRulesScanner_AllowsSafeURL(t *testing.T) {
	scanner := NewRulesScanner([]string{"malware.example"})

	err := scanner.Check("https://github.com/openai")

	require.NoError(t, err)
}

func TestRulesScanner_BlocksConfiguredDomain(t *testing.T) {
	scanner := NewRulesScanner([]string{"malware.example"})

	err := scanner.Check("https://login.malware.example/phish")

	assert.ErrorIs(t, err, ErrUnsafeURL)
}

func TestRulesScanner_BlocksLocalhost(t *testing.T) {
	scanner := NewRulesScanner(nil)

	err := scanner.Check("http://localhost:8080/admin")

	assert.ErrorIs(t, err, ErrUnsafeURL)
}
