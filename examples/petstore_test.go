package main

import (
	_ "embed"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

//go:embed petstore.yaml
var expected string

func TestBuildDoc(t *testing.T) {
	var got string
	require.NotPanics(t, func() {
		got = BuildDocString()
	})

	assert.YAMLEq(t, expected, got)
	//assert.Equal(t, expected, got)
}
