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

set -o errexit
set -o nounset
set -o pipefail

SCRIPT_ROOT=$(dirname "${BASH_SOURCE[0]}")/..
CHECKSUMS="${SCRIPT_ROOT}/pkg-kni/app/upstream_checksums.sha256"

echo "Verifying upstream vendor file checksums..."

if ! sha256sum -c "${CHECKSUMS}" --quiet 2>/dev/null; then
    echo ""
    echo "============================================================"
    echo "Upstream changes detected that might affect forked files"
    echo "in pkg-kni/app/."
    echo ""
    echo "Review requested -- see pkg-kni/app/README.md for directions."
    echo ""
    echo "Once resolved, run:"
    echo "  make -f Makefile.kni update-upstream-checksums"
    echo "============================================================"
    exit 1
fi

echo "All upstream checksums match. No vendor drift detected."
