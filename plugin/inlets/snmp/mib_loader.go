package snmp

import (
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"sync"

	"github.com/sleepinggenius2/gosmi"
)

var m sync.Mutex
var once sync.Once
var cache = make(map[string]bool)

type MibLoader interface {
	appendPath(path string)
	loadModule(path string) error
}

type GosmiMibLoader struct{}

func (loader *GosmiMibLoader) appendPath(path string) {
	m.Lock()
	defer m.Unlock()
	gosmi.AppendPath(path)
}

func (loader *GosmiMibLoader) loadModule(path string) error {
	m.Lock()
	defer m.Unlock()
	_, err := gosmi.LoadModule(path)
	return err
}

func LoadMibsFromPath(paths []string, loader MibLoader) error {
	folders, err := walkPaths(paths)
	if err != nil {
		return err
	}
	for _, path := range folders {
		loader.appendPath(path)
		modules, err := os.ReadDir(path)
		if err != nil {
			slog.Warn("couldn't load module", "path", path, "error", err)
			continue
		}
		for _, entry := range modules {
			info, err := entry.Info()
			if err != nil {
				slog.Warn("couldn't load module", "path", path, "error", err)
				continue
			}
			if info.Mode()&os.ModeSymlink != 0 {
				symlink := filepath.Join(path, info.Name())
				target, err := filepath.EvalSymlinks(symlink)
				if err != nil {
					slog.Warn("couldn't evaluate", "symbolic link", symlink, "error", err)
					continue
				}
				info, err = os.Lstat(target)
				if err != nil {
					slog.Warn("couldn't stat", "target", target, "error", err)
					continue
				}
			}
			if info.Mode().IsRegular() {
				err := loader.loadModule(info.Name())
				if err != nil {
					slog.Warn("couldn't load module", "path", path, "error", err)
					continue
				}
			}
		}
	}
	return nil
}

func walkPaths(paths []string) ([]string, error) {
	once.Do(gosmi.Init)
	folders := []string{}

	for _, mibPath := range paths {
		// Check if we loaded that path already and skip it if so
		m.Lock()
		cached := cache[mibPath]
		cache[mibPath] = true
		m.Unlock()
		if cached {
			continue
		}

		err := filepath.Walk(mibPath, func(path string, info os.FileInfo, err error) error {
			if info == nil {
				slog.Warn("No mibs found")
				if os.IsNotExist(err) {
					slog.Warn("MIB path doesn't exist", "path", mibPath)
				} else if err != nil {
					return err
				}
				return nil
			}

			if info.Mode()&os.ModeSymlink != 0 {
				target, err := filepath.EvalSymlinks(path)
				if err != nil {
					slog.Warn("Couldn't evaluate", "symbolic links", path, "error", err)
				}
				info, err = os.Lstat(target)
				if err != nil {
					slog.Warn("Couldn't stat", "target", target, "error", err)
				}
				path = target
			}
			if info.IsDir() {
				folders = append(folders, path)
			}

			return nil
		})
		if err != nil {
			return folders, fmt.Errorf("couldn't walk path %q: %w", mibPath, err)
		}
	}
	return folders, nil
}
