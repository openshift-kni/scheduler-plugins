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

func TestRunCommand(t *testing.T) {
	validCommit := `[KNI] doc about cascading **KNI SPECIFIC** cherry-picks

We can avoid long chains of cherry-picked from commit
references JUST AND ONLY for KNI-specific changes.

Signed-off-by: Francesco Romani <fromani@redhat.com>`

	testcases := []struct {
		description      string
		args             []string
		expectedExitCode int
	}{
		{
			description:      "pattern 1: program-name <commit-message>",
			args:             []string{validCommit},
			expectedExitCode: exitCodeSuccess,
		},
		{
			description:      "pattern 2: program-name -f <config-file> <commit-message>",
			args:             []string{"-f", "config.json", validCommit},
			expectedExitCode: exitCodeSuccess,
		},
		{
			description:      "pattern 3: missing config data should default to default values",
			args:             []string{"-f", "test-config-missing-origin.json", validCommit},
			expectedExitCode: exitCodeSuccess,
		},
		{
			description:      "invalid: too many non-flag arguments",
			args:             []string{validCommit, "extra-arg"},
			expectedExitCode: exitCodeErrorWrongArguments,
		},
		{
			description:      "invalid: no commit message",
			args:             []string{"-f", "config.json"},
			expectedExitCode: exitCodeErrorWrongArguments,
		},
		{
			description:      "invalid: no arguments at all",
			args:             []string{},
			expectedExitCode: exitCodeErrorWrongArguments,
		},
		{
			description: "invalid: empty lines commit message",
			args: []string{`


`},
			expectedExitCode: exitCodeErrorVerificationFailed,
		},
		{
			description:      "invalid: empty commit message",
			args:             []string{""},
			expectedExitCode: exitCodeErrorVerificationFailed,
		},
		{
			description:      "invalid: not-found config file",
			args:             []string{"-f", "not-found.json", validCommit},
			expectedExitCode: exitCodeErrorProcessingFile,
		},
	}

	for _, tc := range testcases {
		t.Run(tc.description, func(t *testing.T) {
			_, errBuf, err := runCommand(t, tc.args...)

			if tc.expectedExitCode == exitCodeSuccess && err == nil {
				// everything is as expected
				return
			}

			if tc.expectedExitCode == exitCodeSuccess && err != nil {
				t.Fatalf("Received unexpected error %v (stderr=%s)", err, errBuf)
			}

			exiterr, ok := err.(*exec.ExitError)
			if !ok {
				t.Fatalf("Received unexpected error type %T: %v", err, err)
			}

			errCode := exiterr.ExitCode()
			if errCode != tc.expectedExitCode {
				t.Fatalf("Received unexpected exit code: expected %d got %d", tc.expectedExitCode, errCode)
			}
		})
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
		{
			description: "KNI & upstream tags",
			commitMsg: `[KNI][upstream] nrt: test: ensure generation is updated correctly
Add a unit test mainly for GetCachedNRTCopy() to verify the returned
generation. To help with the verification, make FlushNodes not only
report about the new generation (which only updated there) but also
return it so we'd be able to compare the values.

Signed-off-by: Shereen Haj <shajmakh@redhat.com>
(cherry picked from commit ffe2ce2)`,
		},
		{
			// this test depends on the local setup. The expected setup should be that the forked project remote
			// name is called "origin" and the main project from which this project is forked called "upstream"
			description: "KNI and local cherrypick",
			commitMsg: `[KNI][release-4.18] ci: ghactions: ensure golang version in vendor check
make sure we run the vendor check in a controlled environment,
and also make sure to emit the golang version we use.

Signed-off-by: Francesco Romani <fromani@redhat.com>
(cherry picked from commit 2f4974a)`,
		},
		{
			description: "Konflux signed - github email",
			commitMsg: `Update Konflux references to 252e5c9
Signed-off-by: red-hat-konflux <126015336+red-hat-konflux[bot]@users.noreply.github.com>`,
		},
		{
			description: "Konflux signed - ci email",
			commitMsg: `Update Konflux references
Signed-off-by: red-hat-konflux <konflux@no-reply.konflux-ci.dev>`,
		},
		{
			description: "Negative - no KNI tag and not konflux signed",
			commitMsg: `nrt: test: ensure generation is updated correctly
Add a unit test

    Signed-off-by: Shereen Haj <shajmakh@redhat.com>`,
			expectedErr: errMissingTagKNI,
		},
		{
			description: "Negative - no KNI tag but with upstream tag",
			commitMsg: `[upstream] nrt: test: ensure generation is updated correctly
Add a unit test 

Signed-off-by: Shereen Haj <shajmakh@redhat.com>
(cherry picked from commit ffe2ce2)`,
			expectedErr: errMissingTagKNI,
		},
		{
			description: "Negative - KNI & upstream tags without cherrypick hash",
			commitMsg: `[KNI][upstream] nrt: test: ensure generation is updated correctly
Add a unit test

Signed-off-by: Shereen Haj <shajmakh@redhat.com>`,
			expectedErr: errMissingCherryPickReference,
		},
		{
			description: "Negative - Konflux generated without signature",
			commitMsg: `Update Konflux references to 252e5c9


`,
			expectedErr: errMissingTagKNI,
		},
		{
			description: "Negative - invalid cherry-pick reference",
			commitMsg: `[KNI][upstream] nrt: test: ensure generation is updated correctly
Add a unit test 

Signed-off-by: Shereen Haj <shajmakh@redhat.com>
(cherry picked from commit 123a)`,
			expectedErr: errWrongCherryPickReference,
		},
		{
			description: "Negative - empty cherry-pick reference",
			commitMsg: `[KNI][upstream] nrt: test: ensure generation is updated correctly
Add a unit test 

Signed-off-by: Shereen Haj <shajmakh@redhat.com>
(cherry picked from commit )`,
			expectedErr: errMissingCherryPickReference,
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
	return filepath.Abs(filepath.Join(basedir, "..", "..", ".."))
}
