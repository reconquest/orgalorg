# requires setup.sh to be sourced first!

orgalorg_user="orgalorg"

:orgalorg-key() {
    orgalorg -u $orgalorg_user ${ips[*]/#/-o} -k "$(:ssh:get-key)" "${@}"
}
