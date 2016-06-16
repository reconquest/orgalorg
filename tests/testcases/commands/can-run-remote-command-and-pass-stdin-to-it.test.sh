tests:ensure :orgalorg:with-key -C -- wc -l

containers:do tests:assert-stdout "0"

tests:ensure :orgalorg:with-key -C -i <(echo 1) -- wc -l

containers:do tests:assert-stdout "1"
