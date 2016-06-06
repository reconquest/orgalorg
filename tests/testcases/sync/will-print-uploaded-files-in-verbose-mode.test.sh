tests:make-tmp-dir dir

tests:put dir/test-file <<EOF
line1
line2
EOF

tests:put dir/another-file <<EOF
line3
line5
EOF

tests:ensure :orgalorg:with-key -v -e -r /home/orgalorg/ -U dir

tests:assert-stderr-re "tar.*dir/test-file"
tests:assert-stderr-re "tar.*dir/another-file"
