package main

import (
	"fmt"
	"reflect"

	"github.com/zostay/arrest-go"
)

type CreatePetRequest struct {
	Pet Pet `json:"pet" openapi:",refName=Pet"`
}

type Pet struct {
	ID   int64  `json:"id"`
	Name string `json:"name"`
	Tag  string `json:"tag"`
}

type Pets []Pet

type Error struct {
	Code    int32  `json:"code"`
	Message string `json:"message"`
}

func ListPets(limit int32) (Pets, error) {
	// This is where you would put your implementation of ListPets
	return nil, nil
}

func CreatePets(pet Pet) error {
	// This is where you would put your implementation of CreatePets
	return nil
}

func ShowByPetID(petID string) (*Pet, error) {
	// This is where you would put your implementation of ShowByPetID
	return nil, nil
}

func main() {
	fmt.Print(BuildDocString())
}

func BuildDoc() (*arrest.Document, error) {
	doc, err := arrest.NewDocument("Swagger Petstore")
	if err != nil {
		return nil, err
	}

	doc.AddServer("http://petstore.swagger.io/v1")
	doc.PackageMap(
		"pet.v1", "github.com/zostay/arrest-go/examples",
		"pet.v1", "main",
	)

	listPets := arrest.ParametersFromReflect(reflect.TypeOf(ListPets)).
		P(0, func(p *arrest.Parameter) {
			p.Name("limit").In("query").Required()
		})

	doc.Get("/pets").
		Summary("List all pets").
		OperationID("listPets").
		Tags("pets").
		Parameters(listPets).
		Response("200", func(r *arrest.Response) {
			r.Description("A list of pets").
				Header("x-next", arrest.ModelFrom[string](doc), func(h *arrest.Header) {
					h.Description("A link to the next page of responses")
				}).
				Content("application/json", arrest.ModelFrom[Pets](doc))
		}).
		Response("default", func(r *arrest.Response) {
			r.Description("unexpected error").
				Content("application/json", arrest.ModelFrom[Error](doc))
		})

	createPetModel := arrest.ModelFrom[CreatePetRequest](doc, arrest.AsComponent())

	doc.Post("/pets").
		Summary("Create a pet").
		OperationID("createPets").
		Tags("pets").
		RequestBody("application/json", arrest.SchemaRef(createPetModel.MappedName(doc.PkgMap))).
		Response("201", func(r *arrest.Response) { r.Description("Null response") }).
		Response("default", func(r *arrest.Response) {
			r.Description("unexpected error").
				Content("application/json", arrest.ModelFrom[Error](doc))
		})

	showByPetId := arrest.ParametersFromReflect(reflect.TypeOf(ShowByPetID)).
		P(0, func(p *arrest.Parameter) {
			p.Name("petId").In("path").Required().
				Description("The ID of the pet to retrieve")
		})

	doc.Get("/pets/{petId}").
		Summary("Info for a specific pet").
		OperationID("showByPetId").
		Tags("pets").
		Parameters(showByPetId).
		Response("200", func(r *arrest.Response) {
			r.Description("Expected response to a valid request").
				Content("application/json", arrest.ModelFrom[Pet](doc))
		}).
		Response("default", func(r *arrest.Response) {
			r.Description("unexpected error").
				Content("application/json", arrest.ModelFrom[Error](doc))
		})

	if doc.Err() != nil {
		return nil, doc.Err()
	}

	return doc, nil
}

func BuildDocString() string {
	doc, err := BuildDoc()
	if err != nil {
		panic(err)
	}

	rend, err := doc.OpenAPI.Render()
	if err != nil {
		panic(err)
	}

	return string(rend)
}
