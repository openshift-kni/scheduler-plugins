package main

import (
	"bytes"
	"fmt"
	"os/exec"
	"path/filepath"
	goruntime "runtime"
	"testing"
)

const (
	binariesDir = "bin"
	programName = "verify-commit-message"
)

func TestRunWithOnlyNewlines(t *testing.T) {
	commitMessage := `





`
	_, errBuf, err := runCommand(t, commitMessage)
	// run with empty commit message is unsupported and should fail
	if err == nil {
		t.Fatalf("Unexpectedly succeeded: %v (stderr=%s)\n", err, errBuf)
	}
	exitErr, ok := err.(*exec.ExitError)
	if !ok {
		t.Fatalf("Received unexpected error %T %v", err, err)
	}
	errCode := exitErr.ExitCode()
	if errCode != exitCodeErrorVerificationFailed {
		t.Fatalf("Received unexpected exit code %d", errCode)
	}
}

func TestRunWithEmptyArgument(t *testing.T) {
	_, errBuf, err := runCommand(t, "")
	// run with empty commit message is unsupported and should fail
	if err == nil {
		t.Fatalf("Unexpectedly succeeded: %v (stderr=%s)\n", err, errBuf)
	}
	exitErr, ok := err.(*exec.ExitError)
	if !ok {
		t.Fatalf("Received unexpected error %T %v", err, err)
	}
	errCode := exitErr.ExitCode()
	if errCode != exitCodeErrorVerificationFailed {
		t.Fatalf("Received unexpected exit code %d", errCode)
	}
}

func TestRunWithoutArguments(t *testing.T) {
	_, errBuf, err := runCommand(t) // note no arguments intentionally
	// run without arguments should fail
	if err == nil {
		t.Fatalf("Unexpectedly succeeded: %v (stderr=%s)\n", err, errBuf)
	}
	exitErr, ok := err.(*exec.ExitError)
	if !ok {
		t.Fatalf("Received unexpected error %T %v", err, err)
	}
	errCode := exitErr.ExitCode()
	if errCode != exitCodeErrorWrongArguments {
		t.Fatalf("Received unexpected exit code %d", errCode)
	}
}

// runCommand returns stdout as string, stderr as string, error code
func runCommand(t *testing.T, args ...string) (string, string, error) {
	bin, err := getBinPath()
	if err != nil {
		t.Fatalf("failed to find the binary path: %v", err)
	}
	fmt.Printf("going to use %q\n", bin)
	var errBuf bytes.Buffer
	cmd := exec.Command(bin, args...)
	cmd.Stderr = &errBuf
	out, err := cmd.Output()
	fmt.Printf("tool returned <%s>\n", out)
	return string(out), errBuf.String(), err
}

func getBinPath() (string, error) {
	rootDir, err := getRootPath()
	if err != nil {
		return "", err
	}
	return filepath.Join(rootDir, binariesDir, programName), nil
}

func getRootPath() (string, error) {
	_, file, _, ok := goruntime.Caller(0)
	if !ok {
		return "", fmt.Errorf("cannot retrieve tests directory")
	}
	basedir := filepath.Dir(file)
	return filepath.Abs(filepath.Join(basedir, "..", ".."))
}
