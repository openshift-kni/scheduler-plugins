# pkg-kni/app -- Forked Upstream Functions

This directory contains Go source files that were copied and modified from
upstream Kubernetes and API server packages. The modifications allow the
KNI scheduler plugin to use a custom TLS configuration
(`SecureTLSConfig` from `library-go`) without patching vendored code.

## Why functions are forked

The upstream `kube-scheduler` command pipeline (`NewSchedulerCommand` ->
`runCommand` -> `Run` -> `SecureServingInfo.Serve`) hard-codes its TLS
setup. To inject OpenShift's `SecureTLSConfig`, the call chain is copied
locally and modified at specific points:

- `Run` is renamed to `run` and accepts a `serveFunc` callback
- `Serve` is converted to a free function `customServe` that calls
  `buildCustomTLSConfig` to apply `SecureTLSConfig` after the base TLS
  setup

## Forked declarations

### Modified copies (see `patches/` for exact diffs)

| Upstream | Local | Patch |
|----------|-------|-------|
| `NewSchedulerCommand` in `server.go` | `NewSchedulerCommand` in `sched_command.go` | `NewSchedulerCommand.patch` |
| `runCommand` in `server.go` | `runCommand` in `sched_command.go` | `runCommand.patch` |
| `Run` in `server.go` | `run` in `sched_run.go` | `Run_to_run.patch` |
| `SecureServingInfo.Serve` in `secure_serving.go` | `customServe` in `serve.go` | `Serve_to_customServe.patch` |
| `SecureServingInfo.tlsConfig` in `secure_serving.go` | `tlsConfig` in `serve.go` | `tlsConfig.patch` |

### Verbatim copies (no local modifications)

| Upstream | Local |
|----------|-------|
| `buildHandlerChain` in `server.go` | `sched_helpers.go` |
| `installMetricHandler` in `server.go` | `sched_helpers.go` |
| `newEndpointsHandler` in `server.go` | `sched_helpers.go` |
| `tlsHandshakeErrorWriter` (type) in `secure_serving.go` | `serve.go` |
| `tlsHandshakeErrorPrefix` (const) in `secure_serving.go` | `serve.go` |

## Tracking upstream changes

CI checks whether the upstream vendor files have changed since the last
review using file-level SHA256 checksums.

### Files involved

| File | Purpose |
|------|---------|
| `pkg-kni/app/upstream_checksums.sha256` | Stored checksums of the 2 tracked vendor files |
| `pkg-kni/app/patches/*.patch` | Diffs showing the local modifications for each modified function |
| `hack-kni/verify-upstream-sync.sh` | CI script that verifies checksums |
| `hack-kni/update-upstream-patches.sh` | Script to regenerate all patch files |

### Upstream vendor files tracked

- `vendor/k8s.io/kubernetes/cmd/kube-scheduler/app/server.go`
- `vendor/k8s.io/apiserver/pkg/server/secure_serving.go`

## Workflow after `go mod vendor`

When CI fails with "Upstream changes detected", follow these steps:

1. **Review the vendor diff.** See what changed in the upstream files:

   ```bash
   git diff vendor/k8s.io/kubernetes/cmd/kube-scheduler/app/server.go
   git diff vendor/k8s.io/apiserver/pkg/server/secure_serving.go
   ```

2. **Check if changes affect forked functions.** Consult the patches in
   `pkg-kni/app/patches/` to understand which functions were modified
   locally and how.

3. **Adapt local copies if needed.** If the upstream changes touch
   functions that were forked into `pkg-kni/app/`, update the local
   copies to incorporate the upstream changes while preserving local
   modifications.

4. **Regenerate checksums:**

   ```bash
   make -f Makefile.kni update-upstream-checksums
   ```

5. **Update patches if local modifications changed.** If you changed how
   a local function differs from upstream, regenerate all patches:

   ```bash
   make -f Makefile.kni update-upstream-patches
   ```

6. **Commit all updated files** (checksums, patches, and adapted local
   code).

## Make targets

| Target | Description |
|--------|-------------|
| `make -f Makefile.kni verify-upstream-sync` | Check if vendor files match stored checksums (CI) |
| `make -f Makefile.kni update-upstream-checksums` | Regenerate checksums after reviewing upstream changes |
| `make -f Makefile.kni update-upstream-patches` | Regenerate all patch files in `pkg-kni/app/patches/` |
