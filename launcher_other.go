//go:build !windows

package main

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"
)

func runLauncher(baseDir, url string, logs <-chan string, stop func()) {
	fmt.Printf("MDOnline running at %s\n", url)
	fmt.Println("Press Ctrl+C to stop.")

	done := make(chan os.Signal, 1)
	signal.Notify(done, os.Interrupt, syscall.SIGTERM)

	go func() {
		for msg := range logs {
			fmt.Println(msg)
		}
	}()

	<-done
	stop()
}
