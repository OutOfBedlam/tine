package main

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"runtime"
	"strings"
	"time"

	_ "github.com/magefile/mage/mage"
	"github.com/magefile/mage/mg"
	"github.com/magefile/mage/sh"
)

var Default = Build

func Build() error {
	mg.Deps(CheckTmp, CheckGoreleaser)
	if _, err := sh.Output("goreleaser", "build", "--single-target", "--snapshot", "--clean"); err != nil {
		return err
	}
	return nil
}

func Package() error {
	mg.Deps(CheckTmp, CheckGoreleaser)
	if _, err := sh.Output("goreleaser", "release", "--auto-snapshot", "--clean", "--skip=publish"); err != nil {
		return err
	}
	return nil
}

func BuildX(goos string, goarch string) error {
	fmt.Println("BuildX", "......")

	env := map[string]string{"GO111MODULE": "on"}
	env["GOOS"] = goos
	env["GOARCH"] = goarch
	// CGO
	env["CGO_ENABLED"] = "1"

	args := []string{"build"}

	// executable file
	bin := "tine"
	binExt := ""
	if goos == "windows" {
		binExt = ".exe"
	}
	args = append(args, "-o", fmt.Sprintf("./tmp/%s%s", bin, binExt))
	args = append(args, "-ldflags", build_ldflags())
	args = append(args, ".")

	if err := sh.RunV("go", "mod", "tidy"); err != nil {
		return err
	}

	err := sh.RunWithV(env, "go", args...)
	if err != nil {
		return err
	}
	fmt.Println("Build done.")
	return nil
}

func build_ldflags() string {
	// version
	tineVersion := "v"
	if ver, err := CheckVersion(); err == nil {
		tineVersion = ver
	}
	tineSha := ""
	if hash, err := CheckGitSha(); err == nil {
		tineSha = hash
	}
	timeVersion := time.Now().Format("2006-01-02T15:04:05")
	goVersion := strings.TrimPrefix(runtime.Version(), "go")
	mod := "github.com/OutOfBedlam/tine"
	flags := []string{}
	flags = append(flags, fmt.Sprintf(`-X %s/engine.tineVersion=%s`, mod, tineVersion))
	flags = append(flags, fmt.Sprintf(`-X %s/engine.tineSha=%s`, mod, tineSha))
	flags = append(flags, fmt.Sprintf(`-X %s/engine.goVersion=%s`, mod, goVersion))
	flags = append(flags, fmt.Sprintf(`-X %s/engine.timeVersion=%s`, mod, timeVersion))
	return strings.Join(flags, " ")
}

func CheckTmp() error {
	_, err := os.Stat("tmp")
	if err != nil && err != os.ErrNotExist {
		err = os.Mkdir("tmp", 0755)
	} else if err != nil && err == os.ErrExist {
		return nil
	}
	return err
}

func CheckGoreleaser() error {
	const relRepo = "github.com/goreleaser/goreleaser/v2@latest"
	if _, err := sh.Output("goreleaser", "--version"); err != nil {
		err = sh.RunV("go", "install", relRepo)
		if err != nil {
			return err
		}
	}
	return nil
}

func GoreleaserBuild() error {
	if _, err := sh.Output("goreleaser", "build", "--auto-snapshot", "--clean"); err != nil {
		return err
	}
	return nil
}

func CheckVersion() (string, error) {
	buf := &bytes.Buffer{}
	_, err := sh.Exec(nil, buf, io.Discard, "git", "describe", "--tags")
	if err != nil {
		return "", err
	}
	return strings.TrimSuffix(buf.String(), "\n"), err
}

func CheckGitSha() (string, error) {
	buf := &bytes.Buffer{}
	_, err := sh.Exec(nil, buf, io.Discard, "git", "rev-parse", "--short", "HEAD")
	if err != nil {
		return "", err
	}
	return strings.TrimSuffix(buf.String(), "\n"), err
}
