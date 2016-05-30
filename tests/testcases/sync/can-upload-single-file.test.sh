tests:put test-file <<EOF
line1
line2
EOF

tests:ensure :orgalorg-key -e -S test-file -r /home/orgalorg/

:check-uploaded-file() {
    local container_name="$1"

    containers:get-rootfs rootfs "$container_name"

    tests:assert-no-diff "$rootfs/home/orgalorg/test-file" "test-file"
}

containers:foreach :check-uploaded-file
