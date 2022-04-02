package base

import (
	"bytes"
	"fmt"
	"log"
	"os"
	"path"
	"path/filepath"
	"strings"

	"github.com/fatih/color"
)

func TarsGoHome() string {
	dir, err := os.UserHomeDir()
	if err != nil {
		log.Fatal(err)
	}
	home := path.Join(dir, ".tarsgo")
	if _, err := os.Stat(home); os.IsNotExist(err) {
		if err := os.MkdirAll(home, 0o700); err != nil {
			log.Fatal(err)
		}
	}
	return home
}

func TarsGoHomeWithDir(dir string) string {
	home := path.Join(TarsGoHome(), dir)
	if _, err := os.Stat(home); os.IsNotExist(err) {
		if err := os.MkdirAll(home, 0o700); err != nil {
			log.Fatal(err)
		}
	}
	return home
}

func CopyFile(src, dst string, replaces []string) error {
	var err error
	srcinfo, err := os.Stat(src)
	if err != nil {
		return err
	}
	buf, err := os.ReadFile(src)
	if err != nil {
		return err
	}
	var old string
	for i, next := range replaces {
		if i%2 == 0 {
			old = next
			continue
		}
		if !strings.HasSuffix(dst, ".tars") && old == "_SERVANT_" {
			next = FirstUpper(next)
		}
		buf = bytes.ReplaceAll(buf, []byte(old), []byte(next))
	}
	return os.WriteFile(dst, buf, srcinfo.Mode())
}

func CopyDir(src, dst string, replaces, ignores []string) error {
	var err error
	var fds []os.DirEntry
	var srcinfo os.FileInfo

	if srcinfo, err = os.Stat(src); err != nil {
		return err
	}

	if err = os.MkdirAll(dst, srcinfo.Mode()); err != nil {
		return err
	}

	if fds, err = os.ReadDir(src); err != nil {
		return err
	}
	for _, fd := range fds {
		if hasSets(fd.Name(), ignores) {
			continue
		}
		srcfp := path.Join(src, fd.Name())
		dstfp := path.Join(dst, fd.Name())
		if fd.IsDir() {
			if err = CopyDir(srcfp, dstfp, replaces, ignores); err != nil {
				return err
			}
		} else {
			if err = CopyFile(srcfp, dstfp, replaces); err != nil {
				return err
			}
		}
	}
	return nil
}

func hasSets(name string, sets []string) bool {
	for _, ig := range sets {
		if ig == name {
			return true
		}
	}
	return false
}

func Tree(path string, dir string) {
	_ = filepath.Walk(path, func(path string, info os.FileInfo, err error) error {
		if !info.IsDir() {
			fmt.Printf("%s %s (%v bytes)\n", color.GreenString("CREATED"), strings.Replace(path, dir+"/", "", -1), info.Size())
		}
		return nil
	})
}
