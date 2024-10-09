package main

import (
	"net"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestScanConnection(t *testing.T) {
	tests := []struct {
		name     string
		port     string
		lAddr    string
		expected string
	}{
		{
			name:     "Success One Port",
			port:     "8000",
			lAddr:    "localhost:8000",
			expected: "msg=\"Connection to localhost 127.0.0.1:8000 [tcp]\\n\"\n",
		},
		{
			name:     "Success Port Range",
			port:     "8000-9000",
			lAddr:    "localhost:8500",
			expected: "msg=\"Connection to localhost 127.0.0.1:8500 [tcp]\\n\"\n",
		},
		{
			name:     "Failure Port Out Of Range",
			port:     "8000-9000",
			lAddr:    "localhost:9200",
			expected: "msg=\"Could not connect to localhost:8000-9000\\n\"\n",
		},
		{
			name:     "Starting Port Bigger Than Ending Port",
			port:     "9000-8000",
			lAddr:    "localhost:9500",
			expected: "msg=\"Invalid port range. Ensure start and end ports are valid integers and start <= end.\"\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ln, err := net.Listen("tcp", tt.lAddr)
			assert.NoError(t, err)
			defer ln.Close()

			done := make(chan interface{})
			go func() {
				conn, _ := ln.Accept()
				<-done
				if conn != nil {
					conn.Close()
				}
			}()

			logger, logBuf := createTestSlog()
			app := &application{
				config: config{verbose: true, zero: "localhost"},
				logger: logger,
			}

			go func() {
				app.scanConnection(app.config.zero, tt.port)
				close(done)
			}()

			<-done
			assert.Equal(t, tt.expected, logBuf.String())
		})
	}
}
