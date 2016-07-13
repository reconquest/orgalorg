tests:involve tests/testcases/locking/lock.sh

:orgalorg:lock orgalorg_output orgalorg_pid --send-timeout=2000

tests:wait-file-changes "$orgalorg_output" 0.1 10 \
    ssh-test:connect:by-key "${ips[0]}" pkill -f flock

tests:ensure grep -q "ERROR.*${ips[0]}.*heartbeat" "$orgalorg_output"
