package arrest_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/zostay/arrest-go"
)

func TestOperation_Deprecated(t *testing.T) {
	t.Parallel()

	doc, err := arrest.NewDocument("test")
	require.NoError(t, err)

	// Test deprecated operation
	doc.Get("/deprecated-endpoint").
		Summary("Deprecated endpoint").
		Deprecated()

	// Test non-deprecated operation
	doc.Get("/active-endpoint").
		Summary("Active endpoint")

	assert.NoError(t, doc.Err())

	openAPI, err := doc.OpenAPI.Render()
	require.NoError(t, err)

	spec := string(openAPI)

	// Check that the deprecated endpoint is marked as deprecated
	assert.Contains(t, spec, "/deprecated-endpoint")
	assert.Contains(t, spec, "deprecated: true")

	// Check that the active endpoint is not marked as deprecated
	assert.Contains(t, spec, "/active-endpoint")
	// The active endpoint should not have a deprecated field or it should be false
	// Since YAML omits false boolean values, we just check it doesn't have "deprecated: true"
	assert.NotContains(t, spec, "/active-endpoint:\n      get:\n        summary: Active endpoint\n        deprecated: true")
}
