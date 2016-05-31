# requires setup.sh to be sourced first!

orgalorg_user="orgalorg"

:orgalorg:with-key() {
    tests:debug "!!! connecting to hosts: ${ips[@]}"

    orgalorg -u $orgalorg_user ${ips[*]/#/-o} -k "$(:ssh:get-key)" "${@}"
}
