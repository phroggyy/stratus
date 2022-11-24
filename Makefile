dev:
	npx nodemon --exec "make build" -e "go mod"
build:
	go build .