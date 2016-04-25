:usage() {
    cat <<HELP
Usage: run_tests.sh [options] [-v] [(-A | -O)]

Options:
    --keep-containers  Do not remove containers after every testcase.
    --keep-images      Remove containers, but keep base images to speed up
                        starting.
    --containers       Set containers count [default: 3].
    -A                 Run all testcases.
    -O                 Run last failed testcase.
    -v                 Set verbosity level.
HELP
}

:usage:parse-opts() {
    local opts_var="$1"
    shift

    local long_opts=$(paste -sd, <<OPTS
help
keep-containers
keep-images
containers:
OPTS
    )

    set -- $(getopt --long $long_opts -o vhAO:: -- "${@}")

    :containers:set-count 3

    local opts=()

    while :; do
        case "$1" in
            --keep-containers)
                shift
                :hastur:keep-containers
                ;;

            --keep-images)
                shift
                :hastur:keep-images
                ;;

            --containers)
                shift
                :containers:set-count "$1"
                ;;

            -h|--help)
                :usage
                exit 1
                ;;

            --)
                shift
                break
                ;;

            *)
                opts+=("$1")
                shift
        esac
    done

    eval "$opts_var=(${opts})"
}
