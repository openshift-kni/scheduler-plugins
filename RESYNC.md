# Maintaining openshift-kni/scheduler-plugins

openshift-kni/scheduler-plugins is based on upstream kubernetes-sigs/scheduler-plugins.
With every release of kubernetes-sigs/scheduler-plugins, it is necessary to incorporate the upstream changes
while ensuring that our downstream customizations are maintained.

Nonetheless, we have the freedom to choose if we want this changes at all, because there are times when the upstream
changes are not relevant for our work.

## Master Branch Upstream Resync Strategy: upstream merges (preferred approach)
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

Note: following RH internal agreement that resulted in https://github.com/openshift-kni/scheduler-plugins/pull/285/commits/4d80e29f16036cf8a75890785d4bf3c3fb914022, future resyncs that pull `go.mod` updates from u/s will require to run ` make -f Makefile.kni update-vendor` in order to update the vendor resources.

### Patching openshift-kni specific commits

Every commit that is openshift-kni/scheduler-plugins specific should have a prefix of [KNI] 
at the beginning of the commit message.

### Document changes

For the sake of transparency, for every resync process we should update the table in `RESYNC.log.md`. The newest resync should appear in the first row. 

## Release Branch Upstream Resync Strategy

### Key Requirements/Constraints

1. Ensure that all the upstream features/updates related to NUMA resource topology scheduler plugin including test updates are backported to both release-4.10 and release-4.11 branches.
1. As much as possible, keep the base image and golang version intact (go version 1.16 in release-4.10 and go version 1.18 in release-4.11). This is to ensure changes for Z versions are minimal.
1. Commit history and authorship MUST be preserved when code is synched with upstream irrespective of whether it is full resync or a few PRs are cherry picked. One way of achieving this is to use `git cherry-pick -x <commit>` as it allows preservation of original commit reference within the cherry-picked commit.

### Preparing the local repo clone
Clone from a personal fork of openshift-kni/scheduler-plugins via a pushable (ssh) url:

`git clone git@github.com:openshift-kni/scheduler-plugins.git`

Add a remote for upstream and fetch its branches:

`git remote add --fetch upstream https://github.com/kubernetes-sigs/scheduler-plugins`

### Creating a new local branch for the new resync

Branch the target openshift-kni/scheduler master branch to a new resync local branch

`git checkout <release branch>` e.g. `git checkout release-4.10`

`git checkout -b "update-$(git symbolic-ref --short HEAD)-$(date +%Y%m%d)"` e.g. `git checkout -b update-release-4.10-20220707`

### Cherry Pick changes from PRs

Go to the upstream PR you are trying to cherrypick and identify the commit hashes.

`git cherry-pick -x <start-commit-hash>^..<end-commit-hash>`

NOTE: -x flag above is used to preserve original commit reference

Fix conflicts introduced by kni-local changes and send PR for review.

### Patching openshift-kni specific commits

Make sure to run `go mod tidy` and `go mod vendor` to ensure that the repo is in consistent state.
Every commit that is openshift-kni/scheduler-plugins specific should have a prefix of [KNI]
at the beginning of the commit message.

## Master Branch Upstream Resync Strategy: upstream carries

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
