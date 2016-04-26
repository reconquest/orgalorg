if ! type tests:debug &>/dev/null; then
    tests:debug() {
        echo "${@}"
    }
fi

function sudo() {
    tests:debug $(echo -e "\e[1;31m{SUDO} $ ${@}\e[0m")

    command sudo -n "${@}"
}
