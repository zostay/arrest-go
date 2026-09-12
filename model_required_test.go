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

type RequiredPtrEmbed struct {
	Origin string `json:"origin"`
}

type RequiredOverlap struct {
	ID    string `json:"id"`
	Extra string `json:"extra"`
}

type RequiredJot struct {
	RequiredAudit
	*RequiredPtrEmbed
	RequiredOverlap
	ID        string     `json:"id"`
	Title     string     `json:"title"`
	Version   int32      `json:"version"`
	Tags      []string   `json:"tags"`
	DeletedAt *time.Time `json:"deletedAt,omitempty"`
	Note      string     `json:"note,omitempty"`
	Zeroed    string     `json:"zeroed,omitzero"`
	Forced    *string    `json:"forced" openapi:",required"`
	Relaxed   string     `json:"relaxed" openapi:",optional"`
	Unforced  string     `json:"unforced" openapi:",required=false"`
	Extra     *string    `json:"extra,omitempty"`
	Strict    *string    `json:"strict" openapi:",optional=false"`
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
                  origin:
                    type: string
                  extra:
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
                  unforced:
                    type: string
                  strict:
                    type: string
                required:
                  - createdBy
                  - id
                  - title
                  - version
                  - tags
                  - forced
                  - strict
`

func TestModelFrom_Required(t *testing.T) {
	t.Parallel()

	doc, err := arrest.NewDocument("test")
	require.NoError(t, err)

	doc.Get("/jots").Response("200", func(r *arrest.Response) {
		r.Description("OK").Content("application/json", arrest.ModelFrom[RequiredJot](doc))
	})

	require.NoError(t, doc.Err())
	oas, err := doc.Render()
	require.NoError(t, err)
	assert.YAMLEq(t, expected_Required, string(oas))
}

type RequiredVehicle struct {
	VehicleType string       `json:"vehicleType" openapi:",discriminator,defaultMapping=car"`
	Car         *RequiredCar `json:"car" openapi:",oneOf,mapping=car"`
	Bike        RequiredBike `json:"bike" openapi:",oneOf,mapping=bike"`
	Van         *RequiredCar `json:"van" openapi:",oneOf,mapping=van,required"`
}

type RequiredCar struct {
	Doors int32 `json:"doors"`
}

type RequiredBike struct {
	Gears int32 `json:"gears"`
}

const expected_RequiredPolymorphic = `openapi: 3.1.0
info:
  title: test
paths:
  /vehicles:
    get:
      responses:
        "200":
          description: OK
          content:
            application/json:
              schema:
                oneOf:
                  - type: object
                    properties:
                      car:
                        type: object
                        properties:
                          doors:
                            type: integer
                            format: int32
                        required:
                          - doors
                  - type: object
                    properties:
                      bike:
                        type: object
                        properties:
                          gears:
                            type: integer
                            format: int32
                        required:
                          - gears
                    required:
                      - bike
                  - type: object
                    properties:
                      van:
                        type: object
                        properties:
                          doors:
                            type: integer
                            format: int32
                        required:
                          - doors
                    required:
                      - van
                discriminator:
                  propertyName: vehicleType
                  defaultMapping: car
                  mapping:
                    bike: '#/components/schemas/github.com.zostay.arrest-go_test.RequiredBike'
                    car: '#/components/schemas/github.com.zostay.arrest-go_test.RequiredCar'
                    van: '#/components/schemas/github.com.zostay.arrest-go_test.RequiredCar'
`

// Pointer polymorphic variants are optional; a value variant is required
// unless overridden.
func TestModelFrom_RequiredPolymorphic(t *testing.T) {
	t.Parallel()

	doc, err := arrest.NewDocument("test")
	require.NoError(t, err)

	doc.Get("/vehicles").Response("200", func(r *arrest.Response) {
		r.Description("OK").Content("application/json", arrest.ModelFrom[RequiredVehicle](doc))
	})

	require.NoError(t, doc.Err())
	oas, err := doc.Render()
	require.NoError(t, err)
	assert.YAMLEq(t, expected_RequiredPolymorphic, string(oas))
}
