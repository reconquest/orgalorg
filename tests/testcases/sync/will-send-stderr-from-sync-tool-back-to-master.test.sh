tests:put test-file <<EOF
line1
line2
EOF

tests:put sync <<EOF
echo XXX >&2
EOF

containers:do :install-sync-command-into-container "sync"

tests:ensure :orgalorg:with-key -e -r /home/orgalorg/ -S test-file

:check-stderr-returned-from-all-nodes() {
    local container_name="$1"
    local container_ip="$2"

    tests:assert-stderr-re "$container_ip XXX"
}

containers:do :check-stderr-returned-from-all-nodes
