.PHONY: .gitignore

.gitignore:
	find . -type d -name .git -prune -printf '/%P\n' \
		| sed 's#/\.git$$##' \
		| sort \
		| tee .gitignore
