tests:involve tests/containers.sh
tests:involve tests/hastur.sh
tests:involve tests/sudo.sh

for (( i = $_containers_count; i > 0; i-- )) {
    :containers:spawn
}

:containers:list-to-var containers

echo ${#containers}
