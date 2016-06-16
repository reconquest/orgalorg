# requires setup.sh to be sourced first!

orgalorg_user="orgalorg"

:orgalorg:with-key() {
    :orgalorg \
        -u $orgalorg_user ${ips[*]/#/-o} -k "$(:ssh:get-key)" "${@}"
}

:orgalorg:with-password() {
    local password="$1"
    shift

    :expect() {
        expect -f <(cat) -- "${@}" </dev/tty
    }

    go-test:run :expect -u $orgalorg_user ${ips[*]/#/-o} -p "${@}" <<EXPECT
        spawn -noecho orgalorg {*}\$argv

        expect {
            Password: {
                send "$password\r"
                interact
            } eof {
                send_error "\$expect_out(buffer)"
                exit 1
            }
        }
EXPECT
}

:orgalorg() {
    tests:debug "!!! orgalorg ${@}"

    go-test:run orgalorg "${@}"
}
