package main

import (
"context"
"net"
"os"
"fmt"
"github.com/wailsapp/wails/v2/pkg/runtime"
)

const socketPath = "/tmp/custos-app.sock"

func StartIPC(ctx context.Context) error {
	// Clean up old socket if it exists but is dead
	os.Remove(socketPath)

	l, err := net.Listen("unix", socketPath)
	if err != nil {
		return err
	}

	go func() {
		defer l.Close()
		for {
			conn, err := l.Accept()
			if err != nil {
				continue
			}

			// On any connection, show the window!
			runtime.WindowShow(ctx)
			conn.Close()
		}
	}()

	return nil
}

func NotifyExistingInstance() bool {
	conn, err := net.Dial("unix", socketPath)
	if err != nil {
		return false // No existing instance
	}
	defer conn.Close()
	// Connection alone brings up the window because of the listener logic
	fmt.Fprintf(conn, "SHOW")
	return true // Instance exists and was notified
}
