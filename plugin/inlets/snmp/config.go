package snmp

import "time"

type ClientConfig struct {
	Timeout                 time.Duration
	Retries                 int
	Version                 int
	UseUnconnectedUDPSocket bool

	// Version 1 and 2
	Community string

	// Version 2 and 3
	MaxRepetitions uint32

	// Version 3 only
	ContextName  string
	SecLevel     string
	SecName      string
	AuthProtocol string
	AuthPassword string
	PrivProtocol string
	PrivPassword string
	EngineID     string
	EngineBoots  uint32
	EngineTime   uint32

	// path to mib files
	Path       []string
	Translator string
}
