tests:ensure :orgalorg-key -C pwd

containers:do tests:assert-stdout-re "${ips[0]} /home/orgalorg"
