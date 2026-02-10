//go:generate go run main.go

package main

import (
	"fmt"
	"log"
	"os"
	"path"
	"path/filepath"
	"regexp"
	"runtime"
	"strings"
)

func main() {
	_, filename, _, ok := runtime.Caller(0)
	if !ok {
		panic("runtime err")
	}

	rootDir := path.Join(path.Dir(filename), "..")

	dstDir := filepath.Join(rootDir, "internal/libwebp")
	srcFolders := []string{"src", "sharpyuv"}
	srcBaseDir := filepath.Join(rootDir, "libwebp_src")

	// The Go and the Webp source must live side-by-side in the same
	// directory.
	//
	// The custom bindings are named with a "a__" prefix. Keep those.
	fis, err := os.ReadDir(dstDir)
	if err != nil {
		log.Fatal(err)
	}

	keepRe := regexp.MustCompile(`^(a__|\.)`)

	for _, fi := range fis {
		if keepRe.MatchString(fi.Name()) {
			continue
		}
		os.Remove(filepath.Join(dstDir, fi.Name()))
	}

	csourceRe := regexp.MustCompile(`\.[ch]$`)

	for _, srcFolder := range srcFolders {
		srcDir := filepath.Join(srcBaseDir, srcFolder)

		err = filepath.Walk(srcDir, func(path string, fi os.FileInfo, err error) error {
			if err != nil {
				return err
			}
			if fi.IsDir() || !csourceRe.MatchString(fi.Name()) {
				return nil
			}

			filename := filepath.ToSlash(strings.TrimPrefix(path, srcBaseDir))
			filename = strings.TrimPrefix(filename, "/")
			target := filepath.Join(dstDir, fi.Name())

			if err := os.WriteFile(target, fmt.Appendf(nil, `#ifndef LIBWEBP_NO_SRC
#include "../../libwebp_src/%s"
#endif
`, filename), 0o644); err != nil {
				return err
			}

			return nil
		})
		if err != nil {
			log.Fatal(err)
		}
	}
}
