:deps:check() {
    if ! which brctl >/dev/null 2>&1; then
        echo "missing dependency: brctl (bridge-utils)"
        exit 1
    fi >&2

    if ! which expect >/dev/null 2>&1; then
        echo "missing dependency: expect"
        exit 1
    fi >&2

    if ! which hastur >/dev/null 2>&1; then
        echo "missing dependency: hastur (https://github.com/seletskiy/hastur)"
        exit 1
    fi >&2
}
