# Maintaining openshift-kni/scheduler-plugins

openshift-kni/scheduler-plugins is based on upstream kubernetes-sigs/scheduler-plugins.
With every release of kubernetes-sigs/scheduler-plugins, it is necessary to incorporate the upstream changes
while ensuring that our downstream customizations are maintained.

Nonetheless, we have the freedom to choose if we want this changes at all, because there are times when the upstream
changes are not relevant for our work.

## Master Branch Upstream Resync Strategy
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

### Patching openshift-kni specific commits

Every commit that is openshift-kni/scheduler-plugins specific should have a prefix of [KNI] 
at the beginning of the commit message.

### Document changes

For the sake of transparency, for every resync process we should update the following table:

| Resync Date | Merge With Upstream Tag/Commit                                                                         | Author      |
|-------------|--------------------------------------------------------------------------------------------------------|-------------|
| 2023.05.23  | https://github.com/kubernetes-sigs/scheduler-plugins/commit/ab6c864e24ca08a38efc609524ce10bce8d3db3b   | Tal-or      |
| 2022.10.11  | https://github.com/kubernetes-sigs/scheduler-plugins/commit/81fb4607af1d45ebf76eb7fbd0eb7ddba7abc959   | swatisehgal |
| 2022.07.06  | https://github.com/kubernetes-sigs/scheduler-plugins/commit/6aadda4e9213fd0f71807cd6630eb8e58db740fd   | swatisehgal |
| 2022.06.29  | https://github.com/kubernetes-sigs/scheduler-plugins/commit/843d47374bba691f13558806e8fddb866bfb1b9e   | swatisehgal |
| 2022.06.23  | https://github.com/kubernetes-sigs/scheduler-plugins/commit/54bd848cd75ce5c0b6953733b0e477c47aa356a9   | swatisehgal |
| 2022.05.03  | https://github.com/kubernetes-sigs/scheduler-plugins/commit/7a6dcdc99b1ee9a324823eaf98718cfd9e98e805   | fromanirh   |
| 2022.03.10  | https://github.com/kubernetes-sigs/scheduler-plugins/commit/2b3439076c54579c3ecdacfc71ca00a23f1e42f8   | fromanirh   |
| 2022.01.21  | https://github.com/kubernetes-sigs/scheduler-plugins/commit/ec632c3d7e04b7b372f9a6f4338b0dbc53ef3d46   | fromanirh   |
| 2021.12.23  | https://github.com/kubernetes-sigs/scheduler-plugins/commit/7cf6512bd726f0d30b2ab32443af867a0b849da8   | fromanirh   |
| 2021.12.11  | https://github.com/kubernetes-sigs/scheduler-plugins/commit/b8d13e17a3e1f633d72d71276a3da6fecf89f2e3   | Tal-or      |

The newest resync should appear in the first row. 

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
