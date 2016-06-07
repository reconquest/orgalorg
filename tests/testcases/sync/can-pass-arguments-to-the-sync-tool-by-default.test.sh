tests:put test-file <<EOF
line1
line2
EOF

tests:put sync <<EOF
echo 1: \$1
echo 2: \$2
EOF

containers:do :install-sync-command-into-container "sync"

tests:ensure :orgalorg:with-key -g -a -g -b -e -r /home/orgalorg/ -S test-file

:check-node-output() {
    local container_name="$1"
    local container_ip="$2"

    tests:assert-stdout-re "$container_ip.*1: -a"
    tests:assert-stdout-re "$container_ip.*2: -b"
}

containers:do :check-node-output
