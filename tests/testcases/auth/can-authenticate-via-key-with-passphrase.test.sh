passphrase="theone"

tests:ensure ssh-test:local:keygen -f "$(ssh-test:print-key-path)-encrypted" \
    -b 4096 -P "$passphrase"

tests:ensure cat $(ssh-test:print-key-path)

:copy-key() {
    local container_name="$1"
    local container_ip="$2"

    tests:ensure ssh-test:connect:by-key "$container_ip" \
        'cat > ~/.ssh/authorized_keys' \
        < "$(ssh-test:print-key-path)-encrypted.pub"
}

containers:do :copy-key

tests:ensure \
    mv "$(ssh-test:print-key-path)-encrypted" \
    "$(ssh-test:print-key-path)"

tests:eval :orgalorg:with-key-passphrase "bla-$passphrase" -C -- \
    whoami

tests:assert-stdout "invalid passphrase for private key specified"

tests:ensure :orgalorg:with-key-passphrase "$passphrase" -C -- \
    whoami

:check-output() {
    local container_name="$1"
    local container_ip="$2"

    tests:assert-stdout "$container_ip $orgalorg_user"
}

containers:do :check-output
