package main

import (
	"bufio"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"logserver/protocol"
	"net"
	"os"
	"os/signal"
	"sync"
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
	messages := make(chan string, messageBufferSize)
	done := messageWriter(messages)

	listener, err := net.Listen("tcp", ":6839")
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

func messageWriter(messages <-chan string) <-chan struct{} {
	done := make(chan struct{})
	go func() {
		for {
			select {
			case msg, more := <-messages:
				if !more {
					close(done)
					return
				}
				log.Println(msg)
			}
		}
	}()
	return done
}

var ErrContextStop = errors.New("stopping server gracefully")

func listen(ctx context.Context, wg *sync.WaitGroup, listener net.Listener, messages chan string) (exitErr error) {
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
		conn, err := listener.Accept()
		if err != nil {
			select {
			case <-ctx.Done():
				return ErrContextStop
			default:
				log.Printf("Failed to accept connection: %v\n", err)
			}
		}
		addr := conn.RemoteAddr()
		log.Printf("Received connection from %s\n", addr.String())
		wg.Add(1)
		go handleLogProducer(ctx, wg, conn, messages)
	}
}

func handleLogProducer(ctx context.Context, wg *sync.WaitGroup, conn net.Conn, buffer chan <-string) {
	defer wg.Done()
	// Client should send a version specifier first.
	scanner := bufio.NewScanner(conn)
	if scanner.Scan() {
		data := scanner.Bytes()
		version := new(protocol.VersionSpecifier)
		if err := json.Unmarshal(data, version); err != nil {
			log.Printf("Unable to read protocol version: %v\n", err)
			return
		}
		if err := protocol.ValidateVersion(*version); err != nil {
			log.Printf("Invalid version specifier: %v\n", err)
			return
		}
	}
	if err := scanner.Err(); err != nil  {
		log.Printf("Unable to read from stream: %v\n", err)
		return
	}
	log.Println("Received valid version value from client")

	go func() {
		<-ctx.Done()
		if err := conn.Close(); err != nil {
			log.Printf("Failed to close client connection: %v\n", err)
		}
	}()
	for scanner.Scan() {
		buffer <- scanner.Text()
	}
	if err := scanner.Err(); err != nil {
		log.Printf("Error reading from stream: %v\n", err)
		return
	}
	log.Printf("Closed connection from %s\n", conn.RemoteAddr().String())
}
