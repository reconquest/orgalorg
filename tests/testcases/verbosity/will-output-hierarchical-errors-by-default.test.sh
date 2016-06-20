tests:not tests:ensure :orgalorg:with-key -o example.com -C pwd

tests:assert-stderr "└─ can't connect to address"
