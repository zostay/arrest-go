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

	assert.Equal(t, "DocStruct is a structure with documentation.\n", doc)

	assert.Equal(t, map[string]string{
		"Foo": "Foo is a field.\n",
		"Bar": "Bar is also a field.\n",
	}, flds)
}

func TestGoDocForStruct_SadNotAStruct(t *testing.T) {
	t.Parallel()

	_, _, err := arrest.GoDocForStruct(reflect.TypeOf(1))
	assert.ErrorContains(t, err, "expected a struct type")
}

func BenchmarkGoDocForStruct(b *testing.B) {
	b.ReportAllocs()
	typeToTest := reflect.TypeOf(test.DocStruct{})
	for i := 0; i < b.N; i++ {
		_, _, err := arrest.GoDocForStruct(typeToTest)
		if err != nil {
			b.Fatalf("unexpected error: %v", err)
		}
	}
}
