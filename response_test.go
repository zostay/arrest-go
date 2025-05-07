package arrest_test

import (
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/zostay/arrest-go"
)

const contentMediaTypeExpectedDoc = `
openapi: 3.1.0
info:
  title: ContentMediaType Test
paths:
  '/download':
    get:
      responses:
        '200':
          description: A test response
          content:
            image/png: {}
            image/jpeg: {}
            image/webp: {}
        '406':
          description: Accept header not supported
`

func TestResponse_ContentMediaType(t *testing.T) {
	t.Parallel()

	doc, err := arrest.NewDocument("ContentMediaType Test")
	require.NoError(t, err)

	doc.Get("/download").
		Response("200", func(r *arrest.Response) {
			r.Description("A test response").
				ContentMediaType("image/png").
				ContentMediaType("image/jpeg").
				ContentMediaType("image/webp")
		}).
		Response("406", func(r *arrest.Response) {
			r.Description("Accept header not supported")
		})

	require.NoError(t, doc.Err())

	out, err := doc.OpenAPI.Render()
	require.NoError(t, err)
	require.YAMLEq(t, contentMediaTypeExpectedDoc, string(out))
	//require.Equal(t, contentMediaTypeExpectedDoc, string(out))
}
