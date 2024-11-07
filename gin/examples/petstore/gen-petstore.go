package main

import (
	"fmt"
	"reflect"

	"github.com/zostay/arrest-go"
	"github.com/zostay/arrest-go/gin/examples/petstore/message"
)

func main() {
	doc, err := arrest.NewDocument("Swagger Petstore")
	if err != nil {
		panic(err)
	}

	doc.AddServer("http://petstore.swagger.io/v1")

	listPets := arrest.ParametersFromReflect(reflect.TypeOf(message.ListPets)).
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
				Header("x-next", arrest.ModelFrom[string](), func(h *arrest.Header) {
					h.Description("A link to the next page of responses")
				}).
				Content("application/json", arrest.ModelFrom[message.Pets]())
		}).
		Response("default", func(r *arrest.Response) {
			r.Description("unexpected error").
				Content("application/json", arrest.ModelFrom[message.Error]())
		})

	doc.Post("/pets").
		Summary("Create a pet").
		OperationID("createPets").
		Tags("pets").
		RequestBody("application/json", arrest.ModelFrom[message.Pet]()).
		Response("201", func(r *arrest.Response) { r.Description("Null response") }).
		Response("default", func(r *arrest.Response) {
			r.Description("unexpected error").
				Content("application/json", arrest.ModelFrom[message.Error]())
		})

	showByPetId := arrest.ParametersFromReflect(reflect.TypeOf(message.ShowByPetID)).
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
				Content("application/json", arrest.ModelFrom[message.Pet]())
		}).
		Response("default", func(r *arrest.Response) {
			r.Description("unexpected error").
				Content("application/json", arrest.ModelFrom[message.Error]())
		})

	if doc.Err() != nil {
		panic(doc.Err())
	}

	rend, err := doc.OpenAPI.Render()
	if err != nil {
		panic(err)
	}

	fmt.Print(string(rend))

}
