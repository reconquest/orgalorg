# FIXME make it possible to specify non-system root dir
_hastur_root_dir="/var/lib/hastur"

:hastur:keep-containers() {
    :hastur:destroy-containers() {
        echo -n "containers are kept in $_hastur_root_dir... "
    }

    :hastur:destroy-root() {
        :
    }
}

:hastur:keep-images() {
    :hastur:destroy-root() {
        echo -n "root is kept in $_hastur_root_dir... "
    }
}

:hastur() {
    mkdir -p $_hastur_root_dir

    :deps:hastur -q -r $_hastur_root_dir "${@}"
}

:hastur:init() {
    printf "Cheking and initializing hastur... "
    if hastur_out=$(:hastur -S /bin/true 2>&1); then
        printf "ok.\n"
    else
        printf "fail.\n\n%s\n" "$hastur_out"
    fi

    trap :hastur:cleanup EXIT
}

:hastur:destroy-containers() {
    :hastur -Qc \
        | awk '{ print $1 }' \
        | while read container_name; do
            :hastur -D $container_name
        done
}

:hastur:destroy-root() {
    :hastur --free
}

:hastur:cleanup() {
    printf "Cleaning up hastur containers...\n"

    tests:debug() {
        echo "${@}" >&2
    }

    :hastur:destroy-containers

    :hastur:destroy-root

    printf "ok.\n"
}
