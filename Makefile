all: run

# run the interpreter
run:
	go run .

# run Go tests
tests:
	go test -v
t: tests

# run Magnolia tests
test-mgn:
	go build -race -o ./mgn
	./mgn test/std.mgn
	rm ./mgn
tm: test-mgn

# install as "mgn" binary
install:
	cp tools/magnolia.vim ~/.vim/syntax/magnolia.vim
	go build -o ${GOPATH}/bin/mgn
