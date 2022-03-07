package main

import (
	"errors"
	"fmt"
	"io/fs"
	"math/rand"
	"os"
	"path/filepath"
	"strings"
	"time"
)

var (
	cwd       string
	imagesExt = []string{
		".jpg",
		".jpeg",
		".png",
		".gif",
		".svg",
		".bmp",
		".raw",
		".tiff",
		".mov",
		".avi",
		".mp4",
	}
	metDestDirAlready bool = false
)

const (
	destRoot = "output"
)

func main() {
	var err error
	cwd, err = os.Getwd()
	if err != nil {
		fmt.Printf("ERROR getting working directory: %v\n", err)
		os.Exit(1)
	}

	err = filepath.WalkDir(cwd, handleFile)
	if err != nil {
		fmt.Printf("ERROR: %v\n", err)
		os.Exit(1)
	}
}

func handleFile(f string, d fs.DirEntry, fileError error) error {
	// Handle a file met on a filesystem.
	// If It's a directory, walk that dir.
	// If it's a file of correct extension, move it to a computed destination depending on modTime
	if fileError != nil {
		return fileError
	}

	if d.IsDir() {
		// Don't deal with directories. If root dir, skip it entirely, otherwise continue walking
		return skipIfDestDir(f)
	}

	if !isValidExtension(f) {
		// Don't deal with non-images
		return nil
	}

	err := moveFileToNewDestination(f, d)
	if err != nil {
		return err
	}

	return nil
}

func skipIfDestDir(f string) error {
	// Avoid compute on each dir if we found Dest Dir already before
	if metDestDirAlready {
		return nil
	}
	absD, err := filepath.Abs(f)
	if err != nil {
		return err
	}
	// do not handle destRoot if present
	if absD == filepath.Join(cwd, destRoot) {
		metDestDirAlready = true
		return fs.SkipDir
	}
	return nil
}

func isValidExtension(f string) bool {
	for _, ext := range imagesExt {
		if strings.HasSuffix(strings.ToLower(f), ext) {
			return true
		}
	}
	fmt.Printf("Ignoring : %v\n", f)
	return false
}

func moveFileToNewDestination(f string, d fs.DirEntry) error {
	info, err := d.Info()
	if err != nil {
		return err
	}
	fileNameSource := filepath.Base(f)
	modTime := info.ModTime()
	y, m := modTime.Year(), modTime.Month()
	destDir := filepath.Join(destRoot, fmt.Sprint(y), fmt.Sprint(m))
	fmt.Printf("Creating dir %v ...\n", destDir)
	err = os.MkdirAll(destDir, 0777)
	if err != nil {
		return err
	}
	absSource, err := filepath.Abs(f)
	if err != nil {
		return err
	}
	destFile := dealWithAlreadyExistingfile(destDir, fileNameSource)
	absDest, err := filepath.Abs(destFile)
	if err != nil {
		return err
	}
	fmt.Printf("Moving %v to %v ...\n", absSource, absDest)
	return os.Rename(absSource, absDest)
	//return nil
}

func dealWithAlreadyExistingfile(destDir, fileNameSource string) string {
	destFile := filepath.Join(fmt.Sprint(destDir), fileNameSource)
	_, err := os.Stat(destFile)
	if !errors.Is(err, os.ErrNotExist) {
		rand.Seed(time.Now().UnixNano())
		addedSuffix := fmt.Sprint("-", rand.Int())
		destFileExt := filepath.Ext(destFile)
		return filepath.Join(destDir, strings.TrimSuffix(fileNameSource, destFileExt)+addedSuffix+destFileExt)
	}
	return destFile
}
