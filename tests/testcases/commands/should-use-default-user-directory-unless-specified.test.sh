tests:ensure :orgalorg:with-key -C -- pwd

tests:assert-stdout-re "${ips[0]} /home/$orgalorg_user"
