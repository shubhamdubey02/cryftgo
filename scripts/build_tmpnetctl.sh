#!/usr/bin/env bash

set -euo pipefail

# Cryftgo root folder
AVALANCHE_PATH=$( cd "$( dirname "${BASH_SOURCE[0]}" )"; cd .. && pwd )
# Load the constants
source "$AVALANCHE_PATH"/scripts/constants.sh

echo "Building tmpnetctl..."
go build -ldflags\
   "-X github.com/shubhamdubey02/cryftgo/version.GitCommit=$git_commit $static_ld_flags"\
   -o "$AVALANCHE_PATH/build/tmpnetctl"\
   "$AVALANCHE_PATH/tests/fixture/tmpnet/cmd/"*.go
