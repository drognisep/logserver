package main

import (
	"log"
	"logserver/protocol"
)

func messageWriter(messages <-chan *protocol.LogEntry) <-chan struct{} {
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
