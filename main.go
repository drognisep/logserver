package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"logserver/protocol"
	"net"
	"os"
	"os/signal"
	"sync"
	"time"
)

const (
	messageBufferSize = 100
)

func handleSignals() context.Context {
	sigs := make(chan os.Signal, 2)
	signal.Notify(sigs, os.Interrupt, os.Kill)
	ctx, cancel := context.WithCancel(context.Background())

	go func() {
		var hasCancelled bool
		for {
			select {
			case <-sigs:
				if !hasCancelled {
					log.Println("Signal received, stopping gracefully")
					cancel()
					hasCancelled = true
					continue
				}
				log.Println("Signal received, stopping NOW")
				os.Exit(1)
			}
		}
	}()
	return ctx
}

func main() {
	ctx := handleSignals()
	messages := make(chan *protocol.LogEntry, messageBufferSize)
	done := messageWriter(messages)

	listener, err := net.ListenTCP("tcp", &net.TCPAddr{Port: 6839})
	if err != nil {
		log.Fatalf("Failed to start network listener: %v\n", err)
	}
	defer listener.Close()

	wg := &sync.WaitGroup{}
	wg.Add(1)
	if err := listen(ctx, wg, listener, messages); err != nil {
		if err != ErrContextStop {
			log.Fatalf("Error listening for log messages: %v\n", err)
		}
	}
	wg.Wait()
	close(messages)
	<-done
}

var ErrContextStop = errors.New("stopping server gracefully")

func listen(ctx context.Context, wg *sync.WaitGroup, listener *net.TCPListener, messages chan<- *protocol.LogEntry) (exitErr error) {
	defer wg.Done()
	defer func() {
		if r := recover(); r != nil {
			exitErr = fmt.Errorf("listen: panic recovered: %v\n", r)
		}
	}()
	go func() {
		<-ctx.Done()
		if err := listener.Close(); err != nil {
			log.Printf("Error closing listener: %v\n", err)
		}
	}()
	log.Println("Listening for log messages")
	for {
		conn, err := listener.AcceptTCP()
		if err != nil {
			select {
			case <-ctx.Done():
				return ErrContextStop
			default:
				log.Printf("Failed to accept connection: %v\n", err)
			}
		}
		if err := conn.SetKeepAlive(true); err != nil {
			log.Printf("Failed to enable keepalive for client %s: %v\n", conn.RemoteAddr().String(), err)
		} else if err := conn.SetKeepAlivePeriod(5 * time.Second); err != nil {
			log.Printf("SetKeepAlivePeriod: %v\n", err)
		}
		wg.Add(1)
		go handleLogProducer(ctx, wg, conn, messages)
	}
}
