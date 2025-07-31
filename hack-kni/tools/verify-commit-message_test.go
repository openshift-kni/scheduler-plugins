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

func TestRunWithDifferentCommits(t *testing.T) {
	testcases := []struct {
		description      string
		commitMessage    string
		expectedExitCode int
	}{
		{
			description: "valid KNI commit message",
			commitMessage: `[KNI] doc about cascading **KNI SPECIFIC** cherry-picks

We can avoid long chains of cherry-picked from commit
references JUST AND ONLY for KNI-specific changes.

Signed-off-by: Francesco Romani <fromani@redhat.com>`,
			expectedExitCode: 0,
		},
	}

	for _, tc := range testcases {
		t.Run(tc.description, func(t *testing.T) {
			_, _, err := runCommand(t, tc.commitMessage)
			if tc.expectedExitCode == 0 && err == nil {
				// everything is as expected
				return
			}

			if tc.expectedExitCode == 0 && err != nil {
				t.Fatalf("Received unexpected error %v", err)
			}

			exiterr, ok := err.(*exec.ExitError)
			if !ok {
				t.Fatalf("Received unexpected error %T %v", err, err)
			}

			errCode := exiterr.ExitCode()
			if errCode != tc.expectedExitCode {
				t.Fatalf("Received unexpected exit code: expected %d got %d", tc.expectedExitCode, errCode)
			}
		})
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

func TestVerifyCommitMessage(t *testing.T) {
	testcases := []struct {
		description string
		commitMsg   string
		expectedErr error
	}{
		{
			description: "only KNI in local fork",
			commitMsg: `[KNI] hack-kni: skip commit verification for konflux commits
    
    Skip validating konflux commits structure. Usually konflux bot commits
    signed off by either "red-hat-konflux" or "red-hat-konflux[bot]".
    
    Signed-off-by: Shereen Haj <shajmakh@redhat.com>`,
		},
	}
	for _, tc := range testcases {
		t.Run(tc.description, func(t *testing.T) {
			got := verifyCommitMessage(tc.commitMsg)

			if got == nil && tc.expectedErr == nil {
				return
			}

			if tc.expectedErr != nil && got == nil {
				t.Fatalf("expected error %v but recieved nil", tc.expectedErr)
			}

			if tc.expectedErr == nil && got != nil {
				t.Fatalf("unexpected error %v", got)
			}

			if got.Error() != tc.expectedErr.Error() {
				t.Fatalf("mismatching error strings: expected %s, got %s", tc.expectedErr.Error(), got.Error())
			}
		})
	}
}

// runCommand returns stdout as string, stderr as string, error value
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
