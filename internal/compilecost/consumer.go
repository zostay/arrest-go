package compilecost

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"time"
)

// Timing is the wall and CPU time of one cold command.
type Timing struct {
	Wall, CPU time.Duration
}

// ConsumerResult is what a real consumer paid to build, as it is, and with
// arrest-go replaced by the checkout.
type ConsumerResult struct {
	// Build is a cold `go build ./...`.
	Build, BuildReplaced Timing
	// Test is a cold `go test -run XXX ./...`: compiling the tests without
	// running them.
	Test, TestReplaced Timing
}

// MeasureConsumer builds the module at dir cold, twice: as it is, and with
// arrest-go and arrest-go/gin replaced by the checkout at root. The consumer's
// tree is not touched; the replacement lives in a temporary go.mod passed
// with -modfile.
func MeasureConsumer(dir, root string, log func(string, ...any)) (*ConsumerResult, error) {
	root, err := filepath.Abs(root)
	if err != nil {
		return nil, err
	}
	tmp, err := os.MkdirTemp("", "compilecost-consumer-")
	if err != nil {
		return nil, err
	}
	defer func() { _ = os.RemoveAll(tmp) }()

	modfile := filepath.Join(tmp, "go.mod")
	if err := copyFile(filepath.Join(dir, "go.mod"), modfile); err != nil {
		return nil, err
	}
	if err := copyFile(filepath.Join(dir, "go.sum"), filepath.Join(tmp, "go.sum")); err != nil && !os.IsNotExist(err) {
		return nil, err
	}

	run := func(env []string, args ...string) (Timing, error) {
		cmd := exec.Command("go", args...)
		cmd.Dir = dir
		cmd.Env = append(os.Environ(), env...)
		if log != nil {
			log("go %v", args)
		}
		start := time.Now()
		if out, err := cmd.CombinedOutput(); err != nil {
			return Timing{}, fmt.Errorf("go %v: %w\n%s", args, err, out)
		}
		return Timing{
			Wall: time.Since(start),
			CPU:  cmd.ProcessState.UserTime() + cmd.ProcessState.SystemTime(),
		}, nil
	}

	if _, err := run(nil, "mod", "edit", "-modfile="+modfile,
		"-replace=github.com/zostay/arrest-go="+root,
		"-replace=github.com/zostay/arrest-go/gin="+filepath.Join(root, "gin")); err != nil {
		return nil, err
	}
	if _, err := run(nil, "mod", "tidy", "-modfile="+modfile); err != nil {
		return nil, err
	}

	cold := func(args ...string) (Timing, error) {
		cache, err := os.MkdirTemp("", "compilecost-gocache-")
		if err != nil {
			return Timing{}, err
		}
		defer func() { _ = os.RemoveAll(cache) }()
		return run([]string{"GOCACHE=" + cache}, args...)
	}

	var res ConsumerResult
	if res.Build, err = cold("build", "./..."); err != nil {
		return nil, err
	}
	if res.BuildReplaced, err = cold("build", "-modfile="+modfile, "./..."); err != nil {
		return nil, err
	}
	if res.Test, err = cold("test", "-count=1", "-run", "XXX", "./..."); err != nil {
		return nil, err
	}
	if res.TestReplaced, err = cold("test", "-modfile="+modfile, "-count=1", "-run", "XXX", "./..."); err != nil {
		return nil, err
	}
	return &res, nil
}

func copyFile(from, to string) error {
	bs, err := os.ReadFile(from)
	if err != nil {
		return err
	}
	return os.WriteFile(to, bs, 0o644)
}
