package main

import (
	"flag"
	"github.com/wansir/broken-link-scanner/scanner"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"
)

func main() {
	threads := flag.Int("threads", 5, "")
	maxRetries := flag.Int("max-retries", 3, "")
	timeout := flag.Duration("timeout", time.Second*15, "")
	flag.Parse()
	root := flag.Arg(0)

	log.Printf("root: %s", root)

	stopChan := make(chan struct{})
	c := make(chan os.Signal, 2)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-c
		close(stopChan)
		<-c
		os.Exit(0)
	}()
	scanner.NewScanner(*threads, *maxRetries, *timeout).Scan(root, stopChan)
}
