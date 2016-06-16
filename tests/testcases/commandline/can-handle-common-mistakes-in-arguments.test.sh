tests:not tests:ensure :orgalorg -C echo blah
tests:assert-stderr "Usage:"

tests:not tests:ensure :orgalorg -o ./blah -C echo blah
tests:assert-stderr-re "can't open.*blah"
tests:assert-stderr-re "blah.*no such file or directory"

tests:not tests:ensure :orgalorg -p -s -C echo blah
tests:assert-stderr-re "incompatible options"
tests:assert-stderr-re "password authentication is not possible.*stdin"

tests:not tests:ensure :orgalorg -o blah --send-timeout=wazup -C echo dunno
tests:assert-stderr-re "send timeout to number"

tests:not tests:ensure :orgalorg -o blah --recv-timeout=wazup -C echo dunno
tests:assert-stderr-re "receive timeout to number"

tests:not tests:ensure :orgalorg -o blah --conn-timeout=wazup -C echo dunno
tests:assert-stderr-re "connection timeout to number"

tests:not tests:ensure :orgalorg -o blah --keep-alive=wazup -C echo dunno
tests:assert-stderr-re "keep alive time to number"
