package uuid_test

import (
	"testing"

	"github.com/jphastings/recipes/internal/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParseRoundTripsString(t *testing.T) {
	orig, err := uuid.NewUUID("a seed")
	require.NoError(t, err)

	parsed, err := uuid.Parse(orig.String())
	require.NoError(t, err)

	assert.Equal(t, orig, parsed)
	assert.Equal(t, orig.String(), parsed.String())
}

func TestParseRejectsMalformed(t *testing.T) {
	for _, s := range []string{"", "not-a-uuid", "zzzzzzzz-zzzz-zzzz-zzzz-zzzzzzzzzzzz"} {
		_, err := uuid.Parse(s)
		assert.Error(t, err, s)
	}
}
