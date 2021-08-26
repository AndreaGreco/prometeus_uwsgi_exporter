TARGET=uWSGI_exporter

all: main.go
	@echo "Building uWSGI_exporter"
	@go build -ldflags "-X main.VERSION_BUILD_TIME=`date -u '+%Y-%m-%d_%I:%M:%S%p'` -X main.VERSION_BUILD_GIT_HASH=`git rev-parse HEAD` -X main.VERSION_BUILD_GIT_VERSION=`git describe --abbrev=0 --tags`" -o $(TARGET)
clean:
	@go clean
	@rm -rfv $(TARGET)
