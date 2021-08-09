# Magnolia

[![Build Status](https://travis-ci.com/thesephist/magnolia.svg?branch=main)](https://travis-ci.com/thesephist/magnolia)

A friendly, expressive programming language. For a detailed description of the language, see the [language spec](docs/spec.md).

## Development

Magnolia (ab)uses GNU Make to run development workflows and tasks.

- `make run` compiles and runs the Magnolia binary, which opens an interactive REPL
- `make tests` or `make t` runs the Go tes suite for the Magnolia language and interpreter
- `make test-mgn` or `make tm` runs the Magnolia test suite, which tests the standard libraries
- `make install` installs the Mangolia interpreter on your `$GOPATH` as `mgn`, and re-installs Mgn's vim syntax file

