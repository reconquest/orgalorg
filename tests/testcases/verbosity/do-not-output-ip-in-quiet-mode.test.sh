tests:ensure :orgalorg:with-key -q -C pwd

tests:assert-stdout-re "^/home/orgalorg$"
