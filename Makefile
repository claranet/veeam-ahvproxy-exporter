
BIN_NAME = veeam-ahvproxy-exporter
DOCKER_IMAGE_NAME ?= veeam-ahvproxy-exporter
export GOPATH = ${PWD}
export CGO_ENABLED = 0
export GOBUILD_ARGS = -a -tags netgo -ldflags -w
#export GOARCH ?= amd64
#export GOOS ?= linux

all: linux windows docker

linux: prepare
	$(eval export GOOS=linux)
	go build $(GOBUILD_ARGS) -o ./bin/$(BIN_NAME)
	zip ./bin/$(BIN_NAME)-$(GOOS)-$(GOARCH).zip ./bin/$(BIN_NAME)
	rm ./bin/$(BIN_NAME)

clean:
	@echo "Clean up"
	go clean
	rm -rf bin/ src/

docker:
	@echo ">> Compile using docker container"
	@docker build -t "$(DOCKER_IMAGE_NAME)" .

windows: prepare
	$(eval export GOOS=windows)
	go build $(GOBUILD_ARGS) -o ./bin/$(BIN_NAME).exe
	zip ./bin/$(BIN_NAME)-$(GOOS)-$(GOARCH).zip ./bin/$(BIN_NAME).exe
	rm ./bin/$(BIN_NAME).exe
	
prepare:	
	@echo "Create output directory ./bin/"
	go env
	mkdir -p bin/
	@echo "GO get dependencies"
	go get -d
	

.PHONY: all
