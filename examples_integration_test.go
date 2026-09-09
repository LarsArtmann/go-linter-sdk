package linter_test

import (
	"bytes"
	"context"
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

// buildExample compiles an example binary and returns its path.
func buildExample(t *testing.T, target string) string {
	t.Helper()

	bin := filepath.Join(t.TempDir(), filepath.Base(target))

	cmd := exec.CommandContext(context.Background(), "go", "build", "-o", bin, target) //nolint:gosec

	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("go build %s: %v\n%s", target, err, out)
	}

	return bin
}

// runExample executes the binary against dir and returns stdout, stderr, and exit code.
func runExample(t *testing.T, bin, dir string) (string, string, int) {
	t.Helper()

	cmd := exec.CommandContext(context.Background(), bin, dir) //nolint:gosec

	stdoutBuf := &bytes.Buffer{}
	stderrBuf := &bytes.Buffer{}
	cmd.Stdout = stdoutBuf
	cmd.Stderr = stderrBuf

	err := cmd.Run()

	exitCode := 0

	if exitErr, ok := errors.AsType[*exec.ExitError](err); ok {
		exitCode = exitErr.ExitCode()
	} else if err != nil {
		t.Fatalf("failed to run example binary: %v", err)
	}

	return stdoutBuf.String(), stderrBuf.String(), exitCode
}

// Minimal-linter tests.
func TestExampleMinimalLinter_CleanDir(t *testing.T) {
	t.Parallel()

	bin := buildExample(t, "./examples/minimal-linter")

	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "README.md"), []byte("# test"), 0o600); err != nil {
		t.Fatal(err)
	}

	stdout, _, code := runExample(t, bin, dir)
	if code != 0 {
		t.Errorf("expected exit 0, got %d", code)
	}

	if stdout != "" {
		t.Errorf("expected empty stdout, got %q", stdout)
	}
}

func TestExampleMinimalLinter_MissingReadme(t *testing.T) {
	t.Parallel()

	bin := buildExample(t, "./examples/minimal-linter")

	dir := t.TempDir()

	stdout, _, code := runExample(t, bin, dir)
	if code != 1 {
		t.Errorf("expected exit 1, got %d", code)
	}

	if !strings.Contains(stdout, "README.md is missing") {
		t.Errorf("expected finding in stdout, got %q", stdout)
	}
}

// Naked-return-guard tests.
func TestExampleNakedReturnGuard_CleanCode(t *testing.T) {
	t.Parallel()

	bin := buildExample(t, "./examples/naked-return-guard")

	dir := t.TempDir()

	src := "package a\n\nfunc ok() (n int) {\n\tn = 1\n\treturn n\n}\n"
	if err := os.WriteFile(filepath.Join(dir, "a.go"), []byte(src), 0o600); err != nil {
		t.Fatal(err)
	}

	stdout, _, code := runExample(t, bin, dir)
	if code != 0 {
		t.Errorf("expected exit 0, got %d", code)
	}

	if stdout != "" {
		t.Errorf("expected empty stdout, got %q", stdout)
	}
}

func TestExampleNakedReturnGuard_NakedReturnInLongFunction(t *testing.T) {
	t.Parallel()

	bin := buildExample(t, "./examples/naked-return-guard")

	dir := t.TempDir()
	// 12-line body (bodyLen = end - start = 12 > 10) with a naked return.
	src := "package a\n\nfunc bad() (n int, err error) {\n\tn = 1\n\tx := 2\n\ty := 3\n\tz := 4\n\ta := 5\n\tb := 6\n\tc := 7\n\td := 8\n\te := 9\n\tif n > 0 {\n\t\treturn\n\t}\n\treturn n, nil\n}\n"
	if err := os.WriteFile(filepath.Join(dir, "a.go"), []byte(src), 0o600); err != nil {
		t.Fatal(err)
	}

	stdout, _, code := runExample(t, bin, dir)
	if code != 1 {
		t.Errorf("expected exit 1, got %d", code)
	}

	if !strings.Contains(stdout, "naked return in bad() (14 lines)") {
		t.Errorf("expected finding in stdout, got %q", stdout)
	}
}

func TestExampleNakedReturnGuard_SkipsTestFiles(t *testing.T) {
	t.Parallel()

	bin := buildExample(t, "./examples/naked-return-guard")

	dir := t.TempDir()

	src := "package a\n\nfunc bad() (n int) {\n\treturn\n}\n"
	if err := os.WriteFile(filepath.Join(dir, "a_test.go"), []byte(src), 0o600); err != nil {
		t.Fatal(err)
	}

	stdout, _, code := runExample(t, bin, dir)
	if code != 0 {
		t.Errorf("expected exit 0 (test files skipped), got %d", code)
	}

	if stdout != "" {
		t.Errorf("expected empty stdout, got %q", stdout)
	}
}

// Oops-redundant-guard tests.
func TestExampleOopsRedundantGuard_CleanCode(t *testing.T) {
	t.Parallel()

	bin := buildExample(t, "./examples/oops-redundant-guard")

	dir := t.TempDir()

	src := "package a\n\nimport \"github.com/samber/oops\"\n\nfunc good(err error) error {\n\tif err != nil {\n\t\treturn oops.Wrap(err)\n\t}\n\n\treturn doWork()\n}\n\nfunc doWork() error { return nil }\n"
	if err := os.WriteFile(filepath.Join(dir, "a.go"), []byte(src), 0o600); err != nil {
		t.Fatal(err)
	}

	stdout, _, code := runExample(t, bin, dir)
	if code != 0 {
		t.Errorf("expected exit 0, got %d", code)
	}

	if stdout != "" {
		t.Errorf("expected empty stdout, got %q", stdout)
	}
}

func TestExampleOopsRedundantGuard_DetectsRedundantGuard(t *testing.T) {
	t.Parallel()

	bin := buildExample(t, "./examples/oops-redundant-guard")

	dir := t.TempDir()

	src := "package a\n\nimport \"github.com/samber/oops\"\n\nfunc bad(err error) error {\n\tif err != nil {\n\t\treturn oops.Wrap(err)\n\t}\n\n\treturn nil\n}\n"
	if err := os.WriteFile(filepath.Join(dir, "a.go"), []byte(src), 0o600); err != nil {
		t.Fatal(err)
	}

	stdout, _, code := runExample(t, bin, dir)
	if code != 1 {
		t.Errorf("expected exit 1, got %d", code)
	}

	if !strings.Contains(stdout, "Redundant nil guard before oops.Wrap(err)") {
		t.Errorf("expected finding in stdout, got %q", stdout)
	}
}

func TestExampleOopsRedundantGuard_GuardWithElseNotFlagged(t *testing.T) {
	t.Parallel()

	bin := buildExample(t, "./examples/oops-redundant-guard")

	dir := t.TempDir()

	src := "package a\n\nimport (\n\t\"errors\"\n\t\"github.com/samber/oops\"\n)\n\nfunc branches(err error) error {\n\tif err != nil {\n\t\treturn oops.Wrap(err)\n\t} else {\n\t\treturn errors.New(\"x\")\n\t}\n}\n"
	if err := os.WriteFile(filepath.Join(dir, "a.go"), []byte(src), 0o600); err != nil {
		t.Fatal(err)
	}

	stdout, _, code := runExample(t, bin, dir)
	if code != 0 {
		t.Errorf("expected exit 0 (if/else is not the pattern), got %d", code)
	}

	if stdout != "" {
		t.Errorf("expected empty stdout, got %q", stdout)
	}
}

// No-go-mod tests.
func TestExampleNoGoMod_CleanDir(t *testing.T) {
	t.Parallel()

	bin := buildExample(t, "./examples/no-go-mod")

	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "go.mod"), []byte("module test\n\ngo 1.26\n"), 0o600); err != nil {
		t.Fatal(err)
	}

	stdout, _, code := runExample(t, bin, dir)
	if code != 0 {
		t.Errorf("expected exit 0, got %d", code)
	}

	if stdout != "" {
		t.Errorf("expected empty stdout, got %q", stdout)
	}
}

func TestExampleNoGoMod_MissingGoMod(t *testing.T) {
	t.Parallel()

	bin := buildExample(t, "./examples/no-go-mod")

	dir := t.TempDir()

	stdout, _, code := runExample(t, bin, dir)
	if code != 1 {
		t.Errorf("expected exit 1, got %d", code)
	}

	if !strings.Contains(stdout, "Missing go.mod") {
		t.Errorf("expected finding in stdout, got %q", stdout)
	}
}
