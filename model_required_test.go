package arrest_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/zostay/arrest-go"
)

type RequiredAudit struct {
	CreatedBy string  `json:"createdBy"`
	UpdatedBy *string `json:"updatedBy"`
}

type RequiredJot struct {
	RequiredAudit
	ID        string     `json:"id"`
	Title     string     `json:"title"`
	Version   int32      `json:"version"`
	Tags      []string   `json:"tags"`
	DeletedAt *time.Time `json:"deletedAt,omitempty"`
	Note      string     `json:"note,omitempty"`
	Zeroed    string     `json:"zeroed,omitzero"`
	Forced    *string    `json:"forced" openapi:",required"`
	Relaxed   string     `json:"relaxed" openapi:",optional"`
	Hidden    string     `json:"hidden" openapi:"-"`
	Limit     int32      `json:"limit" openapi:",in=query"`
}

const expected_Required = `openapi: 3.1.0
info:
  title: test
paths:
  /jots:
    get:
      responses:
        "200":
          description: OK
          content:
            application/json:
              schema:
                type: object
                properties:
                  createdBy:
                    type: string
                  updatedBy:
                    type: string
                  id:
                    type: string
                  title:
                    type: string
                  version:
                    type: integer
                    format: int32
                  tags:
                    type: array
                    items:
                      type: string
                  deletedAt:
                    type: string
                    format: date-time
                  note:
                    type: string
                  zeroed:
                    type: string
                  forced:
                    type: string
                  relaxed:
                    type: string
                required:
                  - createdBy
                  - id
                  - title
                  - version
                  - tags
                  - forced
`

func TestModelFrom_Required(t *testing.T) {
	t.Parallel()

	doc, err := arrest.NewDocument("test")
	require.NoError(t, err)

	doc.Get("/jots").Response("200", func(r *arrest.Response) {
		r.Description("OK").Content("application/json", arrest.ModelFrom[RequiredJot](doc))
	})

	require.NoError(t, doc.Err())
	oas, err := doc.OpenAPI.Render()
	require.NoError(t, err)
	assert.YAMLEq(t, expected_Required, string(oas))
}
