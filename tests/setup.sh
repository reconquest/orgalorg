:bootstrap-container() {
    local container_name="$1"

    tests:debug "[$container_name] bootstrapping container"

    tests:ensure :ssh:keygen-remote "$container_name"

    tests:ensure containers:run "$container_name" -- \
        /usr/bin/sh -c "
            /usr/bin/useradd -G wheel $(:ssh:get-username)

            /usr/bin/mkdir -p \\\\
                /home/$(:ssh:get-username)/.ssh

            /usr/bin/chown -R \\\\
                $(:ssh:get-username): /home/$(:ssh:get-username)" \

    tests:ensure :ssh:copy-id "$container_name" \
        "$(:ssh:get-username)" < "$(:ssh:get-key).pub"
}

:start-ssh-daemon() {
    local container_name="$1"

    tests:debug "[$container_name] starting sshd..."

    tests:run-background "pid" :ssh:run-daemon "$container_name" "-D"

    while ! containers:is-active "$container_name"; do
        tests:debug "[$container_name] is offline"
    done

    tests:debug "[$container_name] is online"
}

:wait-for-ssh-active() {
    local container_name="$1"
    local container_ip="$2"

    while ! :ssh "$container_ip" "true"; do
        tests:debug "[$container_name] sshd is offline"
    done

    tests:debug "[$container_name] sshs is online"
}

tests:debug "!!! setup"

tests:clone "orgalorg" "bin/"

tests:debug "!!! spawning $(containers:count) containers"

containers:spawn "/bin/true"

tests:debug "!!! generating local key pair"

tests:ensure :ssh:keygen-local "$(:ssh:get-key)"

tests:debug "!!! bootstrapping containers"

containers:foreach :bootstrap-container

tests:debug "!!! starting sshd instances"

containers:foreach :start-ssh-daemon

tests:debug "!!! waiting for sshd"

containers:do :wait-for-ssh-active

containers:get-list containers
containers:get-ip-list ips
