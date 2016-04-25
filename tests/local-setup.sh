tests:involve deps.sh
tests:involve sudo.sh
tests:involve hastur.sh
tests:involve getopt.sh
tests:involve containers.sh

for (( i = $(:containers:count); i > 0; i-- )) {
    :containers:spawn
}

:containers:list-to-var containers

echo ${#containers}
