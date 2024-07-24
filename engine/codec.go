package engine

import (
	"io"
	"slices"
	"sync"
)

type EncoderReg struct {
	Name        string
	Factory     func(EncoderConfig) Encoder
	ContentType string
}

type EncoderConfig struct {
	Writer       io.Writer
	Subformat    string
	Indent       string
	Prefix       string
	Fields       []string
	FormatOption FormatOption
}

var encoders = make(map[string]*EncoderReg)
var encodersLock = sync.RWMutex{}

func RegisterEncoder(reg *EncoderReg) {
	encodersLock.Lock()
	defer encodersLock.Unlock()
	encoders[reg.Name] = reg
}

func UnregisterEncoder(name string) {
	encodersLock.Lock()
	defer encodersLock.Unlock()
	delete(encoders, name)
}

func GetEncoder(name string) *EncoderReg {
	encodersLock.RLock()
	defer encodersLock.RUnlock()
	if reg, ok := encoders[name]; ok {
		return reg
	}
	return nil
}

func EncoderNames() []string {
	encodersLock.RLock()
	defer encodersLock.RUnlock()
	names := make([]string, 0, len(encoders))
	for name := range encoders {
		names = append(names, name)
	}
	slices.Sort(names)
	return names
}

type Encoder interface {
	Encode([]Record) error
}

type DecoderReg struct {
	Name    string
	Factory func(DecoderConfig) Decoder
}

type DecoderConfig struct {
	Reader        io.Reader
	Timeformatter *Timeformatter
	Fields        []string
	Types         []Type
}

var decoders = make(map[string]*DecoderReg)
var decodersLock = sync.RWMutex{}

func RegisterDecoder(reg *DecoderReg) {
	decodersLock.Lock()
	defer decodersLock.Unlock()
	decoders[reg.Name] = reg
}

func UnregisterDecoder(name string) {
	decodersLock.Lock()
	defer decodersLock.Unlock()
	delete(decoders, name)
}

func GetDecoder(name string) *DecoderReg {
	decodersLock.RLock()
	defer decodersLock.RUnlock()
	if reg, ok := decoders[name]; ok {
		return reg
	}
	return nil
}

func DecoderNames() []string {
	decodersLock.RLock()
	defer decodersLock.RUnlock()
	names := make([]string, 0, len(decoders))
	for name := range decoders {
		names = append(names, name)
	}
	slices.Sort(names)
	return names
}

type Decoder interface {
	Decode() ([]Record, error)
}
