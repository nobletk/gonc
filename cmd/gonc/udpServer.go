package main

import (
	"encoding/hex"
	"errors"
	"fmt"
	"log/slog"
	"net"
	"os"
	"os/signal"
	"syscall"
)

type UDPServer struct {
	bytesRcvd int
	bytesSent int
	config    config
	conn      *net.UDPConn
	lAddr     *net.UDPAddr
	rAddr     *net.UDPAddr
	quit      chan interface{}
	sendch    chan string
	logger    *slog.Logger
}

func (app *application) NewUDPServer(addr string) *UDPServer {
	lAddr, err := net.ResolveUDPAddr("udp", addr)
	if err != nil {
		app.logger.Error("failed to resolve local UDP address", "addr", addr, "error", err)
		os.Exit(1)
	}

	return &UDPServer{
		config: app.config,
		lAddr:  lAddr,
		logger: app.logger,
		quit:   make(chan interface{}),
		sendch: make(chan string),
	}
}

func (srv *UDPServer) StartUDP() error {
	ln, err := net.ListenUDP("udp", srv.lAddr)
	if err != nil {
		return err
	}
	srv.logger.Info("starting UDP server", "addr", srv.lAddr)

	if srv.config.verbose {
		fmt.Printf("Listening on [any] %d ...\n", srv.config.port)
	}

	go srv.stopOsSignal()

	rAddr, err := srv.getRemoteAddr(ln)
	if err != nil {
		return err
	}

	srv.rAddr = rAddr
	ln.Close()

	go srv.handleUDPConnection()

	<-srv.quit
	srv.logger.Info("UDP server shutdown successfully")
	return nil
}

func (srv *UDPServer) stopUDP() {
	if srv.config.verbose {
		fmt.Printf(" sent %d, rcvd %d\n", srv.bytesSent, srv.bytesRcvd)
	}
	srv.logger.Info("stopping UDP server")
	close(srv.quit)
	close(srv.sendch)
	if srv.conn != nil {
		srv.conn.Close()
	}
}

func (srv *UDPServer) handleUDPConnection() {
	conn, err := net.DialUDP("udp", srv.lAddr, srv.rAddr)
	if err != nil {
		srv.logger.Error("failed to dial UDP connection", "lAddr", srv.lAddr.String(), "rAddr", srv.rAddr.String(), "error", err)
		return
	}
	srv.conn = conn

	go srv.readUDP(conn)
	go srv.writeUDP(conn)
}

func (srv *UDPServer) getRemoteAddr(conn *net.UDPConn) (*net.UDPAddr, error) {
	buf := make([]byte, 2048)
	var dataRead []byte

	n, rAddr, err := conn.ReadFromUDP(buf)
	if err != nil {
		return nil, err
	}
	srv.bytesRcvd += n
	dataRead = buf[:n]
	if srv.config.verbose {
		fmt.Printf("Connection to [%s] from [%s] [%s]\n", conn.LocalAddr(), rAddr, rAddr.Network())
	}
	srv.logger.Info("received data from the client", "addr", rAddr, "byte", n)
	fmt.Print(string(dataRead))

	return rAddr, nil
}

func (srv *UDPServer) readUDP(conn *net.UDPConn) {
	buf := make([]byte, 2048)
	var dataRead []byte

	for {
		n, rAddr, err := conn.ReadFromUDP(buf)
		if err != nil {
			if errors.Is(err, syscall.ECONNREFUSED) {
				fmt.Print("Connection refused: ")
				srv.stopUDP()
				return
			}
			srv.logger.Error("failed to read from UDP connection", "rAddr", rAddr, "error", err)
			return
		}
		srv.bytesRcvd += n
		dataRead = buf[:n]
		srv.logger.Info("received data from the client", "addr", rAddr, "byte", n)
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

func (srv *UDPServer) writeUDP(conn *net.UDPConn) {
	for msg := range srv.sendch {
		n, err := conn.Write([]byte(msg))
		if err != nil {
			if errors.Is(err, syscall.ECONNREFUSED) {
				fmt.Print("Connection refused: ")
				srv.stopUDP()
				return
			}
			srv.logger.Error("failed to write to UDP connection", "rAddr", conn.RemoteAddr(), "error", err)
			return
		}
		srv.bytesSent += n
		srv.logger.Info("sending message to client", "remoteAddr", conn.RemoteAddr())
		if srv.config.hex {
			fmt.Printf("Sent %d bytes to the socket\n", n)
			fmt.Printf("%s", hex.Dump([]byte(msg)))
		}
	}
}

func (srv *UDPServer) stopOsSignal() {
	sigch := make(chan os.Signal, 1)
	signal.Notify(sigch, syscall.SIGINT, syscall.SIGTERM)
	s := <-sigch
	srv.logger.Info("received operating system signal", "sig", s)
	srv.stopUDP()
}
