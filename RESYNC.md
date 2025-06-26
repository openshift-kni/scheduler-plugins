# Maintaining openshift-kni/scheduler-plugins

openshift-kni/scheduler-plugins is based on upstream kubernetes-sigs/scheduler-plugins.
With every release of kubernetes-sigs/scheduler-plugins, it is necessary to incorporate the upstream changes
while ensuring that our downstream customizations are maintained.

Nonetheless, we have the freedom to choose if we want this changes at all, because there are times when the upstream
changes are not relevant for our work.

## Main Branch Upstream Resync Strategy: upstream merge flow (preferred approach)

### Preparing the local repo clone
Clone from a personal fork of openshift-kni/scheduler-plugins via a pushable (ssh) url:

`git clone git@github.com:openshift-kni/scheduler-plugins.git`

Add a remote for upstream and fetch its branches:

`git remote add --fetch upstream https://github.com/kubernetes-sigs/scheduler-plugins`

### Creating a new local branch for the new resync

Branch the target openshift-kni/scheduler master branch to a new resync local branch 

`git checkout master`

`git checkout -b "resync-$(date +%Y%m%d)"`

### Merge changes from upstream

`git merge upstream/master`

fix conflicts introduced by kni-local changes and send PR for review

Note: following RH internal agreement that resulted in https://github.com/openshift-kni/scheduler-plugins/pull/285/commits/4d80e29f16036cf8a75890785d4bf3c3fb914022,
future resyncs that pull `go.mod` updates from u/s will require to run ` make -f Makefile.kni update-vendor` in order to update the vendor resources.

### Patching openshift-kni specific commits

Every commit that is openshift-kni/scheduler-plugins specific should have a prefix of [KNI] 
at the beginning of the commit message.

### Document changes

For the sake of transparency, for every resync process we should update the table in `RESYNC.log.md`. The newest resync should appear in the first row. 

## Stable Branches Resync Strategy: cherry-pick flow

The cherry pick flow should be used to pull targeted changes from a more recent branch to an older branch that
upstream didn't backport itself.

### Steps

Go to the upstream PR you are trying to cherrypick and identify the commit hashes.

`git cherry-pick -x <start-commit-hash>^..<end-commit-hash>`

NOTE: -x flag above is used to preserve original commit reference.
**We must always use it to preserve proper reference**

Fix conflicts introduced by KNI-local changes and send PR for review.
Since cherry-picking always mutates the original commit hash (see Appendix B below),
cherry-picks should be further amended to have

- the `[KNI]` tag.
- the `[upstream]` tag

*added* to their commit message summary line. Example (note completely made up bogus commits):

```
commit abcdef123456
Author: Francesco Romani <fromani@redhat.com>
Date:   Wed Dec 13 13:46:27 2023 +0100

fix: don't crash when the calendar day is even

The foobarizer was misconfigured to trigger on even days,
should trigger only on leap years. This caused a crash.

Signed-off-by: Francesco Romani <fromani@redhat.com>
```

becomes

```
commit a1b2c3d4e5f6
Author: Francesco Romani <fromani@redhat.com>
Date:   Wed Dec 13 13:46:27 2023 +0100

[KNI][upstream] fix: don't crash when the calendar day is even

The foobarizer was misconfigured to trigger on even days,
should trigger only on leap years. This caused a crash.

Signed-off-by: Francesco Romani <fromani@redhat.com>
(cherry picked from commit abcdef123456)
```

### Justification for the `[KNI]` tag

The presence of *both* the `[KNI]` tag and the original commit reference added by `-x` flag unambiguously
enable us to identify a cherry picked commit.
The presence of the `[KNI]` tag is justified by the fact that that specific commit backported in a branch
is a KNI-specific change, thus should be marked as such.

### Cherry-picking KNI-specific fixes to older branches: aka backporting KNI-specific changes

Should we need to backport fixes to release branches, e.g. build system/CI changes, we would
just omit the `[upstream]` tag and proceed as outlined in the previous section.

Note that there is a fair amount of mess in our history we did before to fully specify the flow.
We can't rewrite history, so we will need to carry that. Using `cherry-pick -x` should be sufficient
to enable us to track the origin of each change given enough effort, effort which the current
process strive to minimize.

#### Cascading changes

When backporting fixes in a cascade way (main -> release-X.Y -> release-X.Y-1 -> release-X.Y-2...) the
most straightforward way is to cherry pick from one branch to the other, so the `(cherry picked from commit...)`
references will get appended creating a potentially long list.
While this is not a problem per se, and acceptable, is also unnecessary, The cherry-picked commits can be either
1. amended to remove all but the reference pointing back to the main branch commit or
2. cherry-picked always from main branch (**note this is just and only for KNI-SPECIFIC CHANGES, which are usually few
and clearly self contained, mostly affecting CI or infra in general, not production code**)
3. last resort: kept as-is. This is tedious, but not wrong.

### Branch-specific changes

Even if we strive to minimize the chances of this occurrence, sometimes it is possible that only a subset
of stable branches will need a fix. Examples are fixing CI failures, or updating metadata, or CI configuration.
In this case all the provisions apply, but the commit message subject line should include the `[release-X.Y]` tag.

Example (note completely made up bogus commits):

```
commit 65ab43de21ef
Author: Francesco Romani <fromani@redhat.com>
Date:   Wed Dec 13 13:46:27 2023 +0100

[KNI][release-4.15] ci: fix: update images for CI tests

The base image v1.23 is out of support.
Bump to 1.42, like we did for version 4.16 a while ago.

Signed-off-by: Francesco Romani <fromani@redhat.com>
```

## Patching openshift-kni specific commits which don't exist upstream

Make sure to run `go mod tidy` and `go mod vendor` to ensure that the repo is in consistent state.
Every commit that is openshift-kni/scheduler-plugins specific should have a prefix of [KNI]
at the beginning of the commit message.

## Appendix A: upstream carries

There are cases on which we cannot resync with upstream using the preferred merge approach described above.
Even though upstream is usually slower and deliberate consuming k8s libraries, there are cases on which
we may want to pull features or fixes in stable branches, and upstream just moved too far.

In these cases we do `upstream carries`.

A `upstream carry` is the target backport of one or more individual commits cherry-picked from upstream PRs
and repacked in a new PR. `upstream carries` are special-purpose in nature, so we can't have strict
guidelines like for `merge`s. Nevertheless, **all** the following guidelines apply.

- The `upstream-carry` PR MUST include the tag `[upstream-carry]` in its title
- The `upstream-carry` PR MUST have the [`upstream-carry` label](https://github.com/openshift-kni/scheduler-plugins/labels/upstream-carry)
- The cherry-picked commits MUST keep **all** the authorship information (see `Cherry Pick changes from PRs` and **always** use `git cherry-pick -x ...`)
- The `upstream carry` PR MAY include one or more cherry-picked commits
- The `upstream carry` PR MAY reference on its github cover letter the upstream PRs from which it takes commits

## Appendix B: notes about the merge process vs the cherry-pick process

The main goal while maintaining this repo are
1. keep as close as possible upstream commits, ideally verbatim (bit-for-bit identical git commits)
2. clearly preserve authorship and origin of the code
2. clearly identify KNI-specific changes, be them backports or KNI-exclusive (e.g. build system, helpers) work

The ideal git flow to enable the above requirements is the merge flow.
See figure 25 in https://git-scm.com/book/en/v2/Git-Branching-Basic-Branching-and-Merging

In this figure, we can think about commits C0-C2,C4 as upstream commits which are imported 1:1 (req 1) in this repo;
KNI specific changes, for example DOWNSTREAM_OWNERS or build-specific configuration can be commits C3,C5.
The merge commit C6 is overhead and a pure git artifact which doesn't affect the original history, so it counts
as harmless overhead and doesn't conflict with our goals.

At some point, though, we will need to pull more changes from upstream.
The ideal flow here is the git fast forward merge. To understand that, we need to perform a digression.

We use git commit hash values as proxy for repo content. IOW, we assume that if git commit hash matches, then
the content of the repo whose HEAD is at any given hash matches with upstream.
But how's a git commit hash computed? PTAL to https://stackoverflow.com/questions/14676715/cherry-pick-a-commit-and-keep-original-sha-code

Most notable, *if the references to parent commit changes, then the commit hash changes*, even if everything else is bit-for-bit identical.
Taking into account that a merge commit is a commit with 2 or more parents (vs a regular commit usually has 1 parent):

- when we pull from upstream using the "upstream merge flow" we effectively do a fast forward rebase, because we don't alter the upstream
  commits, we just add our own independent commits + merge commits. See "fast forward merge" in https://git-scm.com/docs/git-rebase
  Arguably, the name of the "upstream merge flow" is a bit misleading; the name comes from the fact the we use `git merge` to pull changes,
  but it is effectively a fast-forward

- when we pull changes using the cherry-pick process, we cannot use a fast-forward merge. Cherry-picking intrinsically changes
  the parent of a commit, thus alters the git commit hash even if the author information, message and content are bit-for-bit identical.
  This is the reason why we need to preserve the original commit hash (note this is effectively a change in the orignal commit message)
  and we are allowed to further change the commit message adding our `[KNI]` tag.

TODO (fromani) add diagrams to clarify.
