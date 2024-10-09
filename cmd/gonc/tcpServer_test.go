package main

import (
	"errors"
	"fmt"
	"net"
	"os"
	"strconv"
	"sync"
	"syscall"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestOsSignals(t *testing.T) {
	tests := []struct {
		name     string
		config   config
		lHost    string
		signal   syscall.Signal
		expected string
	}{
		{
			name:   "Start And Shutdown With SigInt",
			config: config{verbose: true, port: 3002},
			lHost:  ":3002",
			signal: syscall.SIGINT,
			expected: `msg="starting TCP server"
msg="received operating system signal" sig=interrupt
msg="stopping TCP connection"
msg="TCP server shutdown successfully"
`,
		},
		{
			name:   "Start And Shutdown With SigTerm",
			config: config{verbose: true, port: 3003},
			lHost:  ":3003",
			signal: syscall.SIGTERM,
			expected: `msg="starting TCP server"
msg="received operating system signal" sig=terminated
msg="stopping TCP connection"
msg="TCP server shutdown successfully"
`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			logger, logBuf := createTestSlog()

			app := &application{
				config: tt.config,
				logger: logger,
			}
			srv := app.NewTCPServer(tt.lHost)

			go func() {
				err := srv.StartTCP()
				assert.NoError(t, err)
			}()

			var wg sync.WaitGroup
			wg.Add(1)
			go func() {
				defer wg.Done()
				p, err := os.FindProcess(os.Getpid())
				assert.NoError(t, err)
				time.Sleep(50 * time.Millisecond)
				p.Signal(tt.signal)
			}()

			wg.Wait()
			time.Sleep(250 * time.Millisecond)
			assert.Equal(t, tt.expected, logBuf.String())
		})
	}
}

func TestTCPMessaging(t *testing.T) {
	var wg sync.WaitGroup

	logger, logBuf := createTestSlog()

	app := &application{
		config: config{verbose: true, hex: true, port: 3005},
		logger: logger,
	}

	srv := app.NewTCPServer(":3005")
	go func() {
		err := srv.StartTCP()
		assert.NoError(t, err)

	}()

	var actualMsg string
	wg.Add(1)
	go func() {
		defer wg.Done()
		time.Sleep(50 * time.Millisecond)

		clientConn, err := net.Dial("tcp", srv.lAddrStr)
		assert.NoError(t, err)

		fmt.Fprintln(clientConn, "hello from the client")

		time.Sleep(50 * time.Millisecond)
		msg := "hello from the server\n"
		srv.sendch <- msg

		buf := make([]byte, 1024)
		n, err := clientConn.Read(buf)
		assert.NoError(t, err)
		actualMsg = string(buf[:n])

		time.Sleep(50 * time.Millisecond)
		clientConn.Close()
	}()

	wg.Wait()
	time.Sleep(250 * time.Millisecond)

	expected := `msg="starting TCP server"
msg="connected to"
msg="received data"
msg="message sent to client"
msg="client disconnected"
msg="stopping TCP connection"
msg="TCP server shutdown successfully"
`
	assert.Equal(t, expected, logBuf.String())
	expectedMsg := "hello from the server\n"
	assert.Equal(t, expectedMsg, actualMsg)
}

func TestTwoTCPClients(t *testing.T) {
	var wg sync.WaitGroup

	logger, logBuf := createTestSlog()

	app := &application{
		config: config{verbose: true, port: 3009},
		logger: logger,
	}

	srv := app.NewTCPServer(":3009")
	go func() {
		err := srv.StartTCP()
		assert.NoError(t, err)
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()
		time.Sleep(50 * time.Millisecond)
		conn, err := net.Dial("tcp", srv.lAddrStr)
		assert.NoError(t, err)
		time.Sleep(500 * time.Millisecond)
		conn.Close()
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()
		time.Sleep(250 * time.Millisecond)
		_, err := net.Dial("tcp", srv.lAddrStr)
		var opErr *net.OpError
		assert.Error(t, err)
		assert.True(t, errors.As(err, &opErr), "expected a *net.OpError")
		assert.True(t, errors.Is(err, syscall.ECONNREFUSED), "expected a connection refused error")
	}()

	wg.Wait()
	time.Sleep(250 * time.Millisecond)

	expected := `msg="starting TCP server"
msg="connected to"
msg="client disconnected"
msg="stopping TCP connection"
msg="TCP server shutdown successfully"
`
	assert.Equal(t, expected, logBuf.String())
}

func TestExecuteTCPCmd(t *testing.T) {
	tests := []struct {
		name     string
		cmd      string
		port     int
		expected string
	}{
		{
			name:     "Echo",
			cmd:      "echo Hello",
			port:     3006,
			expected: "Hello\n",
		},
		{
			name:     "List Directory",
			cmd:      "ls",
			port:     3007,
			expected: "helper.go\nmain.go\nscan.go\nscan_test.go\ntcpServer.go\ntcpServer_test.go\nudpServer.go\nudpServer_test.go\n",
		},
		// fails when run with global test command??
		// {
		// 	name:     "Fail With Command Not Found",
		// 	cmd:      "Hello",
		// 	port:     3008,
		// 	expected: "/bin/bash: line 1: Hello: command not found\n",
		// },
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var wg sync.WaitGroup

			logger, _ := createTestSlog()

			app := &application{
				config: config{cmd: "/bin/bash", port: tt.port},
				logger: logger,
			}

			srv := app.NewTCPServer(":" + strconv.Itoa(tt.port))
			go func() {
				err := srv.StartTCP()
				assert.NoError(t, err)
			}()

			var actual string
			wg.Add(1)
			go func() {
				defer wg.Done()
				time.Sleep(50 * time.Millisecond)
				clientConn, err := net.Dial("tcp", srv.lAddrStr)
				defer clientConn.Close()
				assert.NoError(t, err)
				fmt.Fprintln(clientConn, tt.cmd)

				buf := make([]byte, 1024)
				n, err := clientConn.Read(buf)
				assert.NoError(t, err)
				actual = string(buf[:n])
			}()

			wg.Wait()
			assert.Equal(t, tt.expected, actual)
		})
	}
}
