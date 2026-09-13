package compilecost

import (
	"os"
	"path/filepath"
	"runtime"
	"testing"

	"github.com/stretchr/testify/require"
)

// MaxFuncs is the most functions any guarded probe may emit. It is a
// ratchet: it sits just above the worst probe today so that a change which
// widens the cascade fails here.
//
// A probe that names an opaque DSL type emits a few dozen functions (the
// generic set behind ErrHelper, and gin's own); one that reached libopenapi's
// model would emit around 9,000. Anything in between means a type or an
// inlinable body has started to reach the model — TestOpaque in the DSL
// packages says where.
const MaxFuncs = 100

// TestGuard builds a probe package for every public type and fails if any of
// them emits more functions than MaxFuncs. It builds real packages, so it is
// skipped under -short.
func TestGuard(t *testing.T) {
	if testing.Short() {
		t.Skip("builds probe packages; skipped under -short")
	}

	_, file, _, ok := runtime.Caller(0)
	require.True(t, ok)
	root := filepath.Join(filepath.Dir(file), "..", "..")

	probes, err := Probes(root)
	require.NoError(t, err)
	require.NotEmpty(t, probes)

	dir := t.TempDir()
	mod, err := WriteModule(dir, root, probes)
	require.NoError(t, err)
	if testing.Verbose() {
		mod.Log = t.Logf
	}

	results, err := mod.Measure(false)
	require.NoError(t, err)

	for _, r := range results {
		t.Logf("%-28s %6d funcs", r.Probe.Name, r.Funcs)
		if r.Probe.Guarded && r.Funcs > MaxFuncs {
			t.Errorf("%s emits %d functions, more than %d: something newly reachable from %s reaches libopenapi's model",
				r.Probe.Name, r.Funcs, MaxFuncs, r.Probe.Type)
		}
	}
}

// TestProbes_findsEveryType checks that the parser-driven enumeration sees
// the types it must, so a probe list that silently came up empty would not
// pass the guard.
func TestProbes_findsEveryType(t *testing.T) {
	t.Parallel()

	_, file, _, ok := runtime.Caller(0)
	require.True(t, ok)
	root := filepath.Join(filepath.Dir(file), "..", "..")

	probes, err := Probes(root)
	require.NoError(t, err)

	names := map[string]bool{}
	for _, p := range probes {
		names[p.Name] = true
	}
	for _, want := range []string{
		"sig_arrest_document", "sig_arrest_operation", "sig_arrest_model",
		"sig_arrestgin_document", "sig_arrestgin_operation",
		"use_arrest", "use_gin", "ref_blank", "ref_v3",
	} {
		require.True(t, names[want], "missing probe %s", want)
	}

	_, err = os.Stat(filepath.Join(root, "go.mod"))
	require.NoError(t, err, "root should be the module root")
}
