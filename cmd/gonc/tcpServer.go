package main

import (
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net"
	"os"
	"os/exec"
	"os/signal"
	"syscall"
)

type TCPServer struct {
	bytesRcvd int
	bytesSent int
	config    config
	conn      net.Conn
	lAddrStr  string
	ln        net.Listener
	logger    *slog.Logger
	quit      chan interface{}
	sendch    chan string
}

func (app *application) NewTCPServer(addr string) *TCPServer {
	return &TCPServer{
		config:   app.config,
		lAddrStr: addr,
		logger:   app.logger,
		quit:     make(chan interface{}),
		sendch:   make(chan string),
	}
}

func (srv *TCPServer) StartTCP() error {
	ln, err := net.Listen("tcp", srv.lAddrStr)
	if err != nil {
		return err
	}
	srv.ln = ln

	srv.logger.Info("starting TCP server", "addr", srv.lAddrStr)
	if srv.config.verbose {
		fmt.Printf("Listening on [any] %d...\n", srv.config.port)
	}

	go srv.acceptTCP()

	go func() {
		sigch := make(chan os.Signal, 1)
		signal.Notify(sigch, syscall.SIGINT, syscall.SIGTERM)
		s := <-sigch
		srv.logger.Info("received operating system signal", "sig", s)
		srv.stopTCP()
	}()

	<-srv.quit
	srv.logger.Info("TCP server shutdown successfully")
	return nil
}

func (srv *TCPServer) stopTCP() {
	if srv.config.verbose && srv.config.cmd == "" {
		fmt.Printf("sent %d, rcvd %d\n", srv.bytesSent, srv.bytesRcvd)
	}
	srv.logger.Info("stopping TCP connection")
	close(srv.sendch)
	close(srv.quit)
	if srv.conn != nil {
		srv.conn.Close()
	}
}

func (srv *TCPServer) acceptTCP() {
	conn, err := srv.ln.Accept()
	if err != nil {
		select {
		case <-srv.quit:
			return
		default:
			srv.logger.Error("failed to accept listener", "error", err)
		}
	}
	srv.ln.Close()
	srv.conn = conn

	srv.logger.Info("connected to", "remoteAddr", conn.RemoteAddr())

	if srv.config.verbose {
		fmt.Printf("Connection to [%s] from [%s] [%s]\n", conn.LocalAddr(), conn.RemoteAddr(), conn.RemoteAddr().Network())
	}

	if cmd := srv.config.cmd; cmd != "" {
		srv.executeTCPCmd(conn, cmd)
	}
	go srv.readTCP(conn)
	go srv.writeTCP(conn)
}

func (srv *TCPServer) readTCP(conn net.Conn) {
	buf := make([]byte, 2048)
	var dataRead []byte

	for {
		n, err := conn.Read(buf)
		if err != nil {
			if errors.Is(err, io.EOF) || errors.Is(err, syscall.ECONNRESET) {
				srv.logger.Info("client disconnected", "remoteAddr", conn.RemoteAddr())
				srv.stopTCP()
				return
			}

			select {
			case <-srv.quit:
				return
			default:
				srv.logger.Error("failed to read from tcp connection", "error", err)
				return
			}
		}
		srv.bytesRcvd += n
		dataRead = buf[:n]
		srv.logger.Info("received data", "remoteAddr", conn.RemoteAddr(), "bytes", n)
		fmt.Print(string(dataRead))

		if srv.config.hex {
			fmt.Printf("Received %d bytes from the socket\n", n)
			fmt.Printf("%s", hex.Dump(dataRead))
		}

		if n == 0 {
			return
		}
	}
}

func (srv *TCPServer) writeTCP(conn net.Conn) {
	for msg := range srv.sendch {
		n, err := conn.Write([]byte(msg))
		if err != nil {
			srv.logger.Error("failed to write to tcp connection", "error", err)
			return
		}
		srv.bytesSent += n
		srv.logger.Info("message sent to client", "remoteAddr", conn.RemoteAddr())

		if srv.config.hex {
			fmt.Printf("Sent %d bytes to the socket\n", n)
			fmt.Printf("%s", hex.Dump([]byte(msg)))
		}
	}
}

func (srv *TCPServer) executeTCPCmd(conn net.Conn, cmd string) {
	c := exec.Command(cmd)
	c.Stdin = conn
	c.Stdout = conn
	c.Stderr = conn

	if err := c.Run(); err != nil {
		srv.logger.Error("failed to run command", "error", err)
		os.Exit(1)
	}
}
