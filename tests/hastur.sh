# FIXME make it possible to specify non-system root dir
export _hastur_root_dir=${_hastur_root_dir:-/var/lib/hastur}

export _hastur_packages=${_hastur_packages:-bash,coreutils,shadow}

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

:hastur:get-packages() {
    echo $_hastur_packages
}

:hastur() {
    :sudo hastur -q -r $_hastur_root_dir "${@}"
}

:hastur:init() {
    local progress_indicator=$1
    shift

    printf "Cheking and initializing hastur... "

    mkdir -p $_hastur_root_dir

    _hastur_packages=$_hastur_packages",$1"

    local hastur_out

    if hastur_out=$(
        :hastur -p $_hastur_packages -S /usr/bin/true 2>&1 \
            | progress:spinner:spin "$progress_indicator"
    )
    then
        printf "ok.\n"
    else
        printf "fail.\n\n%s\n" "$hastur_out"
        return 1
    fi
}

:hastur:destroy-containers() {
    :hastur -Qc \
        | awk '{ print $1 }' \
        | while read container_name; do
            :hastur -f -D $container_name
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

:hastur:is-active() {
    local container_name=$1
    shift

    :sudo:silent hastur -q -r $_hastur_root_dir -Q "$container_name" --ip \
        2>/dev/null >/dev/null
}
