package main

import (
	"bufio"
	"bytes"
	"fmt"
	"log/slog"
	"os"

	"github.com/spf13/pflag"
)

type config struct {
	cmd     string
	debug   bool
	hex     bool
	listen  bool
	port    int
	udp     bool
	verbose bool
	zero    string
}

type application struct {
	config config
	logger *slog.Logger
}

func main() {
	var cfg config

	pflag.BoolVarP(&cfg.debug, "debug", "d", false, "debug mode for logs")
	pflag.BoolVarP(&cfg.hex, "hex", "x", false, "hex dumping mode")
	pflag.BoolVarP(&cfg.listen, "listenMode", "l", false, "listen mode for inbound connections")
	pflag.BoolVarP(&cfg.udp, "udp", "u", false, "UDP mode")
	pflag.BoolVarP(&cfg.verbose, "verbose", "v", false, "verbose mode")
	pflag.IntVarP(&cfg.port, "port", "p", 0, "local port number")
	pflag.StringVarP(&cfg.zero, "zero", "z", "", "zero-I/O mode [used for scanning]")
	pflag.StringVarP(&cfg.cmd, "exec", "e", "", "program to exec after connect")

	pflag.Usage = func() {
		var buf bytes.Buffer

		buf.WriteString("Usage:\n")
		buf.WriteString("  gonc [-options] hostname port[s] [ports] ...\n")
		buf.WriteString("  gonc -l -p port [-options] [hostname] [port]\n")
		buf.WriteString("Options:\n")

		fmt.Fprintf(os.Stderr, buf.String())
		pflag.PrintDefaults()
	}

	pflag.Parse()

	if len(pflag.Args()) > 0 {
		fmt.Printf("Incorrect argument format!\n")
		pflag.Usage()
		os.Exit(2)
	}

	logger := createLogger(cfg.debug)

	app := &application{
		config: cfg,
		logger: logger,
	}

	if cfg.listen {
		addr := fmt.Sprintf(":%d", cfg.port)
		if cfg.udp {
			srv := app.NewUDPServer(addr)
			go app.readInput(srv.quit, srv.sendch)
			err := srv.StartUDP()
			if err != nil {
				logger.Error("failed to listen to UDP connections", "error", err)
				os.Exit(1)
			}
		} else {
			srv := app.NewTCPServer(addr)
			go app.readInput(srv.quit, srv.sendch)
			err := srv.StartTCP()
			if err != nil {
				logger.Error("failed to listen to TCP connections", "error", err)
				os.Exit(1)
			}
		}
	}

	if cfg.zero != "" {
		host := cfg.zero
		portRange := pflag.Arg(0)
		app.scanConnection(host, portRange)
		os.Exit(0)
	}
}

func createLogger(debug bool) *slog.Logger {
	opts := slog.HandlerOptions{Level: slog.LevelError}

	if debug {
		opts.Level = slog.LevelInfo
	}

	handler := slog.NewTextHandler(os.Stdout, &opts)
	return slog.New(handler)
}

func (app *application) readInput(quit chan interface{}, sendch chan string) {
	reader := bufio.NewReader(os.Stdin)
	go func() {
		select {
		case <-quit:
			app.logger.Info("quit readInput")
			os.Exit(0)
		}
	}()

	for {
		line, err := reader.ReadString('\n')
		if err != nil {
			app.logger.Info("failed to read from standard input", "error", err)
			return
		}
		sendch <- line
	}
}
