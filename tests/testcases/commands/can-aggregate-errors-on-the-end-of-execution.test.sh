tests:not tests:ensure :orgalorg:with-key -e -C echo 1 '&&' exit 1

:check-node-output() {
    local container_ip="$2"

    tests:assert-stdout "$container_ip 1"
    tests:assert-stderr-re "$container_ip.*non-zero code: 1"
}

containers:do :check-node-output
