TARGET=uWSGI_expoter

all: main.go
	@echo "Building uWSGI_expoter"
	@go build -ldflags "-X main.VERSION_BUILD_TIME=`date -u '+%Y-%m-%d_%I:%M:%S%p'` -X main.VERSION_BUILD_GIT_HASH=`git rev-parse HEAD` -X main.VERSION_BUILD_GIT_VERSION=`git describe --abbrev=0 --tags`" -o $(TARGET)
clean:
	@go clean
	@rm -rfv $(TARGET)
