tests:put test-file <<EOF
line1
line2
EOF

tests:put sync <<EOF
sed -un '/NODE/p;/START/q'
EOF

containers:do :install-sync-command-into-container "sync"

tests:ensure :orgalorg:with-key -e -r /home/orgalorg/ -S test-file

:check-node-output() {
    local container_name="$1"
    local container_ip="$2"

    tests:assert-stdout-re "NODE.*$container_ip"
}

containers:do :check-node-output
