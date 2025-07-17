package main

import (
	"bufio"
	"errors"
	"log"
	"os"
	"strings"
)

const (
	exitCodeSuccess                 = 0
	exitCodeErrorWrongArguments     = 1
	exitCodeErrorVerificationFailed = 2
)

const (
	tagKNI = "[KNI]"

	cherryPickLinePrefix = "(cherry picked from commit "
	cherryPickLineSuffix = ")" // yes that simple
)

var (
	errEmptyCommitMessage = errors.New("empty commit message")
	errMissingTagKNI      = errors.New("missing tag: " + tagKNI)
)

type commitMessage struct {
	lines []string
}

func newCommitMessageFromString(text string) commitMessage {
	var cm commitMessage
	scanner := bufio.NewScanner(strings.NewReader(text))
	for scanner.Scan() {
		cm.lines = append(cm.lines, scanner.Text())
	}
	log.Printf("commit message has %d lines", cm.NumLines())
	return cm
}

func (cm commitMessage) NumLines() int {
	return len(cm.lines)
}

func (cm commitMessage) IsEmpty() bool {
	return cm.NumLines() == 0
}

func (cm commitMessage) Summary() string {
	return cm.lines[0]
}

func (cm commitMessage) IsKNISpecific() bool {
	return strings.Contains(cm.Summary(), tagKNI)
}

// CherryPickOrigin returns the commit hash this commit was cherry-picked
// from if this commit has cherry-pick reference; otherwise returns empty string.
func (cm commitMessage) CherryPickOrigin() string {
	for idx := cm.NumLines() - 1; idx > 0; idx-- {
		line := cm.lines[idx] // shortcut
		cmHash, ok := strings.CutPrefix(line, cherryPickLinePrefix)
		if !ok { // we don't have the prefix, so we don't care
			continue
		}
		cmHash, ok = strings.CutSuffix(chHash, cherryPickLineSuffix)
		if !ok { // we don't have the suffix, so we don't care
			continue
		}
		return chMash
	}
	return "" // nothing found
}

func verifyCommitMessage(commitMessage string) error {
	cm := newCommitMessageFromString(commitMessage)
	if cm.IsEmpty() {
		return errEmptyCommitMessage
	}
	if !cm.IsKNISpecific() {
		return errMissingTagKNI
	}
	return nil
}

func main() {
	if len(os.Args) != 2 {
		programName := os.Args[0]
		log.Printf("usage: %s wrong number of arguments expects: %s <commit-message>", programName, programName)
		os.Exit(1)
	}

	err := verifyCommitMessage(os.Args[1])
	if err != nil {
		log.Printf("verification failed: %v", err)
		os.Exit(2)
	}

	os.Exit(0) // all good! redundant but let's be explicit about our success
}
