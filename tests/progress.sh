export _progress_fifo=${_progress_fifo:-""}

:progress:start() {
    _progress_fifo=$(mktemp -u)

    mkfifo $_progress_fifo

    coproc _progress {
        local indicator='|/-\'
        local position=0

        echo end
        while read line; do
            echo -n "${indicator:$position:1}" >&2
            echo -ne '\b' >&2
            position=$(( ($position + 1) % ${#indicator} ))
        done < $_progress_fifo
    }

    trap "{ rm $_progress_fifo; }" EXIT
}

:progress:step() {
    tee $_progress_fifo
}
