package main

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const expected = `
openapi: 3.1.0
info:
  title: Swagger Petstore
servers:
  - url: http://petstore.swagger.io/v1
paths:
  /pets:
    get:
      summary: List all pets
      operationId: listPets
      tags:
        - pets
      parameters:
        - name: limit
          in: query
          required: true
          schema:
            type: integer
            format: int32
      responses:
        '200':
          description: A list of pets
          headers:
            x-next:
              description: A link to the next page of responses
              schema:
                type: string
          content:
            application/json:
              schema:
                type: array
                items:
                  type: object
                  properties:
                    id:
                      type: integer
                      format: int64
                    name:
                      type: string
                    tag:
                      type: string
        default:
          description: unexpected error
          content:
            application/json:
              schema:
                type: object
                properties:
                  code:
                    type: integer
                    format: int32
                  message:
                    type: string
    post:
      summary: Create a pet
      operationId: createPets
      tags:
        - pets
      requestBody:
        content:
          application/json:
            schema:
              type: object
              properties:
                id:
                  type: integer
                  format: int64
                name:
                  type: string
                tag:
                  type: string
      responses:
        '201':
          description: Null response
        default:
          description: unexpected error
          content:
            application/json:
              schema:
                type: object
                properties:
                  code:
                    type: integer
                    format: int32
                  message:
                    type: string
  "/pets/{petId}":
    get:
      summary: Info for a specific pet
      operationId: showByPetId
      tags:
        - pets
      parameters:
        - name: petId
          in: path
          required: true
          description: The ID of the pet to retrieve
          schema:
            type: string
      responses:
        '200':
          description: Expected response to a valid request
          content:
            application/json:
              schema:
                type: object
                properties:
                  id:
                    type: integer
                    format: int64
                  name:
                    type: string
                  tag:
                    type: string
        default:
          description: unexpected error
          content:
            application/json:
              schema:
                type: object
                properties:
                  code:
                    type: integer
                    format: int32
                  message:
                    type: string
`

func TestBuildDoc(t *testing.T) {
	var got string
	require.NotPanics(t, func() {
		got = BuildDocString()
	})

	assert.YAMLEq(t, expected, got)
}
