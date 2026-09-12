package arrest_test

import (
	"bytes"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/zostay/arrest-go"
)

// A model with enough child components that a Go map's iteration order
// would rarely come out the same twice.
type orderedParent struct {
	Fig    orderedFig    `json:"fig" openapi:",refName=Fig"`
	Apple  orderedApple  `json:"apple" openapi:",refName=Apple"`
	Zebra  orderedZebra  `json:"zebra" openapi:",refName=Zebra"`
	Mango  orderedMango  `json:"mango" openapi:",refName=Mango"`
	Banana orderedBanana `json:"banana" openapi:",refName=Banana"`
	Kiwi   orderedKiwi   `json:"kiwi" openapi:",refName=Kiwi"`
}

type orderedFig struct{ Fig string }
type orderedApple struct{ Apple string }
type orderedZebra struct{ Zebra string }
type orderedMango struct{ Mango string }
type orderedBanana struct{ Banana string }
type orderedKiwi struct{ Kiwi string }

func buildOrdered(t *testing.T) []byte {
	t.Helper()

	doc, err := arrest.NewDocument("order")
	require.NoError(t, err)
	doc.PackageMap("fruit", "github.com/zostay/arrest-go_test")

	// Registered in the opposite of name order, and via both paths that
	// register child components.
	arrest.ModelFrom[orderedZebra](doc, arrest.WithComponentName("Zebra"))
	arrest.ModelFrom[orderedParent](doc, arrest.AsComponent())
	doc.SchemaComponent("Aardvark", arrest.ModelFrom[orderedApple](doc))
	doc.Get("/fruit").Response("200", func(r *arrest.Response) {
		r.Content("application/json", arrest.ModelFrom[orderedParent](doc))
	})
	require.NoError(t, doc.Err())

	bs, err := doc.Render()
	require.NoError(t, err)
	return bs
}

// TestRender_componentsAreStable renders the same handlers many times and
// expects the same bytes every time, so that a document regenerated to no
// change does not change.
func TestRender_componentsAreStable(t *testing.T) {
	t.Parallel()

	first := buildOrdered(t)
	for i := 0; i < 20; i++ {
		assert.True(t, bytes.Equal(first, buildOrdered(t)), "render %d differs from the first", i+1)
	}
}

// TestRender_componentsInNameOrder expects components in name order however
// they were registered.
func TestRender_componentsInNameOrder(t *testing.T) {
	t.Parallel()

	rendered := string(buildOrdered(t))
	_, components, ok := strings.Cut(rendered, "components:\n")
	require.True(t, ok, "no components in:\n%s", rendered)
	_, schemas, ok := strings.Cut(components, "  schemas:\n")
	require.True(t, ok)

	var names []string
	for _, line := range strings.Split(schemas, "\n") {
		if strings.HasPrefix(line, "    ") && !strings.HasPrefix(line, "     ") && strings.HasSuffix(line, ":") {
			names = append(names, strings.TrimSuffix(strings.TrimSpace(line), ":"))
		}
	}
	assert.Equal(t, []string{
		"Aardvark", "Zebra",
		"fruit.Apple", "fruit.Banana", "fruit.Fig", "fruit.Kiwi", "fruit.Mango", "fruit.Zebra",
		"fruit.orderedApple", "fruit.orderedBanana", "fruit.orderedFig", "fruit.orderedKiwi",
		"fruit.orderedMango", "fruit.orderedParent", "fruit.orderedZebra",
	}, names)

	// Refresh reloads what Render produced, so it keeps the order.
	doc, err := arrest.NewDocumentFromBytes([]byte(rendered))
	require.NoError(t, err)
	require.NoError(t, doc.Refresh())
	again, err := doc.Render()
	require.NoError(t, err)
	assert.Equal(t, rendered, string(again))
}
