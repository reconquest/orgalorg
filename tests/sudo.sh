function sudo() {
    tests:debug $(echo -e "\e[1;31m{SUDO} $ ${@}\e[0m")

    command sudo -n "${@}"
}
