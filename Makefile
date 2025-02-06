BINARY_NAME=git-download
VERSION=$(shell git describe --tags --always)
BUILD_DIR=dist
LDFLAGS=-ldflags "-s -w -X main.version=${VERSION}"

.PHONY: all windows linux darwin clean

all: windows linux darwin

windows:
	mkdir -p ${BUILD_DIR}
	GOOS=windows GOARCH=amd64 go build ${LDFLAGS} -o ${BUILD_DIR}/${BINARY_NAME}_windows_amd64.exe ./cmd/git-download
	GOOS=windows GOARCH=arm64 go build ${LDFLAGS} -o ${BUILD_DIR}/${BINARY_NAME}_windows_arm64.exe ./cmd/git-download

linux:
	mkdir -p ${BUILD_DIR}
	GOOS=linux GOARCH=amd64 go build ${LDFLAGS} -o ${BUILD_DIR}/${BINARY_NAME}_linux_amd64 ./cmd/git-download
	GOOS=linux GOARCH=arm64 go build ${LDFLAGS} -o ${BUILD_DIR}/${BINARY_NAME}_linux_arm64 ./cmd/git-download

darwin:
	mkdir -p ${BUILD_DIR}
	GOOS=darwin GOARCH=amd64 go build ${LDFLAGS} -o ${BUILD_DIR}/${BINARY_NAME}_darwin_amd64 ./cmd/git-download
	GOOS=darwin GOARCH=arm64 go build ${LDFLAGS} -o ${BUILD_DIR}/${BINARY_NAME}_darwin_arm64 ./cmd/git-download

clean:
	rm -rf ${BUILD_DIR}

# Create release archives
release: all
	cd ${BUILD_DIR} && \
	zip ${BINARY_NAME}_windows_amd64.zip ${BINARY_NAME}_windows_amd64.exe && \
	zip ${BINARY_NAME}_windows_arm64.zip ${BINARY_NAME}_windows_arm64.exe && \
	tar czf ${BINARY_NAME}_linux_amd64.tar.gz ${BINARY_NAME}_linux_amd64 && \
	tar czf ${BINARY_NAME}_linux_arm64.tar.gz ${BINARY_NAME}_linux_arm64 && \
	tar czf ${BINARY_NAME}_darwin_amd64.tar.gz ${BINARY_NAME}_darwin_amd64 && \
	tar czf ${BINARY_NAME}_darwin_arm64.tar.gz ${BINARY_NAME}_darwin_arm64

# Install locally
install: clean
	go build ${LDFLAGS} -o ${BINARY_NAME} ./cmd/git-download
	mv ${BINARY_NAME} /usr/local/bin/

# Run tests
test:
	go test -v ./... 