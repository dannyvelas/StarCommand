BINARY := stc
CMD     := ./cmd/stc

.PHONY: all build clean

all: build

build:
	go build -o ./$(BINARY) $(CMD)

clean:
	rm -f ./$(BINARY)
