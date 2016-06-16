tests:ensure :orgalorg:with-key -q -C -- echo -n hello

tests:debug $(containers:count)

tests:assert-stdout "$(printf "%0.shello" $(seq ${#containers[@]}))"
