:orgalorg:lock() {
    local _output_var="$1"
    local _pid_var="$2"
    shift 2

    local _orgalorg_output="$(tests:get-tmp-dir)/oralorg.stdout"
    local _orgalorg=""

    tests:run-background _orgalorg \
        tests:silence tests:pipe \
            :orgalorg:with-key --lock "${@}" '2>&1' \
                '|' tee "$_orgalorg_output"

    while ! grep -qF "waiting for interrupt" "$_orgalorg_output" 2>/dev/null
    do
        tests:debug "[orgalorg] waiting for global lock..."
        sleep 0.1
    done

    tests:debug "[orgalorg] global lock has been acquired"

    eval $_output_var=\$_orgalorg_output
    eval $_pid_var=\$\(tests:get-background-pid \$_orgalorg\)
}
