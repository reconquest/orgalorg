tests:ensure :orgalorg-key -C -- wc -l

containers:do tests:assert-stdout "0"

tests:ensure :orgalorg-key -C -i <(echo 1) -- wc -l

containers:do tests:assert-stdout "1"
