package main

import (
	"bufio"
	"context"
	"encoding/json"
	"log"
	"logserver/protocol"
	"net"
	"sync"
)

func handleLogProducer(ctx context.Context, wg *sync.WaitGroup, conn *net.TCPConn, buffer chan<- *protocol.LogEntry) {
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
	if err := scanner.Err(); err != nil {
		log.Printf("Unable to read from stream: %v\n", err)
		return
	}

	go func() {
		<-ctx.Done()
		if err := conn.Close(); err != nil {
			log.Printf("Failed to close client connection: %v\n", err)
		}
	}()
	var bytesRead bool
	for {
		bytesRead = false
		for scanner.Scan() {
			entry := new(protocol.LogEntry)
			data := scanner.Bytes()
			entry.Unmarshal(data)
			buffer <- entry
			bytesRead = true
		}
		if err := scanner.Err(); err != nil || !bytesRead {
			_ = conn.Close()
			return
		}
	}
}
