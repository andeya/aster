
# NOTE: This Makefile is only necessary if you 
# plan on developing the msgp tool and library.
# Installation can still be performed with a
# normal `go install`.

# Test output directory
OUT = ./_out

.PHONY: clean wipe install get-deps bench all gomod

$(BIN): */*.go
	@go install ./...

all: install test

install: gomod
	go install ./...

test:
	go test -v ./_examples
	go test -v ./_test
	go test -v ./...

bench:
	go test -bench ./_examples
	go test -bench ./_test
	go test -bench ./...

clean:
	rm -rf $(OUT)

wipe: clean
	$(RM) $(BIN)

get-deps:
	go get -d -t ./...

coverage:
	./coverage.sh