tests:ensure :orgalorg-key -q -C pwd

tests:assert-stdout-re "^/home/orgalorg$"
