tests:ensure :orgalorg:with-key --json -vv -C pwd

tests:assert-stderr-re '"stream":"stderr"'
tests:assert-stderr-re '"body":".*DEBUG.*connection established'
tests:assert-stderr-re '"body":".*DEBUG.*running lock command'
