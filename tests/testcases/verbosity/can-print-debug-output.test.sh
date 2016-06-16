tests:ensure :orgalorg:with-key -vv -C pwd

tests:assert-stderr-re "DEBUG.*stdout.*/home/orgalorg"
