package main

import (
	"fmt"
	"os"
	"path/filepath"
)

func overwriteAndCloseFile(f *os.File, size int64) error {
	defer f.Close()
	if stat, err := f.Stat(); err != nil {
		return err
	} else if stat.Size() != size {
		return fmt.Errorf("file size changed during processing")
	}
	const chunkSize = 1024 * 1024
	buf := make([]byte, chunkSize)
	for i := int64(0); i < size; i += int64(chunkSize) {
		writeSize := int64(chunkSize)
		end := i + int64(chunkSize)
		if end > size {
			writeSize = size - i
		}

		n, err := f.WriteAt(buf[:int(writeSize)], i)
		if err != nil {
			return err
		}
		if n != int(writeSize) {
			return fmt.Errorf("short write: wrote %d bytes, expected %d", n, writeSize)
		}

	}
	if err := f.Sync(); err != nil {
		return fmt.Errorf("error syncing file: %w", err)
	}

	return nil
}

func main() {
	if len(os.Args) < 2 {
		fmt.Println("usage: go-shred <filename>")
		return
	}

	fname := os.Args[1]
	info, err := os.Lstat(fname)
	if err != nil {
		fmt.Println("error getting file info:", err)
		return
	}

	if info.Mode()&os.ModeSymlink != 0 {
		fmt.Println("refusing to shred symlink")
		return
	}

	if !info.Mode().IsRegular() {
		fmt.Println("refusing to shred non-regular file")
		return
	}

	size := info.Size()
	if size < 0 {
		fmt.Println("invalid file size")
		return
	}

	if size == 0 {
		fmt.Println("file is empty, nothing to overwrite")
		return
	}
	f, err := os.OpenFile(fname, os.O_WRONLY, 0)
	if err != nil {
		fmt.Println("error opening file for writing:", err)
		return
	}

	if err := overwriteAndCloseFile(f, size); err != nil {
		fmt.Println("error overwriting file:", err)
		return
	}
	if err := os.Remove(fname); err != nil {
		fmt.Println("error deleting file:", err)
		return
	}
	parentDir := filepath.Dir(fname)
	dir, err := os.Open(parentDir)
	if err == nil {
		_ = dir.Sync()
		_ = dir.Close()
	}
	fmt.Println("file overwritten with null bytes and removed successfully")
}
