:containers:run ${containers[0]} -- /usr/bin/sshd -Dd

sleep 1s
echo prefail
fail
echo postfail



tests:eval 'echo ${#containers[@]}'
