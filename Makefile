TARGET=uWSGI_expoter

all: main.go
	go build -o $(TARGET)

clean:
	rm $(TARGET)
