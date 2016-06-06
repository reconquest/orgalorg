tests:not tests:ensure :orgalorg:with-key -e -C exit 17

tests:assert-stderr "failed to evaluate"
tests:assert-stderr "non-zero code: 17"
