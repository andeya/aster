
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
	go test -v ./...
	go test -v ./_examples

bench:
	go test -bench ./...
	go test -bench ./_examples

clean:
	rm -rf $(OUT)

wipe: clean
	$(RM) $(BIN)

get-deps:
	go get -d -t ./...
