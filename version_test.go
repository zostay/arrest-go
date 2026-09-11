package arrest_test

import (
	"regexp"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/zostay/arrest-go"
)

func TestVersion(t *testing.T) {
	t.Parallel()
	assert.Regexp(t, regexp.MustCompile(`^\d+\.\d+\.\d+`), arrest.Version)
}
