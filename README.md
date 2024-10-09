# GoNC

A Command line networking utility tool inspired by netcat written in Go. It 

reads and writes data across network connections, using the TCP/UDP protocols.

It is a debugging and exploration tool. It also functions as a server, by 

listening for inbound connections on arbitrary ports and then doing the same

reading and writing.

## Features

* Read and Write data using TCP/UDP protocols.

* Function as a TCP/UDP Server by listening for inbound connections. 

## Usage

```
gonc [-options] hostname port[s] [ports] ...
gonc -l -p port [-options] [hostname] [port]
```

The options are the following:

* `-d` or `--debug` : debug mode for logs

* `-e` or `--exec` : program to exec after connect

* `-l` or `--listenMode` : listen mode for inbound connections

* `-p` or `--port` : local port number

* `-u` or `--udp` : UDP mode

* `-v` or `--verbose` : verbose mode

* `-x` or `--hex` : hex dumping mode

* `-z` or `--zero` : zero-I/O mode [used for scanning]

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

