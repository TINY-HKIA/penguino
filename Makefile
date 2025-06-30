build:
	@go build -o bin/penguino

run: build
	@./bin/penguino
