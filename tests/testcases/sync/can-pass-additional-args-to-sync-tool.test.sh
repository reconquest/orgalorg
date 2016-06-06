tests:put test-file <<EOF
line1
line2
EOF

tests:ensure :orgalorg:with-key -emn 'echo 1:$1 2:$2' \
    --args first --args second -r /tmp -S test-file

tests:assert-stdout "1:first 2:second"
