tests:ensure :orgalorg:with-key -C 'whoami'

containers:do tests:assert-stdout "$orgalorg_user"

tests:ensure :orgalorg:with-key -x -C 'whoami'

containers:do tests:assert-stdout "root"
