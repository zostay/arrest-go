# AR! Rest!

The pirate-y OpenAPI 3.0 spec generator for Go. This is built on the very excellent libopenapi library from pb33f. Five out of five stars. Highly recommend. It can handle just about any kind of weird, fancy, or crazy OpenAPI doc you want to write, but it's also like programming your VCR. That's a dated references, but the retro styling of his website suggests.

Anyway, this provides a DSL for generating OpenAPI 3.0 specs in Go. That, by itself, is probably not very interesting, but it does help infer your schemas from Go functions and Go types, which can greatly simplify things. Consider the ubiquitous Petstore example, here using the ARRest DSL:

```go
package main

import (
	"fmt"
	"reflect"

	"github.com/zostay/arrest-go"
)

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
	doc := arrest.NewDocument("Swagger Petstore").
		AddServer("http://petstore.swagger.io/v1")

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
				Header("x-next", arrest.ModelFrom[string](), func(h *arrest.Header) {
					h.Description("A link to the next page of responses")
				}).
				Content("application/json", arrest.ModelFrom[Pets]())
		}).
		Response("default", func(r *arrest.Response) {
			r.Description("unexpected error").
				Content("application/json", arrest.ModelFrom[Error]())
		})

	doc.Post("/pets").
		Summary("Create a pet").
		OperationID("createPets").
		Tags("pets").
		RequestBody("application/json", arrest.ModelFrom[Pet]()).
		Response("201", func(r *arrest.Response) { r.Description("Null response") }).
		Response("default", func(r *arrest.Response) {
			r.Description("unexpected error").
				Content("application/json", arrest.ModelFrom[Error]())
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
				Content("application/json", arrest.ModelFrom[Pet]())
		}).
		Response("default", func(r *arrest.Response) {
			r.Description("unexpected error").
				Content("application/json", arrest.ModelFrom[Error]())
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

```

# Features

The aim is to be able to speed up building your specs while still giving you full flexibility. This tool is aimed at the high level. If it doesn't do something yet and you think it should, PRs welcome. However, in the meantime, you can just use the `OpenAPI` struct directly to modify it using "hard mode". In face, if you can come up with the standardized method this way first, that will make the PR all the easier.

# The DSL

That's a Domain Specific Language for those who may not know. This library is based around building a sort of DSL inside of Go. It makes heavy use of method chaining to help abbreviate the code necessary to build a spec. This takes a little getting used.

The chaining is setup to be used in paragraphs or stanzas around the primary components. We consider the following pieces to be primary components:

* Document
* Operation

Therefore, the encouraged style is to build up the Document and then build up each Operation. Each primary component returns the component object and then all the methods of the primary only return that object.

```go
doc := arrest.NewDocument("My API").AddServer("http://example.com")

doc.Get("/path").Summary("Get a thing") // and on

doc.Post("/path").Summary("Post a thing") // and on
```

Within each primary there are secondary components. To avoid breaking up your primary paragraphs, these do not return the different types for configuration but make use of callbacks or inner constructors.

```go
doc.Get("/path").Summary("Get a thing").
    OperationID("getThing").
    Parameters(
        arrest.ParametersFromReflect(reflect.TypeOf(GetThing)).
            p.P(0, func(p *arrest.Parameter) {
                p.Name("id").In("query").Required()
            })
    ).
    Response("200", func(r *arrest.Response) {
        r.Description("The thing").
            Content("application/json", arrest.ModelFrom[Thing]())
    })
```

Finally, when you're finished you need to check to see if errors occurred. The system should not panic, but it records errors as it goes. You can check for them at the end, or any time in the middle by calling the `Err()` method on the document. You can check for errors in operations or other parts as well. Errors all flow upward to the parent component, but you can check them at any level:

```go
if doc.Err() != nil {
    // handle the error here, of course
}
```

# Installation

```shell
go get github.com/zostay/arrest-go
```

# Special Thanks

Thanks to [pb33f](https://pb33f.io/) for the excellent [libopenapi](https://github.com/pb33f/libopenapi) library. This library is built on top of that one.

Also thanks to [Adrian Hesketh](https://github.com/a-h) whose [github.com/a-h/rest](https://github.com/a-h/rest) library provided some inspiration for this library.
