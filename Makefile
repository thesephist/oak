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

# run oak build tests
test-bundle:
	go run . build --entry test/main.oak \
		-o /tmp/oak-test.oak \
		--include std.test:test/std.test,str.test:test/str.test,math.test:test/math.test,sort.test:test/sort.test,random.test:test/random.test,fmt.test:test/fmt.test,json.test:test/json.test,datetime.test:test/datetime.test,path.test:test/path.test,cli.test:test/cli.test,md.test:test/md.test,crypto.test:test/crypto.test,syntax.test:test/syntax.test
	go run . /tmp/oak-test.oak

# run oak build --web tests
test-js:
	go run . build --entry test/main.oak \
		-o /tmp/oak-test.js \
		--web \
		--include std.test:test/std.test,str.test:test/str.test,math.test:test/math.test,sort.test:test/sort.test,random.test:test/random.test,fmt.test:test/fmt.test,json.test:test/json.test,datetime.test:test/datetime.test,path.test:test/path.test,cli.test:test/cli.test,md.test:test/md.test,crypto.test:test/crypto.test,syntax.test:test/syntax.test
	node /tmp/oak-test.js

# install as "oak" binary
install:
	cp tools/oak.vim ~/.vim/syntax/oak.vim
	go build -o ${GOPATH}/bin/oak

# ci in travis
ci: tests test-oak test-bundle
