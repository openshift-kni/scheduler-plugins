package main

import (
	"bufio"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

const (
	exitCodeSuccess                 = 0
	exitCodeErrorWrongArguments     = 1
	exitCodeErrorProcessingFile     = 2
	exitCodeErrorVerificationFailed = 3
)

const (
	exitCodeGitMalformedObject = 129
)

const (
	tagKNI      = "[KNI]"
	tagUpstream = "[upstream]"

	cherryPickLinePrefix = "(cherry picked from commit "
	cherryPickLineSuffix = ")" // yes that simple

	signedOffByPrefix = "Signed-off-by: "

	konfluxUsername = "red-hat-konflux"

	defaultOriginName   = "origin"
	defaultUpstreamName = "upstream"
	referenceBranchName = "master"

	resyncBranchPrefix = "resync-"

	headBranchName = "HEAD"
)

var (
	errEmptyCommitMessage         = errors.New("empty commit message")
	errMissingTagKNI              = errors.New("missing tag: " + tagKNI)
	errMissingCherryPickReference = errors.New("missing cherry pick reference")
	errWrongCherryPickReference   = errors.New("wrong cherry pick reference")
)

type commitMessage struct {
	lines []string
}

type config struct {
	// OriginName is the git remote that points to the clone of this repo
	OriginName string `json:"originName"`
	// UpstreamName is the git remote that points kubernetes-sigs/scheduler-plugins repo
	UpstreamName string `json:"upstreamName"`
	// TriggerBranch is the branch name that triggers the verification
	TriggerBranch string `json:"triggerBranch"`
}

var conf = config{
	OriginName:   defaultOriginName,
	UpstreamName: defaultUpstreamName,
}

var sourceBranch *string

func newCommitMessageFromString(text string) commitMessage {
	var cm commitMessage
	scanner := bufio.NewScanner(strings.NewReader(text))
	for scanner.Scan() {
		cm.lines = append(cm.lines, scanner.Text())
	}
	log.Printf("commit message has %d lines", cm.numLines())
	return cm
}

func (cm commitMessage) numLines() int {
	return len(cm.lines)
}

func (cm commitMessage) isEmpty() bool {
	return cm.numLines() == 0
}

func (cm commitMessage) summary() string {
	return cm.lines[0]
}

func (cm commitMessage) isKNISpecific() bool {
	return strings.Contains(cm.summary(), tagKNI)
}

func (cm commitMessage) isUpstream() bool {
	return strings.Contains(cm.summary(), tagUpstream)
}

func (cm commitMessage) isKonflux() bool {
	for idx := cm.numLines() - 1; idx > 0; idx-- {
		line := cm.lines[idx] // shortcut
		line = strings.TrimSpace(line)
		signedOff, ok := strings.CutPrefix(line, signedOffByPrefix)
		if !ok {
			continue
		}
		if strings.HasPrefix(signedOff, konfluxUsername) {
			return true
		}
	}
	return false // nothing found
}

func isResyncBranch(branch string) bool {
	return strings.HasPrefix(branch, resyncBranchPrefix)
}

// cherryPickOrigin returns the commit hash this commit was cherry-picked
// from if this commit has cherry-pick reference; otherwise returns empty string.
func (cm commitMessage) cherryPickOrigin() string {
	for idx := cm.numLines() - 1; idx > 0; idx-- {
		line := cm.lines[idx] // shortcut
		cmHash, ok := strings.CutPrefix(line, cherryPickLinePrefix)
		if !ok { // we don't have the prefix, so we don't care
			continue
		}
		cmHash, ok = strings.CutSuffix(cmHash, cherryPickLineSuffix)
		if !ok { // we don't have the suffix, so we don't care
			continue
		}
		return cmHash
	}
	return "" // nothing found
}

func validateCommitMessage(commitMessage string) error {
	cm := newCommitMessageFromString(commitMessage)

	if cm.isKonflux() {
		return nil
	}
	return verifyHumanCommitMessage(cm)
}

func getCommitMessageByHash(commitHash string) (string, error) {
	cmd := exec.Command("git", "show", "--format=%B", commitHash)
	out, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("failed to get commit message for %s: %v", commitHash, err)
	}
	return string(out), nil
}

func verifyLastNCommits(numCommits int) error {
	log.Printf("considering %d commits in PR whose head is %s:\n", numCommits, *sourceBranch)

	// Get the list of commits
	cmd := exec.Command("git", "log", *sourceBranch, "--oneline", "--no-merges", "-n", fmt.Sprintf("%d", numCommits))
	out, err := cmd.Output()
	if err != nil {
		return fmt.Errorf("failed to get commit list: %v", err)
	}
	log.Println(string(out))

	commitLines := strings.Split(strings.TrimSpace(string(out)), "\n")
	for _, line := range commitLines {
		log.Printf("examining %q", line)

		fields := strings.Fields(line)
		if len(fields) == 0 {
			continue
		}
		commitHash := fields[0]

		commit, err := getCommitMessageByHash(commitHash)
		if err != nil {
			return err
		}
		log.Printf("verifying message:\n%q\n", commit)
		if err = validateCommitMessage(commit); err != nil {
			return err
		}
	}

	return nil
}

func verifyHumanCommitMessage(cm commitMessage) error {
	if cm.isEmpty() {
		return errEmptyCommitMessage
	}

	if !cm.isKNISpecific() {
		return errMissingTagKNI
	}

	cpOrigin := cm.cherryPickOrigin()
	upstream := cm.isUpstream()

	if cpOrigin == "" {
		if upstream {
			return errMissingCherryPickReference
		}
	}

	if cpOrigin != "" {
		remoteName := conf.OriginName
		if upstream {
			remoteName = conf.UpstreamName
		}

		err := isCommitInBranch(remoteName, cpOrigin)
		if err != nil {
			return err
		}
	}

	return nil
}

func isCommitInBranch(remoteName, cpOrigin string) error {
	cmd := exec.Command("git", "branch", "-r", "--contains", cpOrigin)
	out, err := cmd.Output()
	if err != nil {
		if isMalformedObjectErr(err, cpOrigin) {
			return errWrongCherryPickReference
		}
		return err
	}

	outStr := string(out)

	if !strings.Contains(outStr, remoteName+"/"+referenceBranchName) {
		return errWrongCherryPickReference
	}
	return nil
}

func isMalformedObjectErr(err error, objHash string) bool {
	if exitErr, ok := err.(*exec.ExitError); ok {
		if string(exitErr.Stderr) == fmt.Sprintf("error: malformed object name %s\n", objHash) && exitErr.ExitCode() == exitCodeGitMalformedObject {
			return true
		}
	}
	return false
}

func main() {
	var configFileName = flag.String("f", "config.json", "config file path")
	var numCommits = flag.Int("n", 0, "number of last commits to verify (defaults to 0)")
	sourceBranch = flag.String("b", "", "source branch name")
	flag.Parse()

	if *numCommits < 0 {
		programName := os.Args[0]
		log.Printf("usage: %s -b <branch-name> -n <number-of-commits> [-f config-file]", programName)
		log.Printf("  -b: source branch name (if empty, the tool will use the current branch)")
		log.Printf("  -n: number of last commits to verify (if not provided it defaults to 0)")
		log.Printf("  -f: config file path (optional, remote defaults are origin and upstream)")
		os.Exit(exitCodeErrorWrongArguments)
	}

	if *numCommits == 0 {
		log.Printf("number of commits to verify is 0, skipping verification")
		os.Exit(exitCodeSuccess)
	}

	if strings.TrimSpace(*sourceBranch) == "" {
		*sourceBranch = headBranchName
		log.Printf("using branch: %s", *sourceBranch)
	}

	if isResyncBranch(*sourceBranch) {
		log.Printf("WARN: resync branch no commit enforcement will be triggered\n")
		os.Exit(exitCodeSuccess)
	}

	err := processConfigFile(*configFileName)
	if err != nil {
		log.Printf("error processing config file: %v", err)
		os.Exit(exitCodeErrorProcessingFile)
	}

	err = verifyLastNCommits(*numCommits)
	if err != nil {
		log.Printf("verification failed: %v", err)
		os.Exit(exitCodeErrorVerificationFailed)
	}

	os.Exit(exitCodeSuccess) // all good! redundant but let's be explicit about our success
}

func processConfigFile(filePath string) error {
	// Use os.OpenRoot to prevent directory traversal attacks
	// This ensures file access is scoped to the validated absolute path
	root, err := os.OpenRoot(filepath.Dir(filePath))
	if err != nil {
		return fmt.Errorf("error opening root directory for %s: %v", filePath, err)
	}
	defer func() {
		_ = root.Close()
	}()

	file, err := root.Open(filepath.Base(filePath))
	if err != nil {
		return fmt.Errorf("error opening file %s: %v", filePath, err)
	}
	defer func() {
		_ = file.Close()
	}()

	fileContent, err := io.ReadAll(file)
	if err != nil {
		return fmt.Errorf("error reading content from %s: %v", filePath, err)
	}

	err = json.Unmarshal(fileContent, &conf) // keep it flexible
	if err != nil {
		return fmt.Errorf("error parsing %s: %v", filePath, err)
	}
	return nil
}
