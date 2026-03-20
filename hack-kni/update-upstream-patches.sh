#!/usr/bin/env bash
#
# Copyright 2026 Red Hat, Inc.
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#     http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.

# Regenerate all upstream-vs-local patch files in pkg-kni/app/patches/.
# Uses awk to extract top-level Go functions by signature, so it does not
# rely on hardcoded line numbers.

set -o errexit
set -o nounset
set -o pipefail

SCRIPT_ROOT=$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)
PATCHES_DIR="${SCRIPT_ROOT}/pkg-kni/app/patches"

SERVER_GO="vendor/k8s.io/kubernetes/cmd/kube-scheduler/app/server.go"
SECURE_SERVING_GO="vendor/k8s.io/apiserver/pkg/server/secure_serving.go"

# extract_func extracts a top-level Go function or method from a file.
# It matches the first line containing the given pattern and prints
# everything up to and including the closing brace at column 0.
extract_func() {
    local file="$1" pattern="$2"
    ${SCRIPT_ROOT}/bin/extract-func "$file" "$pattern"
}

mkdir -p "${PATCHES_DIR}"

TMPDIR=$(mktemp -d)
trap 'rm -rf "${TMPDIR}"' EXIT

generate_patch() {
    local upstream_file="$1" upstream_pattern="$2"
    local local_file="$3" local_pattern="$4"
    local upstream_label="$5" local_label="$6"
    local patch_name="$7"

    extract_func "${upstream_file}" "${upstream_pattern}" > "${TMPDIR}/upstream"
    extract_func "${local_file}" "${local_pattern}" > "${TMPDIR}/local"

    diff -u "${TMPDIR}/upstream" "${TMPDIR}/local" \
        --label "${upstream_label}" \
        --label "${local_label}" \
        > "${PATCHES_DIR}/${patch_name}" || true

    if [ ! -s "${PATCHES_DIR}/${patch_name}" ]; then
        rm -f "${PATCHES_DIR}/${patch_name}"
        echo "  IDENTICAL  ${patch_name} (no diff, removed)"
    else
        echo "  GENERATED  ${patch_name}"
    fi
}

echo "Regenerating upstream patches..."

generate_patch \
    "${SERVER_GO}" "NewSchedulerCommand" \
    "pkg-kni/app/sched_command.go" "NewSchedulerCommand" \
    "upstream/NewSchedulerCommand (${SERVER_GO})" \
    "local/NewSchedulerCommand (pkg-kni/app/sched_command.go)" \
    "NewSchedulerCommand.patch"

generate_patch \
    "${SERVER_GO}" "runCommand" \
    "pkg-kni/app/sched_command.go" "runCommand" \
    "upstream/runCommand (${SERVER_GO})" \
    "local/runCommand (pkg-kni/app/sched_command.go)" \
    "runCommand.patch"

generate_patch \
    "${SERVER_GO}" "Run" \
    "pkg-kni/app/sched_run.go" "run" \
    "upstream/Run (${SERVER_GO})" \
    "local/run (pkg-kni/app/sched_run.go)" \
    "Run_to_run.patch"

generate_patch \
    "${SECURE_SERVING_GO}" "Serve" \
    "pkg-kni/app/serve.go" "customServe" \
    "upstream/SecureServingInfo.Serve (${SECURE_SERVING_GO})" \
    "local/customServe (pkg-kni/app/serve.go)" \
    "Serve_to_customServe.patch"

generate_patch \
    "${SECURE_SERVING_GO}" "tlsConfig" \
    "pkg-kni/app/serve.go" "tlsConfig" \
    "upstream/SecureServingInfo.tlsConfig (${SECURE_SERVING_GO})" \
    "local/tlsConfig (pkg-kni/app/serve.go)" \
    "tlsConfig.patch"

echo ""
echo "Patches updated in ${PATCHES_DIR}/"
