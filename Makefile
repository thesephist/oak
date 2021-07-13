all: run

# run the interpreter
run:
	go run .

# run Go tests
test:
	go test -v
t: test
