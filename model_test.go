package arrest_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/zostay/arrest-go"
)

const expected = `openapi: 3.1.0
info:
  title: test
paths:
  /simple:
    put:
      requestBody:
        content:
          text/plain:
            schema:
              type: string
              oneOf:
              - title: Foo
                const: foo
                description: Foo is a test value
              - const: bar
                title: Bar
                description: Bar is a test value
`

func TestModelFrom_WithOneOf(t *testing.T) {
	t.Parallel()

	myEnum := arrest.ModelFrom[string]().OneOf(
		arrest.Enumeration{
			Const:       "foo",
			Title:       "Foo",
			Description: "Foo is a test value",
		},
		arrest.Enumeration{
			Const:       "bar",
			Title:       "Bar",
			Description: "Bar is a test value",
		},
	)

	assert.NoError(t, myEnum.Err())

	doc, err := arrest.NewDocument("test")
	require.NoError(t, err)

	doc.Put("/simple").
		RequestBody("text/plain", myEnum)

	assert.NoError(t, doc.Err())

	oas, err := doc.OpenAPI.Render()
	require.NoError(t, err)

	assert.YAMLEq(t, expected, string(oas))
	//assert.Equal(t, expected, string(oas))
}
