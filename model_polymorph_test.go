package arrest_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/zostay/arrest-go"
)

type Pet struct {
	PetType string        `json:"petType" openapi:",discriminator,defaultMapping=dog"`
	Dog     PolymorphDog  `json:",inline,omitempty" openapi:",oneOf,mapping=dog"`
	Cat     PolymorphCat  `json:",inline,omitempty" openapi:",oneOf,mapping=cat"`
	Bird    PolymorphBird `json:",inline,omitempty" openapi:",oneOf,mapping=bird"`
}

type PolymorphDog struct {
	Breed string `json:"breed"`
	Name  string `json:"name"`
}

type PolymorphCat struct {
	Lives int    `json:"lives"`
	Name  string `json:"name"`
}

type PolymorphBird struct {
	CanFly bool   `json:"canFly"`
	Name   string `json:"name"`
}

// Test with component references
type Vehicle struct {
	VehicleType string     `json:"vehicleType" openapi:",discriminator,defaultMapping=car"`
	Car         *Car       `json:"car,omitempty" openapi:",oneOf,mapping=car,refName=Car"`
	Truck       *Truck     `json:"truck,omitempty" openapi:",oneOf,mapping=truck,refName=Truck"`
	Motorcycle  Motorcycle `json:"motorcycle,omitempty" openapi:",oneOf,mapping=motorcycle,refName=Motorcycle"`
}

type Car struct {
	Doors int    `json:"doors"`
	Brand string `json:"brand"`
}

type Truck struct {
	Capacity int    `json:"capacity"`
	Brand    string `json:"brand"`
}

type Motorcycle struct {
	CCs   int    `json:"ccs"`
	Brand string `json:"brand"`
}

// Test anyOf composition
type FlexiblePet struct {
	PetType string `json:"petType" openapi:",discriminator,defaultMapping=mammal"`
	Mammal  `json:",inline,omitempty" openapi:",anyOf,mapping=mammal"`
	Bird    PolymorphBird `json:",inline,omitempty" openapi:",anyOf,mapping=bird"`
}

type Mammal struct {
	FurColor string `json:"furColor"`
	Name     string `json:"name"`
}

const expected_ImplicitPolymorphicPet = `openapi: 3.1.0
info:
  title: test
paths:
  /pets:
    post:
      requestBody:
        content:
          application/json:
            schema:
              oneOf:
              - type: object
                properties:
                  breed:
                    type: string
                  name:
                    type: string
              - type: object
                properties:
                  lives:
                    type: integer
                    format: int32
                  name:
                    type: string
              - type: object
                properties:
                  canFly:
                    type: boolean
                  name:
                    type: string
              discriminator:
                propertyName: petType
                mapping:
                  dog: '#/components/schemas/github.com/zostay/arrest-go_test.PolymorphDog'
                  cat: '#/components/schemas/github.com/zostay/arrest-go_test.PolymorphCat'
                  bird: '#/components/schemas/github.com/zostay/arrest-go_test.PolymorphBird'
                defaultMapping: dog
`

func TestImplicitPolymorphicPet(t *testing.T) {
	t.Parallel()

	doc, err := arrest.NewDocument("test")
	require.NoError(t, err)

	petModel := arrest.ModelFrom[Pet](doc)
	assert.NoError(t, petModel.Err())

	doc.Post("/pets").
		RequestBody("application/json", petModel)

	assert.NoError(t, doc.Err())

	oas, err := doc.OpenAPI.Render()
	require.NoError(t, err)

	assert.YAMLEq(t, expected_ImplicitPolymorphicPet, string(oas))
}

const expected_ImplicitPolymorphicVehicle = `openapi: 3.1.0
info:
  title: test
paths:
  /vehicles:
    post:
      requestBody:
        content:
          application/json:
            schema:
              oneOf:
              - $ref: '#/components/schemas/github.com/zostay/arrest-go_test.Car'
              - $ref: '#/components/schemas/github.com/zostay/arrest-go_test.Truck'
              - $ref: '#/components/schemas/github.com/zostay/arrest-go_test.Motorcycle'
              discriminator:
                propertyName: vehicleType
                mapping:
                  car: '#/components/schemas/Car'
                  truck: '#/components/schemas/Truck'
                  motorcycle: '#/components/schemas/Motorcycle'
                defaultMapping: car
components:
  schemas:
    github.com.zostay.arrest-go_test.Car:
      type: object
      properties:
        doors:
          type: integer
          format: int32
        brand:
          type: string
    github.com.zostay.arrest-go_test.Truck:
      type: object
      properties:
        capacity:
          type: integer
          format: int32
        brand:
          type: string
    github.com.zostay.arrest-go_test.Motorcycle:
      type: object
      properties:
        ccs:
          type: integer
          format: int32
        brand:
          type: string
`

func TestImplicitPolymorphicVehicleWithComponents(t *testing.T) {
	t.Parallel()

	doc, err := arrest.NewDocument("test")
	require.NoError(t, err)

	vehicleModel := arrest.ModelFrom[Vehicle](doc)
	assert.NoError(t, vehicleModel.Err())

	doc.Post("/vehicles").
		RequestBody("application/json", vehicleModel)

	assert.NoError(t, doc.Err())

	oas, err := doc.OpenAPI.Render()
	require.NoError(t, err)

	assert.YAMLEq(t, expected_ImplicitPolymorphicVehicle, string(oas))
}

const expected_ImplicitPolymorphicAnyOf = `openapi: 3.1.0
info:
  title: test
paths:
  /flexible-pets:
    post:
      requestBody:
        content:
          application/json:
            schema:
              anyOf:
              - type: object
                properties:
                  furColor:
                    type: string
                  name:
                    type: string
              - type: object
                properties:
                  canFly:
                    type: boolean
                  name:
                    type: string
              discriminator:
                propertyName: petType
                mapping:
                  mammal: '#/components/schemas/github.com/zostay/arrest-go_test.Mammal'
                  bird: '#/components/schemas/github.com/zostay/arrest-go_test.PolymorphBird'
                defaultMapping: mammal
`

func TestImplicitPolymorphicAnyOf(t *testing.T) {
	t.Parallel()

	doc, err := arrest.NewDocument("test")
	require.NoError(t, err)

	petModel := arrest.ModelFrom[FlexiblePet](doc)
	assert.NoError(t, petModel.Err())

	doc.Post("/flexible-pets").
		RequestBody("application/json", petModel)

	assert.NoError(t, doc.Err())

	oas, err := doc.OpenAPI.Render()
	require.NoError(t, err)

	assert.YAMLEq(t, expected_ImplicitPolymorphicAnyOf, string(oas))
}
