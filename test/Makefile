.POSIX:

default: test

test:
	@if command -v grc >/dev/null 2>&1; then \
		grc go test -short -v -timeout 30m; \
	else \
		go test -short -v -timeout 30m; \
	fi
