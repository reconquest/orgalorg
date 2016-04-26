export _containers_count=${_containers_count:-1}

:containers:set-count() {
    _containers_count=$1
}

:containers:count() {
    echo "$_containers_count"
}

:containers:spawn() {
    tests:pipe :hastur -p $(:hastur:get-packages) -kS ${@:-/bin/true}
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

    tests:run-background :containers:spawn -n "$container_name" "${@}"
}

:containers:list-to-var() {
    local var_name="$1"

    eval "$var_name=()"
    while read container_name; do
        eval "$var_name+=($container_name)"
    done < <(:containers:list)
}

:containers:get-ip() {
    local container_name="$1"

    :hastur -Q $container_name --ip
}
