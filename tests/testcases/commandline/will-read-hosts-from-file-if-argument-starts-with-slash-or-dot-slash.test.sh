xargs -n1 <<< "${ips[@]}" | tests:put hosts

tests:ensure :orgalorg:with-key -o ./hosts -C echo hello '|' wc -l

tests:assert-stdout "$(containers:count)"
