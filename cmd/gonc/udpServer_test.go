package main

import (
	"fmt"
	"net"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestUDPMessaging(t *testing.T) {
	var wg sync.WaitGroup

	logger, logBuf := createTestSlog()

	app := &application{
		config: config{verbose: true, hex: true, port: 7000},
		logger: logger,
	}

	ready := make(chan interface{})

	srv := app.NewUDPServer(":7000")
	go func() {
		err := srv.StartUDP()
		assert.NoError(t, err)
		<-ready
		srv.stopUDP()
	}()

	var actualMsg string
	wg.Add(1)
	go func() {
		defer wg.Done()
		time.Sleep(50 * time.Millisecond)

		clientConn, err := net.Dial("udp", ":7000")
		assert.NoError(t, err)

		fmt.Fprintln(clientConn, "Hello from the client")
		time.Sleep(50 * time.Millisecond)
		msg := "Hello from the server\n"
		srv.sendch <- msg

		buf := make([]byte, 1024)
		n, err := clientConn.Read(buf)
		assert.NoError(t, err)
		actualMsg = string(buf[:n])

		time.Sleep(50 * time.Millisecond)
		fmt.Fprintln(clientConn, "2nd msg")

		time.Sleep(50 * time.Millisecond)
		clientConn.Close()
		close(ready)
	}()

	wg.Wait()
	time.Sleep(250 * time.Millisecond)
	expected := `msg="starting UDP server"
msg="received data from the client"
msg="sending message to client"
msg="received data from the client"
`
	assert.Equal(t, expected, logBuf.String())
	expectedMsg := "Hello from the server\n"
	assert.Equal(t, expectedMsg, actualMsg)
}

func TestUDPShutdownOnSendFail(t *testing.T) {
	var wg sync.WaitGroup

	logger, logBuf := createTestSlog()

	app := &application{
		config: config{verbose: true, port: 7001},
		logger: logger,
	}

	srv := app.NewUDPServer(":7001")
	go func() {
		err := srv.StartUDP()
		assert.NoError(t, err)
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()
		time.Sleep(50 * time.Millisecond)

		clientConn, err := net.Dial("udp", ":7001")
		assert.NoError(t, err)

		fmt.Fprintln(clientConn, "Hello from the client")
		clientConn.Close()
		time.Sleep(50 * time.Millisecond)
		msg := "Hello from the server\n"
		srv.sendch <- msg
	}()

	wg.Wait()
	time.Sleep(250 * time.Millisecond)
	expected := `msg="starting UDP server"
msg="received data from the client"
msg="sending message to client"
msg="stopping UDP server"
msg="UDP server shutdown successfully"
`
	assert.Equal(t, expected, logBuf.String())
}

func TestTwoUDPClients(t *testing.T) {
	var wg sync.WaitGroup

	logger, logBuf := createTestSlog()

	app := &application{
		config: config{verbose: true, port: 7003},
		logger: logger,
	}

	srv := app.NewUDPServer(":7003")
	go func() {
		err := srv.StartUDP()
		assert.NoError(t, err)
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()
		time.Sleep(50 * time.Millisecond)
		conn, err := net.Dial("udp", ":7003")
		assert.NoError(t, err)
		fmt.Fprintln(conn, "Hello from the 1st client")
		time.Sleep(500 * time.Millisecond)
		conn.Close()
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()
		time.Sleep(250 * time.Millisecond)
		conn, err := net.Dial("udp", ":7003")
		assert.NoError(t, err)
		fmt.Fprintln(conn, "Hello from the 2nd client")
	}()

	wg.Wait()
	time.Sleep(250 * time.Millisecond)

	expected := `msg="starting UDP server"
msg="received data from the client"
`
	assert.Equal(t, expected, logBuf.String())
}
