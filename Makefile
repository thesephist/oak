all: ci

# run the interpreter
run:
	go run .

# run the autoformatter (from system Oak)
fmt:
	oak fmt --changes --fix
f: fmt

# run Go tests
tests:
	go test .
t: tests

# run Oak tests
test-oak:
	go run -race . test/main.oak
tk: test-oak

# run Oak build --web tests
test-js:
	go run . build --entry test/main.oak \
		-o /tmp/oak-test-main.js \
		--web \
		--include std.test:test/std.test;
	node /tmp/oak-test-main.js

# install as "oak" binary
install:
	cp tools/oak.vim ~/.vim/syntax/oak.vim
	go build -o ${GOPATH}/bin/oak

# ci in travis
ci: tests test-oak
