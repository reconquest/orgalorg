tests:make-tmp-dir dir

tests:put dir/test-file <<EOF
line1
line2
EOF

tests:ensure :orgalorg-key -S dir -r /home/orgalorg/

for container_name in "${containers[@]}"; do
    containers:get-rootfs rootfs "$container_name"

    file_name=$(tests:pipe ls $rootfs/home/orgalorg/tmp/*/root/dir/test-file)

    tests:assert-no-diff "$file_name" "dir/test-file"
done
