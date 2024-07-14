//go:build ignore

package main

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/OutOfBedlam/tine/engine"
)

func main() {
	fmt.Print(engine.DefaultConfigString)
	fmt.Print("\n#\n# inets and outlets\n#\n\n")

	plugins, err := os.ReadDir("plugin")
	if err != nil {
		panic(err)
	}
	for _, lookingfor := range []string{"inlet.toml", "flow.toml", "outlet.toml"} {
		for _, f := range plugins {
			if !f.IsDir() {
				continue
			}
			tomlPath := filepath.Join("plugin", f.Name(), lookingfor)
			if _, err := os.Stat(tomlPath); err != nil {
				continue
			}
			content, err := os.ReadFile(tomlPath)
			if err != nil {
				fmt.Println("# gen-config error:", tomlPath, err.Error())
				panic(err)
			}
			fmt.Printf("%s\n", string(content))
		}
	}
	os.Exit(0)
}
