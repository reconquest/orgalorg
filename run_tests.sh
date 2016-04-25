#!/bin/bash

set -euo pipefail

cd tests/

source deps.sh
source sudo.sh
source hastur.sh
source getopt.sh
source containers.sh

:usage:parse-opts opts "${@}"

:deps:check-hastur
:deps:check-tests.sh

:hastur:init

:deps:tests.sh -d testcases -s local-setup.sh "${opts[@]:--A}"
