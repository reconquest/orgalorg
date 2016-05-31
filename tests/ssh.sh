:ssh:get-key() {
    printf "ssh-key"
}

:ssh:get-username() {
    printf "orgalorg"
}

:ssh:run-daemon() {
    local container_name=$1
    shift

    containers:run "$container_name" -- \
        /usr/bin/sshd "${@:--Dd}"
}

:ssh:keygen-local() {
    local output_file=$1
    shift

    ssh-keygen -P '' -f "$output_file"
}

:ssh:keygen-remote() {
    local container_name=$1
    shift

    containers:run "$container_name" -- \
        /usr/bin/ssh-keygen "${@:--A}"
}

:ssh:copy-id() {
    local container_name=$1
    local username=$2
    shift

    containers:run "$container_name" -- \
        /usr/bin/tee -a /home/$username/.ssh/authorized_keys > /dev/null
}

:ssh:run-with-key() {
    local ip=$1
    local user=$2
    local identity=$3

    shift 3

    ssh \
        -oStrictHostKeyChecking=no \
        -oPasswordAuthentication=no \
        -oControlPath=none \
        -i "$identity" \
        -l "$user" \
        "$ip" "${@}"
}

:ssh() {
    local ip=$1
    shift

    :ssh:run-with-key "$ip" \
        "$(:ssh:get-username)" "$(:ssh:get-key)" \
        "${@}"
}
