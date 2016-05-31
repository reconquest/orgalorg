tests:ensure :orgalorg:with-key -C -- echo -n hello '|' wc -l

tests:assert-stdout "$(containers:count)"
