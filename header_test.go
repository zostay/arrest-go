package arrest_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/zostay/arrest-go"
)

const headerTestExpectedDoc = `
openapi: 3.1.0
info:
  title: Header Test
paths:
  '/header':
    get:
      responses:
        '200':
          description: A test response
          headers:
            x-foo:
              description: A test header
              schema:
                type: string
`

func TestHeader(t *testing.T) {
	t.Parallel()

	doc, err := arrest.NewDocument("Header Test")
	require.NoError(t, err)

	doc.Get("/header").
		Response("200", func(r *arrest.Response) {
			r.Description("A test response").
				Header("x-foo", arrest.ModelFrom[string](), func(h *arrest.Header) {
					h.Description("A test header")
				})
		})

	require.NoError(t, doc.Err())

	out, err := doc.OpenAPI.Render()
	require.NoError(t, err)
	assert.YAMLEq(t, headerTestExpectedDoc, string(out))
}
