#!/bin/bash

set -euo pipefail

which hastur >/dev/null 2>&1
if [ $? -ne 0 ]; then
    echo "missing dependency: hastur"
    exit 1
fi

tests_lib=tests/lib/tests.sh

if [ ! -f $tests_lib ]; then
    echo "missing dependency: tests.sh"
    echo "trying fix this via updating git submodules"

    git submodule init
    git submodule update

    if [ ! -f $tests_lib ]; then
        echo "file '$tests_lib' not found"
        exit 1
    fi
fi

source $tests_lib
source tests/hastur.sh
source tests/getopt.sh

:hastur:init

$tests_lib -d tests/testcases -s tests/local-setup.sh ${@:--A}
