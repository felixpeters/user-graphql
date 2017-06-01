test:
	go test ./...

compile:
	go build -o bin/user-graphql

build-image:
	@echo "### Compiling"
	GOOS=linux go build -o bin/user-graphql
	@echo "### Building image"
	docker build -t user-graphql .

run-docker: build-image
	docker run -it -p 8080:8080 user-graphql
