build:
	go build -ldflags "-X 'main.buildTime=$(shell date)' -X 'main.version=v.0.0.1-$(shell git log --pretty=format:"%h" | head -n 1)-$(shell date -u +%Y%m%d-%H%M%S)'"
