package main

import (
	_ "embed"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

//go:embed petstore.yaml
var expected string

func TestBuildDoc(t *testing.T) {
	e := gin.Default()
	doc, err := BuildDoc(e)
	require.NoError(t, err)

	got, err := doc.OpenAPI.Render()
	require.NoError(t, err)
	assert.YAMLEq(t, expected, string(got))
}
