_containers_count=3

:containers:spawn() {
    tests:eval :hastur -kS ${@:-/bin/true}
}

:containers:destroy() {
    local container_name=$1

    tests:eval :hastur -D "$container_name"
}

:containers:list() {
    tests:pipe :hastur -Qc | awk '{ print $1 }'
}

:containers:wipe() {
    :list-containers | while read container_name; do
        :containers:destroy "$container_name"
    done
}

:containers:run() {
    local container_name=$1
    shift

    :hastur -Sn "$container_name" "${@}"
}

:containers:list-to-var() {
    local var_name="$1"

    eval "$var_name=()"
    while read container_name; do
        eval "$var_name+=($container_name)"
    done < <(:containers:list)
}
