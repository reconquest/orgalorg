orgalorg_output="$(tests:get-tmp-dir)/oralorg.stdout"

tests:run-background orgalorg_pid \
    tests:silence tests:pipe \
        :orgalorg-key --stop-at-lock '2>&1' '|' tee $orgalorg_output

while ! cat "$orgalorg_output" 2>/dev/null | grep -qF "global lock acquired"
do
    tests:debug "[orgalorg] waiting for global lock..."
    sleep 0.1
done

tests:debug "[orgalorg] global lock has been acquired"

tests:not tests:ensure :orgalorg-key --stop-at-lock
tests:assert-stderr "lock already"

pkill -INT -P $!

_exited_with_ctrl_c=130

wait $! || tests:assert-equals "$_exited_with_ctrl_c" "$?"
