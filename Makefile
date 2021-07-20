all: run

# run the interpreter
run:
	go run .

# run Go tests
tests:
	go test -v
t: tests

# install as "mgn" binary
install:
	cp tools/magnolia.vim ~/.vim/syntax/magnolia.vim
	go build -o ${GOPATH}/bin/mgn
