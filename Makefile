# Build settings.
GOARCH ?= amd64

ifeq ($(OS),Windows_NT) 
PLUGIN_BINARY := ./bin/pgsql-go.dll
COMPILER ?= x86_64-w64-mingw32-gcc # Cross-compiler for Windows
else
PLUGIN_BINARY := ./bin/pgsql-go.so
COMPILER ?= gcc
endif

all: $(PLUGIN_BINARY)

$(PLUGIN_BINARY):
	mkdir -p ./bin
	go build -buildmode c-shared -o $(PLUGIN_BINARY) ./

.PHONY: clean
clean:
	rm -rf ./bin/*

