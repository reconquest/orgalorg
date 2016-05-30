tests:ensure :orgalorg-key -C 'echo -n hello' '|' wc -l

tests:assert-stdout "$(containers:count)"
