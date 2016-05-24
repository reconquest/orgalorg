:build:init() {
    printf "[build] building go binary... "

    if build_out=$(go build -o orgalorg -v 2>&1 | tee /dev/stderr); then
        printf "ok.\n"
    else
        printf "fail.\n\n%s\n" "$build_out"
        return 1
    fi
}
