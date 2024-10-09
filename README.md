# GoNC

A Command line networking utility tool inspired by netcat written in Go. It
reads and writes data across network connections, using the TCP/UDP protocols.
It is a debugging and exploration tool. It also functions as a server, by
listening for inbound connections on arbitrary ports and then doing the same
reading and writing.

## Features

* Read and Write data using TCP/UDP protocols.

* Function as a TCP/UDP Server by listening for inbound connections. 

* Debug mode for debugging.

* Scanning for port or range of ports on a hostname.

* Execute specified process, piping the input and output to and from the client.

## Usage

```
gonc [-options] hostname port[s] [ports] ...
gonc -l -p port [-options] [hostname] [port]
```

The options are the following:

* `-l` or `--listenMode` : listen mode for inbound connections

* `-p` or `--port` : local port number

* `-u` or `--udp` : UDP mode

* `-d` or `--debug` : debug mode for logs

```
gonc -d -l -p 8888 
time=2024-10-09T22:10:00.957+02:00 level=INFO msg="starting TCP server" addr=:8888
time=2024-10-09T22:10:04.760+02:00 level=INFO msg="connected to" remoteAddr=127.0.0.1:52168
time=2024-10-09T22:10:12.634+02:00 level=INFO msg="received data" remoteAddr=127.0.0.1:52168 bytes=13
hello server
hi client
time=2024-10-09T22:10:18.169+02:00 level=INFO msg="message sent to client" remoteAddr=127.0.0.1:52168
time=2024-10-09T22:10:21.888+02:00 level=INFO msg="client disconnected" remoteAddr=127.0.0.1:52168
time=2024-10-09T22:10:21.888+02:00 level=INFO msg="stopping TCP connection"
time=2024-10-09T22:10:21.888+02:00 level=INFO msg="TCP server shutdown successfully"
```

* `-v` or `--verbose` : verbose mode

```
gonc -v -l -p 8888 
Listening on [any] 8888...
Connection to [127.0.0.1:8888] from [127.0.0.1:52168] [tcp]
hello server
hi client
sent 10, rcvd 13
```

* `-e` or `--exec` : program to exec after connect

```
# listen on localhost port 8888 executing /bin/bash
gonc -l -p 8888 -e /bin/bash
# then from a new connection connect to the previous server
gonc localhost 8888
echo "Hello!"
```

* `-x` or `--hex` : hex dumping mode

```
gonc -v -x -l -p 8888
Listening on [any] 8888...
Connection to [127.0.0.1:8888] from [127.0.0.1:42374] [tcp]
Hi from the client!
Received 20 bytes from the socket
00000000  48 69 20 66 72 6f 6d 20  74 68 65 20 63 6c 69 65  |Hi from the clie|
00000010  6e 74 21 0a                                       |nt!.|
Hey from the server!
Sent 21 bytes to the socket
00000000  48 65 79 20 66 72 6f 6d  20 74 68 65 20 73 65 72  |Hey from the ser|
00000010  76 65 72 21 0a                                    |ver!.|
sent 21, rcvd 20
```

* `-z` or `--zero` : zero-I/O mode [used for scanning]

```
gonc -v -d -z localhost 8888
time=2024-10-09T22:19:32.265+02:00 level=INFO msg="Connection to localhost 127.0.0.1:8888 [tcp]\n"
Connection to localhost 127.0.0.1:8888 [tcp]
```

## Getting started

### Clone the repo

```shell
git clone https://github.com/nobletk/gonc
# then build the binary
make build
```

### Go
```shell
go install https://github.com/nobletk/gonetcat/cmd/gonc@latest
```

