package main

import (
	"bufio"
	"encoding/json"
	"errors"
	"log"
	"os"
	"os/exec"
	"strings"
)

const (
	exitCodeSuccess                 = 0
	exitCodeErrorWrongArguments     = 1
	exitCodeErrorVerificationFailed = 2
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
	// originName is the git remote that points to the clone of this repo
	originName string
	// upstreamName is the git remote that points kubernetes-sigs/scheduler-plugins repo
	upstreamName string
}

var conf = config{
	originName:   defaultOriginName,
	upstreamName: defaultUpstreamName,
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
		remoteName := conf.originName
		if upstream {
			remoteName = conf.upstreamName
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
		return err
	}
	outStr := string(out)

	if !strings.Contains(outStr, remoteName+"/"+referenceBranchName) {
		return errWrongCherryPickReference
	}
	return nil
}

func main() {
	if len(os.Args) != 2 {
		programName := os.Args[0]
		log.Printf("usage: %s wrong number of arguments expects: %s <commit-message>", programName, programName)
		os.Exit(exitCodeErrorWrongArguments)
	}

	data, err := os.ReadFile("config.json") // TODO make it configurable
	if err != nil {
		log.Printf("error reading config.json: %v", err)
		os.Exit(exitCodeErrorWrongArguments)
	}
	err = json.Unmarshal(data, &conf)
	if err != nil {
		log.Printf("error parsing config.json: %v", err)
		os.Exit(exitCodeErrorWrongArguments)
	}

	err = verifyCommitMessage(os.Args[1])
	if err != nil {
		log.Printf("verification failed: %v", err)
		os.Exit(exitCodeErrorVerificationFailed)
	}

	os.Exit(exitCodeSuccess) // all good! redundant but let's be explicit about our success
}
