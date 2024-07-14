package engine

import (
	"io"
	"slices"
	"sync"
)

type Compressor struct {
	Name            string
	Factory         func(io.Writer) io.WriteCloser
	ContentEncoding string
}

var compressors = make(map[string]*Compressor)
var compressorsLock = sync.RWMutex{}

func RegisterCompressor(reg *Compressor) {
	compressorsLock.Lock()
	defer compressorsLock.Unlock()
	compressors[reg.Name] = reg
}

func UnregisterCompressor(name string) {
	compressorsLock.Lock()
	defer compressorsLock.Unlock()
	delete(compressors, name)
}

func GetCompressor(name string) *Compressor {
	compressorsLock.RLock()
	defer compressorsLock.RUnlock()
	if reg, ok := compressors[name]; ok {
		return reg
	}
	return nil
}

func CompressorNames() []string {
	compressorsLock.RLock()
	defer compressorsLock.RUnlock()
	names := make([]string, 0, len(compressors))
	for name := range compressors {
		names = append(names, name)
	}
	slices.Sort(names)
	return names
}

type Decompressor struct {
	Name    string
	Factory func(io.Reader) io.ReadCloser
}

var decompressors = make(map[string]*Decompressor)
var decompressorsLock = sync.RWMutex{}

func RegisterDecompressor(reg *Decompressor) {
	decompressorsLock.Lock()
	defer decompressorsLock.Unlock()
	decompressors[reg.Name] = reg
}

func UnregisterDecompressor(name string) {
	decompressorsLock.Lock()
	defer decompressorsLock.Unlock()
	delete(decompressors, name)
}

func GetDecompressor(name string) *Decompressor {
	decompressorsLock.RLock()
	defer decompressorsLock.RUnlock()
	if reg, ok := decompressors[name]; ok {
		return reg
	}
	return nil
}

func DecompressorNames() []string {
	decompressorsLock.RLock()
	defer decompressorsLock.RUnlock()
	names := make([]string, 0, len(compressors))
	for name := range decompressors {
		names = append(names, name)
	}
	slices.Sort(names)
	return names
}
