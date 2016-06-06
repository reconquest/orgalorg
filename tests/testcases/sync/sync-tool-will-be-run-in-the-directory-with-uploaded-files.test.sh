tests:put test-file <<EOF
line1
line2
EOF

tests:ensure :orgalorg:with-key -e -n "ls" -r /tmp -S test-file

tests:assert-stdout test-file
