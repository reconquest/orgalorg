_long_opts=$(paste -sd, <<OPTS
keep-containers
keep-images
containers-count:
OPTS
)

set -- $(getopt --long $_long_opts -o AOv -- "${@}")

while :; do
    case "$1" in
        --keep-containers)
            :hastur:keep-containers
            ;;
        --keep-images)
            :hastur:keep-images
            ;;
        --)
            shift
            break
            ;;
    esac

    shift
done
