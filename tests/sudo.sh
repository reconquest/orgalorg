if ! type tests:debug &>/dev/null; then
    tests:debug() {
        echo "${@}"
    }
fi

function :sudo() {
    {
        printf "\e[1;31m{sudo} $ %s\e[0m\n" "$1"
        printf "\e[1;31m       .  %s\e[0m\n" "${@:2}"
    } | _tests_indent

    :sudo:silent "${@}"
}

function :sudo:silent() {
    command sudo -n "${@}"
}
