all: run

# run the interpreter
run:
	go run .

# run Go tests
tests:
	go test .
t: tests

# run Magnolia tests
test-mgn:
	go run -race . test/main.mgn
tm: test-mgn

# install as "mgn" binary
install:
	cp tools/magnolia.vim ~/.vim/syntax/magnolia.vim
	go build -o ${GOPATH}/bin/mgn
