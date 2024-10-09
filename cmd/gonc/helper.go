package main

import (
	"bytes"
	"log/slog"
	"slices"
	"sync"
)

func createTestSlog() (*slog.Logger, *bytes.Buffer) {
	var mu sync.Mutex
	var buf bytes.Buffer
	attr := []string{"msg", "error", "sig"}
	logger := slog.New(slog.NewTextHandler(&buf, &slog.HandlerOptions{
		ReplaceAttr: func(groups []string, a slog.Attr) slog.Attr {
			mu.Lock()
			defer mu.Unlock()

			if len(groups) == 0 && slices.Contains(attr, a.Key) {
				return a
			}
			return slog.Attr{}
		},
	}))
	return logger, &buf
}
