RUN = go run -race .
LDFLAGS = -ldflags="-s -w"

all: ci

# run the interpreter
run:
	${RUN}

# run the autoformatter (from system Oak)
fmt:
	oak fmt --changes --fix
f: fmt

# run Go tests
tests:
	go test -race .
t: tests

# run Oak tests
test-oak:
	${RUN} test/main.oak
tk: test-oak

# run oak build tests
test-bundle:
	${RUN} build --entry test/main.oak \
		-o /tmp/oak-test.oak \
		--include std.test:test/std.test,str.test:test/str.test,math.test:test/math.test,sort.test:test/sort.test,random.test:test/random.test,fmt.test:test/fmt.test,json.test:test/json.test,datetime.test:test/datetime.test,path.test:test/path.test,http.test:test/http.test,debug.test:test/debug.test,cli.test:test/cli.test,md.test:test/md.test,crypto.test:test/crypto.test,syntax.test:test/syntax.test
	${RUN} /tmp/oak-test.oak

# run oak build --web tests
test-js:
	${RUN} build --entry test/main.oak \
		-o /tmp/oak-test.js \
		--web \
		--include std.test:test/std.test,str.test:test/str.test,math.test:test/math.test,sort.test:test/sort.test,random.test:test/random.test,fmt.test:test/fmt.test,json.test:test/json.test,datetime.test:test/datetime.test,path.test:test/path.test,http.test:test/http.test,debug.test:test/debug.test,cli.test:test/cli.test,md.test:test/md.test,crypto.test:test/crypto.test,syntax.test:test/syntax.test
	node /tmp/oak-test.js

# build for a specific GOOS target
build-%:
	GOOS=$* go build ${LDFLAGS} -o oak-$* .

# build for all OS targets
build: build-linux build-darwin build-windows build-openbsd

# build Oak sources for the website
site:
	oak build --entry www/src/app.js.oak --output www/static/js/bundle.js --web
	oak build --entry www/src/highlight.js.oak --output www/static/js/highlight.js --web

# build Oak source for the website on file change, using entr
site-w:
	ls www/src/app.js.oak | entr -cr make site

# generate static site pages
site-gen:
	oak www/src/gen.oak

# install as "oak" binary
install:
	cp tools/oak.vim ~/.vim/syntax/oak.vim
	go build ${LDFLAGS} -o ${GOPATH}/bin/oak

# ci in travis
ci: tests test-oak test-bundle
