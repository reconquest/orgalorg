_tests_lib=lib/tests.sh

:deps:check-hastur() {
    which hastur >/dev/null 2>&1
    if [ $? -ne 0 ]; then
        echo "missing dependency: hastur"
        exit 1
    fi
}

:deps:hastur() {
    sudo hastur "${@}"
}

:deps:check-tests.sh() {
    if [ ! -f $_tests_lib ]; then
        echo "missing dependency: tests.sh"
        echo "trying fix this via updating git submodules"

        git submodule init
        git submodule update

        if [ ! -f $_tests_lib ]; then
            echo "file '$_tests_lib' not found"
            exit 1
        fi
    fi
}

:deps:tests.sh() {
    $_tests_lib "${@}"
}
