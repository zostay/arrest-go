package arrest_test

import (
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/zostay/arrest-go"
	"github.com/zostay/arrest-go/internal/test"
)

func TestGoDocForStruct(t *testing.T) {
	t.Parallel()

	doc, flds, err := arrest.GoDocForStruct(reflect.TypeOf(test.DocStruct{}))
	require.NoError(t, err)

	require.Equal(t, "DocStruct is a structure with documentation.\n", doc)
	require.Len(t, flds, 0)
}

func TestGoDocForStruct_SadNotAStruct(t *testing.T) {
	t.Parallel()

	_, _, err := arrest.GoDocForStruct(reflect.TypeOf(1))
	assert.ErrorContains(t, err, "expected a struct type")
}
