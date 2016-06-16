tests:ensure :orgalorg:with-key --json -C pwd

tests:assert-stdout-re '"stream":"stdout"'
tests:assert-stdout-re "\"node\":\".*${ips[0]}.*\""
tests:assert-stdout-re "\"body\":\"/home/$orgalorg_user"
