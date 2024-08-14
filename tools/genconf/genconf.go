//go:build ignore

package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"unicode"

	"github.com/OutOfBedlam/tine/engine"
)

func main() {
	baseDir := ".."
	fmt.Print(engine.DefaultConfigString)
	fmt.Print("\n#\n# inets and outlets\n#\n\n")

	_, err := os.ReadDir(filepath.Join(baseDir, "plugins"))
	if err != nil {
		panic(err)
	}
	dirFor(
		[]string{
			filepath.Join(baseDir, "plugins"),
			filepath.Join(baseDir, "x"),
		},
		[]string{"*_inlet.toml", "*_flow.toml", "*_outlet.toml"})

	os.Exit(0)
}

func dirFor(dirs []string, lookingForList []string) {
	for _, lookingFor := range lookingForList {
		for _, dir := range dirs {
			plugins, err := os.ReadDir(dir)
			if err != nil {
				panic(err)
			}
			for _, f := range plugins {
				if !f.IsDir() {
					continue
				}
				tomlMatches, err := filepath.Glob(filepath.Join(dir, f.Name(), lookingFor))
				if err != nil || len(tomlMatches) == 0 {
					continue
				}
				for _, tomlPath := range tomlMatches {
					content, err := os.ReadFile(tomlPath)
					if err != nil {
						fmt.Println("# gen-config error:", tomlPath, err.Error())
						panic(err)
					}
					for _, line := range strings.Split(string(content), "\n") {
						if strings.HasPrefix(strings.TrimSpace(line), "#") || strings.TrimSpace(line) == "" {
							fmt.Printf("%s\n", line)
						} else {
							whitespace := []rune{}
							remains := ""
							for i, c := range line {
								if unicode.IsSpace(c) {
									whitespace = append(whitespace, c)
								} else {
									remains = line[i:]
									break
								}
							}
							if remains[0] == '[' {
								fmt.Printf("%s#%s\n", string(whitespace), remains)
							} else {
								fmt.Printf("%s# %s\n", string(whitespace), remains)
							}
						}
					}
				}
			}
		}
	}
}
