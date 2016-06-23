tests:not tests:ensure :orgalorg:with-key -e -C exit 17

tests:assert-stderr "remote execution failed"
tests:assert-stderr "code 17 ($(containers:count) nodes)"
