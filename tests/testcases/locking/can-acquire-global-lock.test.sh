orgalorg_output="$(tests:get-tmp-dir)/oralorg.stdout"

tests:run-background orgalorg \
    tests:silence tests:pipe \
        :orgalorg-key --stop-at-lock '2>&1' '|' tee $orgalorg_output

while ! cat "$orgalorg_output" 2>/dev/null | grep -qF "waiting for interrupt"
do
    tests:debug "[orgalorg] waiting for global lock..."
    sleep 0.1
done

tests:debug "[orgalorg] global lock has been acquired"

tests:not tests:ensure :orgalorg-key --stop-at-lock
tests:assert-stderr "lock already"

orgalorg_pid=$(tests:get-background-pid "$orgalorg")

pkill -INT -P "$orgalorg_pid"

_exited_with_ctrl_c=130

wait "$orgalorg_pid" || tests:assert-equals "$_exited_with_ctrl_c" "$?"
