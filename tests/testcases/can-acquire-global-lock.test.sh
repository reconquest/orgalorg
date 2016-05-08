#tests:silence tests:eval :ssh "${ips[0]}" pwd

orgalorg_output="$(tests:get-tmp-dir)/oralorg.stdout"

tests:run-background orgalorg_pid \
    tests:silence tests:pipe \
        "orgalorg ${ips[@]/#/-o} --stop-after-lock 2>&1 | tee $orgalorg_output"

while ! cat "$orgalorg_output" 2>/dev/null | grep -qF "global lock acquired"
do
    tests:debug "[orgalorg] waiting for global lock..."
    sleep 0.1
done

tests:debug "[orgalorg] global lock has been acquired"

tests:not tests:ensure orgalorg ${ips[@]/#/-o} --stop-after-lock
tests:assert-stderr "lock already"

pkill -INT -P $!

wait $! || true
