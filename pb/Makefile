# Generate the Go files from the dns.proto protobuf, you need the utilities
# from: https://github.com/golang/protobuf to make this work.
# The generate dns.pb.go is checked into git, so for normal builds we don't need
# to run this generation step.

PROTOC_VERSION:=3.9.1
OS            :=
ARCH          :=
ifeq ($(OS),Windows_NT)
	OS=win
	ifeq ($(PROCESSOR_ARCHITECTURE),AMD64)
		ARCH=64
	endif
	ifeq ($(PROCESSOR_ARCHITECTURE),x86)
		ARCH=32
	endif
else
	UNAME_S := $(shell uname -s)
	ifeq ($(UNAME_S),Linux)
		OS=linux
	endif
	ifeq ($(UNAME_S),Darwin)
		OS=osx
	endif
	UNAME_P := $(shell uname -p)
	ifeq ($(UNAME_P),x86_64)
		ARCH=x86_64
	endif
	ifneq ($(filter %86,$(UNAME_P)),)
		ARCH=x86_32
	endif
endif
PROTOC_ZIP:=protoc-$(PROTOC_VERSION)-$(OS)-$(ARCH).zip

all: dns.pb.go

dns.pb.go: dns.proto
	protoc --go_out=plugins=grpc:. dns.proto

.PHONY: clean
clean:
	rm dns.pb.go

.PHONY: buildDeps
buildDeps:
	curl -OL https://github.com/google/protobuf/releases/download/v$(PROTOC_VERSION)/$(PROTOC_ZIP)
	mkdir -p $(HOME)/go/bin
	unzip -o $(PROTOC_ZIP) -d $(HOME)/go bin/protoc
	rm -f $(PROTOC_ZIP)
