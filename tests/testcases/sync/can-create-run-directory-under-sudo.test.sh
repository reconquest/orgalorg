tests:put test-file <<EOF
line1
line2
EOF

tests:ensure :orgalorg:with-key -exmn 'true' -S test-file

:check-uploaded-file() {
    local container_ip="$2"

    # We need ssh there, because of /var/run is a tmpfs and not accessible
    # from host.
    tests:ensure ssh-test:connect:by-key "$container_ip" cat "/var/run/orgalorg/*/test-file"
    tests:assert-no-diff "test-file" "stdout"
}

containers:do :check-uploaded-file
