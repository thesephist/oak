all: ci

# run the interpreter
run:
	go run .

# run Go tests
tests:
	go test .
t: tests

# run Oak tests
test-oak:
	go run -race . test/main.oak
tk: test-oak

# install as "oak" binary
install:
	cp tools/oak.vim ~/.vim/syntax/oak.vim
	go build -o ${GOPATH}/bin/oak

# ci in travis
ci: tests test-oak
