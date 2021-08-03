package protocol

import (
	"encoding/json"
	"fmt"
)

type LogEntry struct {
	Level   string `json:"level"`
	Service string `json:"serviceName"`
	Message string `json:"message"`
}

func (l *LogEntry) Unmarshal(bytes []byte) {
	if err := json.Unmarshal(bytes, l); err != nil {
		l.Message = string(bytes)
		l.Level = "INFO"
	}
}

func (l *LogEntry) String() string {
	var service string
	var level string
	if l.Service != "" {
		service = fmt.Sprintf("[%s] ", l.Service)
	}
	if l.Level != "" {
		level = fmt.Sprintf("%s: ", l.Level)
	}
	return service + level + l.Message
}
