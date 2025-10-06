package gin

import (
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/zostay/arrest-go"
)

func TestGinOperation_Deprecated(t *testing.T) {
	t.Parallel()

	arrestDoc, err := arrest.NewDocument("test")
	require.NoError(t, err)

	router := gin.New()
	doc := NewDocument(arrestDoc, router)

	// Test deprecated operation
	doc.Get("/deprecated-endpoint").
		Summary("Deprecated endpoint").
		Deprecated()

	// Test non-deprecated operation
	doc.Get("/active-endpoint").
		Summary("Active endpoint")

	assert.NoError(t, arrestDoc.Err())

	openAPI, err := arrestDoc.OpenAPI.Render()
	require.NoError(t, err)

	spec := string(openAPI)

	// Check that the deprecated endpoint is marked as deprecated
	assert.Contains(t, spec, "/deprecated-endpoint")
	assert.Contains(t, spec, "deprecated: true")

	// Check that the active endpoint is not marked as deprecated
	assert.Contains(t, spec, "/active-endpoint")
}
