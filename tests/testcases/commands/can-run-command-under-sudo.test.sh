tests:ensure :orgalorg-key -C 'whoami'

containers:do tests:assert-stdout "$orgalorg_user"

tests:ensure :orgalorg-key -x -C 'whoami'

containers:do tests:assert-stdout "root"
