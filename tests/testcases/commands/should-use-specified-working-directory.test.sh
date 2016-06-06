tests:ensure :orgalorg:with-key -r /tmp -C -- pwd

tests:assert-stdout-re "${ips[0]} /tmp"
