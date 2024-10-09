package main

import (
	"fmt"
	"net"
	"os"
	"strconv"
	"strings"
)

func (app *application) scanConnection(host, portRange string) {
	var ports []string

	if strings.Contains(portRange, "-") {
		parts := strings.Split(portRange, "-")
		if len(parts) != 2 {
			app.logger.Error("Invalid port range format")
			os.Exit(1)
		}
		startPort, err1 := strconv.Atoi(parts[0])
		endPort, err2 := strconv.Atoi(parts[1])
		if err1 != nil || err2 != nil || startPort > endPort {
			app.logger.Error("Invalid port range. Ensure start and end ports are valid integers and start <= end.")
			return
		}

		for p := startPort; p <= endPort; p++ {
			ports = append(ports, strconv.Itoa(p))
		}
	} else {
		ports = append(ports, portRange)
	}

	for _, port := range ports {
		addr := net.JoinHostPort(host, port)
		conn, err := net.Dial("tcp", addr)
		if err == nil {
			msg := fmt.Sprintf("Connection to %s %s [%s]\n", host, conn.RemoteAddr(), conn.RemoteAddr().Network())
			app.logger.Info(msg)
			if app.config.verbose {
				fmt.Printf(msg)
			}
			conn.Close()
			return
		}
	}
	msg := fmt.Sprintf("Could not connect to %s:%s\n", host, portRange)
	app.logger.Info(msg)
	if app.config.verbose {
		fmt.Printf(msg)
	}
	return
}
