// Command compilecost reports what it costs a consumer to compile against
// this checkout of arrest-go. See internal/compilecost for what it measures
// and why; scripts/compile-cost and `make compile-cost` run it.
//
// By default it builds one probe package per public type of arrest and
// arrest/gin, plus a small realistic consumer of each, and prints how many
// functions the compiler emitted into each probe and how long each takes to
// recompile on its own. With -cold it also builds the probe module from an
// empty cache and prints the critical path and the time by module. With
// -consumer it builds a real module cold, as it is and with this checkout
// replacing the released arrest-go, and prints both.
package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/zostay/arrest-go/internal/compilecost"
)

func main() {
	root := flag.String("root", "", "arrest-go checkout to measure (default: the main module of the current directory)")
	timing := flag.Bool("time", true, "time each probe's leaf recompile (slow while the cascade stands)")
	cold := flag.Bool("cold", false, "also build the probe module from an empty cache and report the critical path")
	consumer := flag.String("consumer", "", "path to a real consumer module to build cold, as is and with this checkout")
	keep := flag.Bool("keep", false, "keep the probe module directory and print its path")
	asJSON := flag.Bool("json", false, "print results as JSON")
	verbose := flag.Bool("v", false, "print the go commands as they run")
	flag.Parse()

	if err := run(*root, *timing, *cold, *consumer, *keep, *asJSON, *verbose); err != nil {
		fmt.Fprintln(os.Stderr, "compilecost:", err)
		os.Exit(1)
	}
}

type report struct {
	Probes   []compilecost.Result        `json:"probes"`
	Cold     *coldReport                 `json:"cold,omitempty"`
	Consumer *compilecost.ConsumerResult `json:"consumer,omitempty"`
}

type coldReport struct {
	Wall, CPU, Total time.Duration
	CriticalPath     []compilecost.Action
	CriticalTotal    time.Duration
	ByModule         []compilecost.ModuleTime
}

func run(root string, timing, cold bool, consumer string, keep, asJSON, verbose bool) error {
	if root == "" {
		out, err := exec.Command("go", "list", "-m", "-f", "{{.Dir}}", "github.com/zostay/arrest-go").Output()
		if err != nil {
			return fmt.Errorf("finding the arrest-go module (pass -root): %w", err)
		}
		root = strings.TrimSpace(string(out))
	}

	var log func(string, ...any)
	if verbose {
		log = func(format string, args ...any) { fmt.Fprintf(os.Stderr, format+"\n", args...) }
	}

	probes, err := compilecost.Probes(root)
	if err != nil {
		return err
	}
	dir, err := os.MkdirTemp("", "compilecost-")
	if err != nil {
		return err
	}
	if keep {
		fmt.Fprintln(os.Stderr, "probe module:", dir)
	} else {
		defer func() { _ = os.RemoveAll(dir) }()
	}

	mod, err := compilecost.WriteModule(dir, root, probes)
	if err != nil {
		return err
	}
	mod.Log = log

	var rep report
	if rep.Probes, err = mod.Measure(timing); err != nil {
		return err
	}

	if cold {
		graph, wall, cpu, err := mod.ColdBuild()
		if err != nil {
			return err
		}
		path, total := graph.CriticalPath()
		rep.Cold = &coldReport{
			Wall: wall, CPU: cpu, Total: graph.Total(),
			CriticalPath: path, CriticalTotal: total,
			ByModule: graph.ByModule(),
		}
	}

	if consumer != "" {
		if rep.Consumer, err = compilecost.MeasureConsumer(consumer, root, log); err != nil {
			return err
		}
	}

	if asJSON {
		enc := json.NewEncoder(os.Stdout)
		enc.SetIndent("", "  ")
		return enc.Encode(rep)
	}
	print(rep, timing, consumer)
	return nil
}

func print(rep report, timing bool, consumer string) {
	fmt.Println("Functions emitted into each probe package (the instantiation cascade):")
	fmt.Println()
	for _, r := range rep.Probes {
		what := r.Probe.Type
		switch r.Probe.Kind {
		case "use":
			what = "a small consumer of " + strings.TrimPrefix(r.Probe.Name, "use_")
		case "ref":
			what = "reference: " + strings.TrimPrefix(r.Probe.Name, "ref_")
		}
		if timing {
			fmt.Printf("  %-28s %7d funcs  %6.2fs  %s\n", r.Probe.Name, r.Funcs, r.Leaf.Seconds(), what)
		} else {
			fmt.Printf("  %-28s %7d funcs  %s\n", r.Probe.Name, r.Funcs, what)
		}
	}

	if rep.Cold != nil {
		c := rep.Cold
		fmt.Println()
		fmt.Printf("Cold build of the probe module: %.1fs wall, %.1fs CPU, %.1fs of compile actions\n",
			c.Wall.Seconds(), c.CPU.Seconds(), c.Total.Seconds())
		fmt.Printf("\nCritical path (%.1fs):\n", c.CriticalTotal.Seconds())
		for _, a := range c.CriticalPath {
			if a.Duration() < 200*time.Millisecond {
				continue
			}
			fmt.Printf("  %6.2fs  %-6s %s\n", a.Duration().Seconds(), a.Mode, a.Package)
		}
		fmt.Println("\nCompile time by module:")
		for i, m := range c.ByModule {
			if i >= 15 {
				break
			}
			fmt.Printf("  %6.2fs  %s\n", m.Time.Seconds(), m.Module)
		}
	}

	if rep.Consumer != nil {
		r := rep.Consumer
		fmt.Printf("\nCold builds of %s, as released and with this checkout:\n\n", filepath.Clean(consumer))
		fmt.Printf("  %-24s %8s %8s   %8s %8s\n", "", "wall", "CPU", "wall", "CPU")
		fmt.Printf("  %-24s %7.1fs %7.1fs   %7.1fs %7.1fs\n", "go build ./...",
			r.Build.Wall.Seconds(), r.Build.CPU.Seconds(), r.BuildReplaced.Wall.Seconds(), r.BuildReplaced.CPU.Seconds())
		fmt.Printf("  %-24s %7.1fs %7.1fs   %7.1fs %7.1fs\n", "go test -run XXX ./...",
			r.Test.Wall.Seconds(), r.Test.CPU.Seconds(), r.TestReplaced.Wall.Seconds(), r.TestReplaced.CPU.Seconds())
	}
}
