tests:clone orgalorg bin/

tests:debug "!!! spawning $(:containers:count) containers"

for (( i = $(:containers:count); i > 0; i-- )); do
    tests:ensure :containers:spawn
done

tests:debug "!!! generating local key pair"

tests:ensure :ssh:keygen-local "$(:ssh:get-key)"

:containers:get-list "containers"

tests:debug "!!! bootstrapping containers"

for container_name in "${containers[@]}"; do
    tests:ensure :ssh:keygen-remote "$container_name"

    tests:ensure :containers:run "$container_name" -- \
        /usr/bin/sh -c "
            /usr/bin/useradd -G wheel $(:ssh:get-username)

            /usr/bin/mkdir -p \\\\
                /home/$(:ssh:get-username)/.ssh

            /usr/bin/chown -R \\\\
                $(:ssh:get-username): /home/$(:ssh:get-username)" \

    tests:ensure :ssh:copy-id "$container_name" \
        "$(:ssh:get-username)" < "$(:ssh:get-key).pub"
done

tests:debug "!!! running SSH daemon on containers"

for container_name in "${containers[@]}"; do
    tests:run-background "pid" :ssh:run-daemon "$container_name" "-D"

    while ! :containers:is-active "$container_name" :; do
        tests:debug "[$container_name] is offline"
    done

    tests:debug "[$container_name] is online"
done

:containers:get-ip-list "ips"

for container_index in "${!containers[@]}"; do
    container_name=${containers[$container_index]}

    while ! :ssh "${ips[$container_index]}" "true"; do
        tests:debug "[$container_name] SSH daemon is not running"
    done

    tests:debug "[$container_name] SSH daemon is online"
done
