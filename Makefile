vendor/%/Makefile:
	bash -c 'source vendor/github.com/reconquest/import.bash/import.bash && \
		import "$*"'

include vendor/github.com/reconquest/test-runner.bash/Makefile
include vendor/github.com/reconquest/go-test.bash/Makefile
