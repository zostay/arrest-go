package compilecost

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestOpaque_findsEveryLeak checks the checker against a package that breaks
// the rule in every way it must catch, and only those.
func TestOpaque_findsEveryLeak(t *testing.T) {
	t.Parallel()

	violations, err := Opaque("testdata/leaky")
	require.NoError(t, err)

	joined := strings.Join(violations, "\n")
	for _, want := range []string{
		"exported type ViaField reaches libopenapi: ViaField -> field M -> ",
		"exported type ViaMethod reaches libopenapi: ViaMethod.model -> ",
		"exported type ViaHidden reaches libopenapi: ViaHidden -> field h -> hidden -> field sp -> ",
		"Clean.BadBody is not //go:noinline but its body reaches libopenapi",
		"BadSig is not //go:noinline but its signature reaches libopenapi",
		"ViaMethod.model is not //go:noinline but its signature reaches libopenapi",
	} {
		assert.Contains(t, joined, want)
	}
	assert.NotContains(t, joined, "Clean reaches")
	assert.NotContains(t, joined, "Good ")
	assert.NotContains(t, joined, "Untouched")
	assert.Len(t, violations, 6, "\n%s", joined)
}
