tests:not tests:ensure :orgalorg:with-key -e -C echo 1 '&&' exit 1
tests:assert-stderr-re "exited with non-zero.*all $(containers:count) nodes"
tests:assert-stderr "code 1 ($(containers:count) nodes)"
