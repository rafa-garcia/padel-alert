package util

import (
	"regexp"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGenerateID(t *testing.T) {
	// Generate multiple IDs to test
	id1 := GenerateID()
	id2 := GenerateID()
	id3 := GenerateID()

	// Test 1: Verify IDs are non-empty
	assert.NotEmpty(t, id1)
	assert.NotEmpty(t, id2)
	assert.NotEmpty(t, id3)

	// Test 2: Verify IDs are unique
	assert.NotEqual(t, id1, id2)
	assert.NotEqual(t, id2, id3)
	assert.NotEqual(t, id1, id3)

	// Test 3: Verify IDs match expected UUID format
	uuidRegex := regexp.MustCompile(`^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$`)
	assert.True(t, uuidRegex.MatchString(id1), "ID %s is not a valid UUID", id1)
	assert.True(t, uuidRegex.MatchString(id2), "ID %s is not a valid UUID", id2)
	assert.True(t, uuidRegex.MatchString(id3), "ID %s is not a valid UUID", id3)

	// Test 4: Verify ID length
	assert.Len(t, id1, 36) // Standard UUID length with hyphens
}
