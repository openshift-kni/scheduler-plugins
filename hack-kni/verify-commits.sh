#!/bin/bash

set -e -o pipefail

function finish {
	if [ -f "$commit_msg_filename" ]; then
		rm -f "$commit_msg_filename"
	fi
}
trap finish EXIT

echo "checking branch: [$TRIGGER_BRANCH]"

shopt -s extglob
if [[ "$TRIGGER_BRANCH" == resync-* ]]; then
  echo "WARN: resync branch no commit enforcement will be triggered"
  exit 0
fi

if (( $COMMITS <= 0 )); then
	echo "WARN: no changes detected"
	exit 0
fi

echo "considering ${COMMITS} commits in PR whose head is $TRIGGER_BRANCH (into $UPSTREAM_BRANCH):"
echo "---"
git log --oneline --no-merges -n ${COMMITS}
echo "---"

# list commits
for commitish in $( git log --oneline --no-merges -n ${COMMITS} | cut -d' ' -f 1); do
  echo "CHECK: $commitish"
  .github/hooks/commit-msg $( git log --format=%s -n 1 "$commitish" )
  if [[ "$?" != "0" ]]; then
    echo "-> FAIL: $commitish"
    exit 20
  fi
  echo "-> PASS: $commitish"
done

