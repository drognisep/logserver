package main

import (
	"encoding/json"
	"log"
	"logserver/protocol"
	"net"
)

var (
	version = &protocol.VersionSpecifier{MajorVersion: 1}
)

// This is just for testing the client side of the interaction
func main() {
	conn, err := net.Dial("tcp", ":6839")
	if err != nil {
		log.Fatalf("Failed to establish connection: %v\n", err)
	}
	defer conn.Close()

	jsonVersion := MarshalNewline(version)
	if _, err := conn.Write(jsonVersion); err != nil {
		log.Panicf("Failed to write version specifier to stream: %v\n", err)
	}

	if _, err := conn.Write([]byte("Hello, logs!\n")); err != nil {
		log.Panicf("Failed to send log message to server: %v\n", err)
	}
}

func Unmarshal(data []byte, v interface{}) {
	if err := json.Unmarshal(data, v); err != nil {
		panic(err)
	}
}

func Marshal(v interface{}) []byte {
	data, err := json.Marshal(v)
	if err != nil {
		panic(err)
	}
	return data
}

func MarshalNewline(v interface{}) []byte {
	data := Marshal(v)
	return []byte(string(data) + "\n")
}
