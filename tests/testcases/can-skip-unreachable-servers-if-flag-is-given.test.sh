tests:not tests:ensure :orgalorg:with-key -o example.com -C whoami

tests:ensure :orgalorg:with-key -o example.com -w -C whoami

tests:assert-stderr-re "WARNING.*can't create runner.*example.com"

:check-node-output() {
    local container_ip="$2"

    tests:assert-stdout "$container_ip $orgalorg_user"
}

containers:do :check-node-output
