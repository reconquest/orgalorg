tests:ensure :orgalorg-key -v -C 'echo 1; echo 2 >&2'

tests:assert-stderr "<stdout> 1"
tests:assert-stderr "<stderr> 2"
