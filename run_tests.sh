#!/bin/bash

set -euo pipefail

cd tests/

source progress.sh

source deps.sh
source sudo.sh
source hastur.sh
source getopt.sh
source containers.sh

:usage:parse-opts opts "${@}"

:progress:start

:deps:check-hastur
:deps:check-tests.sh

:hastur:init openssh,pam

:deps:tests.sh -d testcases -s local-setup.sh "${opts[@]:--A}"
