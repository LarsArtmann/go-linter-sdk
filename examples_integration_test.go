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
