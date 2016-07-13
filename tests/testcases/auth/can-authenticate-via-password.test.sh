password="123456"

:set-ssh-password() {
    local container_ip="$2"

    ssh-test:connect:by-key "$container_ip" sudo -n chpasswd \
        <<< "$orgalorg_user:$password"
}

containers:do :set-ssh-password

tests:ensure :orgalorg:with-password "$password" -C -- whoami

:check-output() {
    local container_name="$1"
    local container_ip="$2"

    tests:assert-stdout "$container_ip $orgalorg_user"
}

containers:do :check-output
