package gin_test

import (
	"testing"

	"github.com/zostay/arrest-go/internal/compilecost"
)

// TestOpaque checks that naming a type of this package costs an importer
// nothing: no exported type reaches libopenapi, and every function that
// touches it is //go:noinline. See arrest's state.go for why.
func TestOpaque(t *testing.T) {
	t.Parallel()

	violations, err := compilecost.Opaque(".")
	if err != nil {
		t.Fatal(err)
	}
	for _, v := range violations {
		t.Error(v)
	}
}
