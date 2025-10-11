package arrest_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/zostay/arrest-go"
)

type TestReq struct {
	Test TestType `json:"test" openapi:",refName=TestType"`
}

type TestType struct {
	Field string `json:"field"`
}

func TestComponentRefFromTagAlone(t *testing.T) {
	t.Parallel()

	const expected = `openapi: 3.1.0
info:
  title: ComponentRefFromTagAlone
paths:
  /test:
    post:
      requestBody:
        content:
          application/json:
            schema:
              type: object
              properties:
                test:
                  $ref: '#/components/schemas/test.v1.TestType'
components:
  schemas:
    test.v1.TestType:
      type: object
      properties:
        field:
          type: string
`

	doc, err := arrest.NewDocument("ComponentRefFromTagAlone")
	if err != nil {
		t.Fatalf("could not create document: %v", err)
	}

	doc.PackageMap(
		"test.v1", "github.com/zostay/arrest-go",
		"test.v1", "github.com/zostay/arrest-go_test",
		"test.v1", "command-line-arguments_test",
	)

	doc.Post("/test").
		RequestBody("application/json", arrest.ModelFrom[TestReq](doc))

	assert.NoError(t, doc.Err())
	got, err := doc.OpenAPI.Render()
	assert.NoError(t, err)
	assert.YAMLEq(t, expected, string(got))
	//assert.Equal(t, expected, string(got))
}
