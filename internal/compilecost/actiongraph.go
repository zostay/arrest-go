package compilecost

import (
	"encoding/json"
	"os"
	"regexp"
	"sort"
	"strings"
	"time"
)

// Action is one node of the action graph the go command writes with
// -debug-actiongraph: a build, link or vet of one package.
type Action struct {
	ID        int      `json:"ID"`
	Mode      string   `json:"Mode"`
	Package   string   `json:"Package"`
	Deps      []int    `json:"Deps"`
	TimeStart JSONTime `json:"TimeStart"`
	TimeDone  JSONTime `json:"TimeDone"`
}

// Duration is how long the action itself ran.
func (a Action) Duration() time.Duration {
	if a.TimeStart.IsZero() || a.TimeDone.IsZero() {
		return 0
	}
	return a.TimeDone.Sub(a.TimeStart.Time)
}

// JSONTime tolerates the empty string the go command writes for actions
// that never ran.
type JSONTime struct{ time.Time }

// UnmarshalJSON implements json.Unmarshaler.
func (t *JSONTime) UnmarshalJSON(bs []byte) error {
	var s string
	if err := json.Unmarshal(bs, &s); err != nil {
		return err
	}
	if s == "" {
		t.Time = time.Time{}
		return nil
	}
	parsed, err := time.Parse(time.RFC3339Nano, s)
	if err != nil {
		return err
	}
	t.Time = parsed
	return nil
}

// ActionGraph is a build's action graph.
type ActionGraph struct {
	Actions []Action
	// Main is the module path of the module that was built, whose packages
	// have no dot in their first path element and would otherwise be read
	// as standard library.
	Main string
}

// ReadActionGraph reads the file -debug-actiongraph wrote.
func ReadActionGraph(file string) (*ActionGraph, error) {
	bs, err := os.ReadFile(file)
	if err != nil {
		return nil, err
	}
	var actions []Action
	if err := json.Unmarshal(bs, &actions); err != nil {
		return nil, err
	}
	return &ActionGraph{Actions: actions}, nil
}

// CriticalPath returns the longest chain of dependent actions by duration —
// what a machine with cores to spare still has to wait for — from the leaf
// up to the root, and its total.
func (g *ActionGraph) CriticalPath() ([]Action, time.Duration) {
	n := len(g.Actions)
	total := make([]time.Duration, n)
	next := make([]int, n)
	for i := range next {
		next[i] = -1
	}
	// Actions are numbered so that dependencies precede dependents, but do
	// not rely on it: visit each node after its dependencies.
	done := make([]bool, n)
	var visit func(int)
	visit = func(i int) {
		if done[i] {
			return
		}
		done[i] = true
		var best time.Duration
		for _, d := range g.Actions[i].Deps {
			if d < 0 || d >= n {
				continue
			}
			visit(d)
			if total[d] > best {
				best = total[d]
				next[i] = d
			}
		}
		total[i] = best + g.Actions[i].Duration()
	}
	root := -1
	for i := range g.Actions {
		visit(i)
		if root < 0 || total[i] > total[root] {
			root = i
		}
	}
	if root < 0 {
		return nil, 0
	}

	var path []Action
	for i := root; i >= 0; i = next[i] {
		path = append(path, g.Actions[i])
	}
	return path, total[root]
}

// ModuleTime is the compile time one module's packages consumed.
type ModuleTime struct {
	Module string
	Time   time.Duration
}

var modulePattern = regexp.MustCompile(`^(github\.com/[^/]+/[^/]+|golang\.org/x/[^/]+|[^/]+\.[^/]+/[^/]+(?:/v\d+)?)`)

// ByModule sums build time by module, largest first. Standard library
// packages are reported together as "std".
func (g *ActionGraph) ByModule() []ModuleTime {
	sums := map[string]time.Duration{}
	for _, a := range g.Actions {
		if a.Mode != "build" {
			continue
		}
		mod := "std"
		if m := modulePattern.FindString(a.Package); m != "" {
			mod = m
		} else if g.Main != "" && (a.Package == g.Main || strings.HasPrefix(a.Package, g.Main+"/")) {
			mod = g.Main
		}
		sums[mod] += a.Duration()
	}
	out := make([]ModuleTime, 0, len(sums))
	for mod, d := range sums {
		out = append(out, ModuleTime{Module: mod, Time: d})
	}
	sort.Slice(out, func(i, j int) bool { return out[i].Time > out[j].Time })
	return out
}

// Total is the build time of every action added together: what a machine
// with one core would wait for.
func (g *ActionGraph) Total() time.Duration {
	var total time.Duration
	for _, a := range g.Actions {
		total += a.Duration()
	}
	return total
}
