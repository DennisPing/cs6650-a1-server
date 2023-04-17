.PHONY: all server clean

all: server

server:
	go build -o ./server

clean:
	rm -f ./server
