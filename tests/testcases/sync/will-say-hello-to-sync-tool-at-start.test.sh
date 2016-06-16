tests:put test-file <<EOF
line1
line2
EOF

tests:put sync <<EOF
sed -un '/HELLO/{p;q}'
EOF

containers:do :install-sync-command-into-container "sync"

tests:ensure :orgalorg:with-key -e -r /home/orgalorg/ -S test-file

containers:do tests:assert-stdout-re "HELLO"
