tests:ensure :orgalorg:with-key -x -C 'whoami' '&&' 'echo' '\$HOME'

containers:do tests:assert-stdout "root"
containers:do tests:assert-stdout "/root"
