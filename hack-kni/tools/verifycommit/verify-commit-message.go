package main

import (
	"bufio"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"log"
	"os"
	"os/exec"
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
}

var conf = config{
	OriginName:   defaultOriginName,
	UpstreamName: defaultUpstreamName,
}

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

func (cm commitMessage) isKonflux() bool {
	for idx := cm.numLines() - 1; idx > 0; idx-- {
		line := cm.lines[idx] // shortcut
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
func verifyCommitMessage(commitMessage string) error {
	cm := newCommitMessageFromString(commitMessage)

	if cm.isKonflux() {
		return nil
	}
	return verifyHumanCommitMessage(cm)
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
	flag.Parse()

	if flag.NArg() != 1 {
		programName := os.Args[0]
		log.Printf("usage: %s [-f config-file] <commit-message>", programName)
		os.Exit(exitCodeErrorWrongArguments)
	}

	data, err := os.ReadFile(*configFileName)
	if err != nil {
		fmt.Printf("error reading %s: %v", *configFileName, err)
		os.Exit(exitCodeErrorProcessingFile)
	}
	err = json.Unmarshal(data, &conf) // keep it flexible
	if err != nil {
		log.Printf("error parsing %s: %v", *configFileName, err)
		os.Exit(exitCodeErrorProcessingFile)
	}

	err = verifyCommitMessage(flag.Arg(0))
	if err != nil {
		log.Printf("verification failed: %v", err)
		os.Exit(exitCodeErrorVerificationFailed)
	}

	os.Exit(exitCodeSuccess) // all good! redundant but let's be explicit about our success
}
