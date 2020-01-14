tests:put test-file <<EOF
line1
line2
EOF

tests:put sync <<EOF
sed -un '/NODE/p;/START/q'
EOF

containers:do :install-sync-command-into-container "sync"

tests:ensure :orgalorg:with-key -e -r /home/orgalorg/ -S test-file

:check-all-nodes-present-in-list() {
    local container_name="$1"
    local container_ip="$2"

    tests:assert-stdout-re "NODE.*$container_ip"
}

:check-all-nodes-has-current-flag-set-correctly() {
    tests:get-stdout \
        | grep CURRENT \
        | tr '[]@:' '    ' > output

    tests:ensure awk '$1 != $4 { exit 1 }' '<' output
}

containers:do :check-all-nodes-present-in-list
:check-all-nodes-has-current-flag-set-correctly
