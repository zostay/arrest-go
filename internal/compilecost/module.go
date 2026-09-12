package compilecost

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

// Module is a probe module on disk: a go.mod that replaces arrest-go and
// arrest-go/gin with the checkout under Root, and one directory per probe.
type Module struct {
	Dir    string
	Root   string
	Probes []Probe
	// Env is added to every go command's environment.
	Env []string
	// Log, when set, receives the commands as they run.
	Log func(format string, args ...any)
}

// WriteModule writes a probe module for the checkout at root into dir and
// resolves its dependencies, so that it is ready to build.
func WriteModule(dir, root string, probes []Probe) (*Module, error) {
	root, err := filepath.Abs(root)
	if err != nil {
		return nil, err
	}

	gomod := fmt.Sprintf(`module probe

go %s

require (
	github.com/zostay/arrest-go v0.0.0
	github.com/zostay/arrest-go/gin v0.0.0
)

replace github.com/zostay/arrest-go => %s

replace github.com/zostay/arrest-go/gin => %s
`, goVersion(root), root, filepath.Join(root, "gin"))
	if err := os.WriteFile(filepath.Join(dir, "go.mod"), []byte(gomod), 0o644); err != nil {
		return nil, err
	}

	for _, p := range probes {
		pdir := filepath.Join(dir, p.Name)
		if err := os.MkdirAll(pdir, 0o755); err != nil {
			return nil, err
		}
		if err := os.WriteFile(filepath.Join(pdir, "probe.go"), []byte(p.Source), 0o644); err != nil {
			return nil, err
		}
	}

	m := &Module{Dir: dir, Root: root, Probes: probes}
	if _, err := m.run("mod", "tidy"); err != nil {
		return nil, err
	}
	return m, nil
}

// goVersion reads the go directive from the checkout's go.mod so the probe
// module builds with the same language version.
func goVersion(root string) string {
	bs, err := os.ReadFile(filepath.Join(root, "go.mod"))
	if err != nil {
		return "1.25"
	}
	for _, line := range strings.Split(string(bs), "\n") {
		if v, ok := strings.CutPrefix(line, "go "); ok {
			return strings.TrimSpace(v)
		}
	}
	return "1.25"
}

// Result is what one probe cost.
type Result struct {
	Probe Probe
	// Funcs is the number of text symbols the compiler emitted into the
	// probe's object: the direct measure of the instantiation cascade.
	Funcs int
	// Leaf is how long the probe alone took to recompile with every
	// dependency already cached. Zero when timing was not requested.
	Leaf time.Duration
}

// Measure builds every probe and counts the functions emitted into each.
// With timing, it then recompiles each probe alone and records how long that
// took; timing is sequential and slow, roughly two seconds per probe while
// the cascade stands.
func (m *Module) Measure(timing bool) ([]Result, error) {
	if _, err := m.run("build", "./..."); err != nil {
		return nil, err
	}

	out, err := m.run("list", "-export", "-f", "{{.ImportPath}} {{.Export}}", "./...")
	if err != nil {
		return nil, err
	}
	exports := map[string]string{}
	for _, line := range strings.Split(strings.TrimSpace(out), "\n") {
		if path, file, ok := strings.Cut(line, " "); ok {
			exports[strings.TrimPrefix(path, "probe/")] = file
		}
	}

	results := make([]Result, 0, len(m.Probes))
	for _, p := range m.Probes {
		file, ok := exports[p.Name]
		if !ok || file == "" {
			return nil, fmt.Errorf("no export data for probe %s", p.Name)
		}
		n, err := countFuncs(file)
		if err != nil {
			return nil, err
		}
		r := Result{Probe: p, Funcs: n}
		if timing {
			r.Leaf, err = m.timeLeaf(p)
			if err != nil {
				return nil, err
			}
		}
		results = append(results, r)
	}
	return results, nil
}

// countFuncs counts the text symbols in a compiled package archive.
func countFuncs(archive string) (int, error) {
	out, err := exec.Command("go", "tool", "nm", archive).Output()
	if err != nil {
		return 0, fmt.Errorf("go tool nm %s: %w", archive, err)
	}
	n := 0
	for _, line := range bytes.Split(out, []byte("\n")) {
		if bytes.Contains(line, []byte(" T ")) {
			n++
		}
	}
	return n, nil
}

// timeLeaf changes the probe's source so the build cache cannot answer, and
// times the rebuild of the probe alone.
func (m *Module) timeLeaf(p Probe) (time.Duration, error) {
	file := filepath.Join(m.Dir, p.Name, "probe.go")
	src := p.Source + "\n// recompile " + strconv.FormatInt(time.Now().UnixNano(), 10) + "\n"
	if err := os.WriteFile(file, []byte(src), 0o644); err != nil {
		return 0, err
	}
	start := time.Now()
	if _, err := m.run("build", "./"+p.Name); err != nil {
		return 0, err
	}
	return time.Since(start), nil
}

// ColdBuild builds the whole probe module with an empty build cache and
// returns the action graph, which carries the per-package timings. Wall is
// how long the build took; CPU is the user and system time it consumed.
func (m *Module) ColdBuild() (graph *ActionGraph, wall, cpu time.Duration, err error) {
	cache, err := os.MkdirTemp("", "compilecost-gocache-")
	if err != nil {
		return nil, 0, 0, err
	}
	defer func() { _ = os.RemoveAll(cache) }()

	agFile := filepath.Join(m.Dir, "actiongraph.json")
	cmd := m.command("build", "-debug-actiongraph="+agFile, "./...")
	cmd.Env = append(cmd.Env, "GOCACHE="+cache)
	start := time.Now()
	if out, err := cmd.CombinedOutput(); err != nil {
		return nil, 0, 0, fmt.Errorf("cold build: %w\n%s", err, out)
	}
	wall = time.Since(start)
	cpu = cmd.ProcessState.UserTime() + cmd.ProcessState.SystemTime()

	graph, err = ReadActionGraph(agFile)
	if graph != nil {
		graph.Main = "probe"
	}
	return graph, wall, cpu, err
}

func (m *Module) command(args ...string) *exec.Cmd {
	cmd := exec.Command("go", args...)
	cmd.Dir = m.Dir
	cmd.Env = append(os.Environ(), "GOFLAGS=-mod=mod")
	cmd.Env = append(cmd.Env, m.Env...)
	if m.Log != nil {
		m.Log("go %s", strings.Join(args, " "))
	}
	return cmd
}

func (m *Module) run(args ...string) (string, error) {
	cmd := m.command(args...)
	var stderr bytes.Buffer
	cmd.Stderr = &stderr
	out, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("go %s: %w\n%s", strings.Join(args, " "), err, stderr.String())
	}
	return string(out), nil
}
