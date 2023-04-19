.PHONY: all server clean

all: server

server:
	go build -o httpserver

clean:
	rm -f httpserver
