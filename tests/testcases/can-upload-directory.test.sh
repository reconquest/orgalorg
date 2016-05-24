tests:make-tmp-dir dir

tests:put dir/test-file <<EOF
line1
line2
EOF

tests:put dir/another-file <<EOF
line3
line5
EOF

tests:ensure :orgalorg-key -e -S dir -r /home/orgalorg/

for container_name in "${containers[@]}"; do
    containers:get-rootfs rootfs "$container_name"

    tests:assert-no-diff \
        "$rootfs/home/orgalorg/dir/test-file" "dir/test-file"
    tests:assert-no-diff \
        "$rootfs/home/orgalorg/dir/another-file" "dir/another-file"
done
