tests:make-tmp-dir dir

tests:put dir/test-file <<EOF
line1
line2
EOF

tests:put dir/another-file <<EOF
line3
line5
EOF

tests:ensure :orgalorg-key -e -r /home/orgalorg/ -S dir

:check-uploaded-directory() {
    local container_name="$1"

    containers:get-rootfs rootfs "$container_name"

    tests:assert-no-diff \
        "$rootfs/home/orgalorg/dir/test-file" "dir/test-file"
    tests:assert-no-diff \
        "$rootfs/home/orgalorg/dir/another-file" "dir/another-file"
}

containers:foreach :check-uploaded-directory
