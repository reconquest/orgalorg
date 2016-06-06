tests:involve tests/testcases/locking/lock.sh

:orgalorg:lock orgalorg_output orgalorg_pid

tests:not tests:ensure :orgalorg:with-key --lock
tests:assert-stderr "lock already"

pkill -INT -P "$orgalorg_pid"

_exited_with_ctrl_c=130

wait "$orgalorg_pid" || tests:assert-equals "$_exited_with_ctrl_c" "$?"
