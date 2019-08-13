go-test:set-prefix "$(tests:print-current-testcase | sed 's/\W/_/g')-"

ssh-test:set-username "orgalorg"

:run-on-container() {
    # $container_name comes from the outer scope

    containers:run "$container_name" -- /usr/bin/$1 "${@:2}"
}

ssh-test:set-remote-runner :run-on-container

:bootstrap-container() {
    local container_name="$1"

    tests:debug "[$container_name] bootstrapping container"

    tests:ensure ssh-test:remote:keygen

    tests:ensure containers:run "$container_name" -- \
        < "$(ssh-test:print-key-path).pub" \
        /usr/bin/sh -c "
            ssh-keygen -A

            useradd -G wheel $(ssh-test:print-username)

            sed -r \"/wheel.*NOPASSWD/s/^#//\" -i /etc/sudoers

            mkdir -p \\\\
                /home/$(ssh-test:print-username)/.ssh

            cat > /home/$(ssh-test:print-username)/.ssh/authorized_keys

            chown -R \\\\
                $(ssh-test:print-username): /home/$(ssh-test:print-username)"
}

:start-ssh-daemon() {
    local container_name="$1"

    tests:debug "[$container_name] starting sshd..."

    tests:run-background "pid" ssh-test:remote:run-daemon

    until containers:is-active "$container_name"; do
        tests:debug "[$container_name] is offline"
    done

    tests:debug "[$container_name] is online"
}

:wait-for-ssh-active() {
    local container_name="$1"
    local container_ip="$2"

    until ssh-test:connect:by-key "$container_ip" "true"; do
        tests:debug "[$container_name] sshd is offline"
    done

    tests:debug "[$container_name] sshd is online"
}

:install-sync-command-into-container() {
    local file_name="$1"
    local container_name="$2"

    containers:get-rootfs rootfs "$container_name"

    tests:ensure sudo mkdir -p "$rootfs/usr/lib/orgalorg/"
    tests:ensure sudo cp "$file_name" "$rootfs/usr/lib/orgalorg/"
    tests:ensure sudo chmod +x "$rootfs/usr/lib/orgalorg/sync"
}

tests:debug "!!! setup"

tests:clone "orgalorg" "bin/"

tests:debug "!!! spawning $(containers:count) containers"

containers:spawn "/bin/true"

tests:debug "!!! generating local key pair"

tests:ensure ssh-test:local:keygen

tests:debug "!!! bootstrapping containers"

containers:foreach :bootstrap-container

tests:debug "!!! starting sshd instances"

containers:foreach :start-ssh-daemon

tests:debug "!!! waiting for sshd"

containers:do :wait-for-ssh-active

containers:get-list containers
containers:get-ip-list ips
