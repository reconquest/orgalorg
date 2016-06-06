tests:involve tests/testcases/locking/lock.sh

:orgalorg:lock orgalorg_output orgalorg_pid

tests:ensure :orgalorg:with-key --no-lock-fail -C -- echo 1
tests:assert-stderr "continuing"
tests:assert-stdout "1"

pkill -INT -P "$orgalorg_pid"

_exited_with_ctrl_c=130

wait "$orgalorg_pid" || tests:assert-equals "$_exited_with_ctrl_c" "$?"
