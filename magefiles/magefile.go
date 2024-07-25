package main

import (
	"archive/zip"
	"bytes"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"time"

	"github.com/Masterminds/semver/v3"
	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing/object"
	_ "github.com/magefile/mage/mage"
	"github.com/magefile/mage/mg"
	"github.com/magefile/mage/sh"
)

var Default = Build

var vLastVersion string
var vLastCommit string
var vIsNightly bool
var vBuildVersion string

func Build() error {
	mg.Deps(CheckTmp, GetVersion)
	mod := "github.com/OutOfBedlam/tine"
	target := "tine"
	timestamp := time.Now().Format("2006-01-02T15:04:05")
	gitSHA := vLastCommit[0:8]
	goVersion := strings.TrimPrefix(runtime.Version(), "go")

	fmt.Println("Build", "version", vBuildVersion, "sha", gitSHA, "go", goVersion, "time", timestamp)
	env := map[string]string{
		"GO111MODULE": "on",
		"CGO_ENABLED": "1",
	}
	args := []string{"build"}
	ldflags := strings.Join([]string{
		"-X", fmt.Sprintf("%s/engine.tineVersion=%s", mod, vBuildVersion),
		"-X", fmt.Sprintf("%s/engine.tineSha=%s", mod, gitSHA),
		"-X", fmt.Sprintf("%s/engine.goVersion=%s", mod, goVersion),
		"-X", fmt.Sprintf("%s/engine.timeVersion=%s", mod, timestamp),
	}, " ")
	args = append(args, "-ldflags", ldflags)

	// executable file
	ext := ""
	if runtime.GOOS == "windows" {
		ext = ".exe"
	}
	args = append(args, "-o", fmt.Sprintf("./tmp/%s%s", target, ext))
	// source directory
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

func CheckTmp() error {
	_, err := os.Stat("tmp")
	if err != nil && err != os.ErrNotExist {
		err = os.Mkdir("tmp", 0755)
	} else if err != nil && err == os.ErrExist {
		return nil
	}
	return err
}

func GetVersion() error {
	repo, err := git.PlainOpen(".")
	if err != nil {
		return err
	}
	headRef, err := repo.Head()
	if err != nil {
		return err
	}

	headCommit, err := repo.CommitObject(headRef.Hash())
	if err != nil {
		return err
	}

	var lastTag *object.Tag
	iter, err := repo.TagObjects()
	if err != nil {
		return err
	}
	iter.ForEach(func(tagObj *object.Tag) error {
		if !strings.HasPrefix(tagObj.Name, "v") {
			return nil
		}
		if lastTag == nil {
			lastTag = tagObj
		} else {
			lastCommit, _ := lastTag.Commit()
			tagCommit, _ := tagObj.Commit()
			if tagCommit.Author.When.Sub(lastCommit.Author.When) > 0 {
				lastTag = tagObj
			}
		}
		return nil
	})

	lastTagCommit, err := lastTag.Commit()
	if err != nil {
		return err
	}
	vLastVersion = lastTag.Name
	vLastCommit = headCommit.Hash.String()
	vIsNightly = lastTagCommit.Hash.String() != vLastCommit
	lastTagSemVer, err := semver.NewVersion(vLastVersion)
	if err != nil {
		return err
	}

	if lastTagSemVer.Prerelease() == "" {
		if vIsNightly {
			vBuildVersion = fmt.Sprintf("v%d.%d.%d-snapshot", lastTagSemVer.Major(), lastTagSemVer.Minor(), lastTagSemVer.Patch()+1)
		} else {
			vBuildVersion = fmt.Sprintf("v%d.%d.%d", lastTagSemVer.Major(), lastTagSemVer.Minor(), lastTagSemVer.Patch())
		}
	} else {
		suffix := lastTagSemVer.Prerelease()
		if vIsNightly && strings.HasPrefix(suffix, "rc") {
			n, _ := strconv.Atoi(suffix[2:])
			suffix = fmt.Sprintf("rc%d-snapshot", n+1)
		}
		vBuildVersion = fmt.Sprintf("v%d.%d.%d-%s", lastTagSemVer.Major(), lastTagSemVer.Minor(), lastTagSemVer.Patch(), suffix)
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

func Test() error {
	mg.Deps(CheckTmp)
	env := map[string]string{
		"GO111MODULE": "on",
		"CGO_ENABLED": "1",
	}
	testArgs := []string{
		"test",
		"-cover",
		"-coverprofile", "./tmp/coverage.out",
		"./...",
	}

	if err := sh.RunWithV(env, "go", testArgs...); err != nil {
		return err
	}
	if output, err := sh.Output("go", "tool", "cover", "-func", "./tmp/coverage.out"); err != nil {
		return err
	} else {
		lines := strings.Split(output, "\n")
		fmt.Println(lines[len(lines)-1])
	}
	return nil
}

func Package() error {
	return PackageX(runtime.GOOS, runtime.GOARCH)
}

func PackageX(targetOS string, targetArch string) error {
	mg.Deps(CleanPackage, GetVersion, CheckTmp)
	bdir := fmt.Sprintf("tine-%s-%s-%s", vBuildVersion, targetOS, targetArch)
	_, err := os.Stat("dist")
	if err != os.ErrNotExist {
		os.RemoveAll(filepath.Join("dist", bdir))
	}
	os.MkdirAll(filepath.Join("dist", bdir), 0755)

	if targetOS == "windows" {
		if err := os.Rename(filepath.Join("tmp", "tine.exe"), filepath.Join("dist", bdir, "tine.exe")); err != nil {
			return err
		}
	} else {
		if err := os.Rename(filepath.Join("tmp", "tine"), filepath.Join("./dist", bdir, "tine")); err != nil {
			return err
		}
	}

	err = archivePackage(fmt.Sprintf("./dist/%s.zip", bdir), filepath.Join("./dist", bdir))
	if err != nil {
		return err
	}

	os.RemoveAll(filepath.Join("dist", bdir))
	return nil
}

func archivePackage(dst string, src ...string) error {
	archive, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer archive.Close()
	zipWriter := zip.NewWriter(archive)

	for _, file := range src {
		archiveAddEntry(zipWriter, file, fmt.Sprintf("dist%s", string(os.PathSeparator)))
	}
	return zipWriter.Close()
}

func archiveAddEntry(zipWriter *zip.Writer, entry string, prefix string) error {
	stat, err := os.Stat(entry)
	if err != nil {
		return err
	}
	if stat.IsDir() {
		entries, err := os.ReadDir(entry)
		if err != nil {
			return err
		}
		entryName := strings.TrimPrefix(entry, prefix)
		entryName = strings.ReplaceAll(strings.TrimPrefix(entryName, string(filepath.Separator)), "\\", "/")
		entryName = entryName + "/"
		_, err = zipWriter.Create(entryName)
		if err != nil {
			return err
		}
		fmt.Println("Archive D", entryName)
		for _, ent := range entries {
			archiveAddEntry(zipWriter, filepath.Join(entry, ent.Name()), prefix)
		}
	} else {
		fd, err := os.Open(entry)
		if err != nil {
			return err
		}
		defer fd.Close()

		entryName := strings.TrimPrefix(entry, prefix)
		entryName = strings.ReplaceAll(strings.TrimPrefix(entryName, string(filepath.Separator)), "\\", "/")
		fmt.Println("Archive F", entryName)
		finfo, _ := fd.Stat()
		hdr := &zip.FileHeader{
			Name:               entryName,
			UncompressedSize64: uint64(finfo.Size()),
			Method:             zip.Deflate,
			Modified:           finfo.ModTime(),
		}
		hdr.SetMode(finfo.Mode())

		w, err := zipWriter.CreateHeader(hdr)
		if err != nil {
			return err
		}
		if _, err := io.Copy(w, fd); err != nil {
			return err
		}
	}
	return nil
}

func CleanPackage() error {
	entries, err := os.ReadDir("./dist")
	if err != nil {
		if err != os.ErrNotExist {
			return nil
		}
	}

	for _, ent := range entries {
		if err = os.RemoveAll(filepath.Join("./dist", ent.Name())); err != nil {
			return err
		}
	}
	return nil
}
