#!/bin/bash

set -euo pipefail

_base_dir=$(dirname "$(readlink -f "${BASH_SOURCE[0]}")")
source "$_base_dir/vendor/github.com/reconquest/import.bash/import.bash"

import "github.com/reconquest/hastur"
import "github.com/reconquest/containers"
import "github.com/reconquest/progress"
import "github.com/reconquest/test-runner"
import "github.com/reconquest/tests.sh"

include tests/ssh.sh
include tests/build.sh
include tests/orgalorg.sh

test-runner:set-custom-opts \
    --keep-containers \
    --keep-images \
    --containers-count:

test-runner:handle-custom-opt() {
    case "$1" in
        --keep-containers)
            :hastur:keep-containers
            ;;

        --keep-images)
            :hastur:keep-images
            ;;

        --containers-count)
            :containers:set-count "$2"
            ;;
    esac
}

which brctl >/dev/null 2>&1
if [ $? -ne 0 ]; then
    echo "missing dependency: brctl (bridge-utils)"
    exit 1
fi

which hastur >/dev/null 2>&1
if [ $? -ne 0 ]; then
    echo "missing dependency: hastur"
    exit 1
fi

progress:spinner:new _progress_spinner

{
    :build:init

    hastur:init openssh,pam,util-linux,tar

} 2> >(progress:spinner:spin "$_progress_spinner" > /dev/null)

:cleanup() {
    containers:wipe

    progress:spinner:stop "$_progress_spinner"
}

trap :cleanup EXIT

test-runner:run "${@}"
