tests:not tests:ensure :orgalorg:with-key --json -o example.com -C pwd

tests:assert-stderr-re '"stream":"stderr"'
tests:assert-stderr-re '"body":".*ERROR.*create runner for address'
tests:not tests:assert-stderr '└─'
