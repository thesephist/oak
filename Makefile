all: run

# run the interpreter
run:
	go run .

# run Go tests
test:
	go test -v
t: test

# install as "mgn" binary
install:
	cp tools/magnolia.vim ~/.vim/syntax/magnolia.vim
	go build -o ${GOPATH}/bin/mgn
