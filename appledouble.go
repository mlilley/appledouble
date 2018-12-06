package main

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
)

// Examines file content to determine whether likely to be an AppleDouble file.
// Looks for 0x00051607 in first 4 bytes; common to all AppleDouble files.
func isAppleDoubleFormat(path string) (bool, error) {
	f, err := os.Open(path)
	if err != nil {
		return false, err
	}
	defer f.Close()

	var b [4]byte
	_, err = io.ReadFull(f, b[:])
	if err != nil {
		if err == io.EOF || err == io.ErrUnexpectedEOF {
			return false, nil
		}
		return false, err
	}

	return (b[0] == 0x00 && b[1] == 0x05 && b[2] == 0x16 && b[3] == 0x07), nil
}

// Determines whether filename matches that of an AppleDouble file (._ prefix).
func isAppleDoubleFilename(path string) bool {
	base := filepath.Base(path)
	return strings.HasPrefix(base, "._")
}

func consumeFromFile(f *os.File) {
	// Custom split func that splits on NUL, and skips empty strings
	nulSplitter := func(data []byte, atEOF bool) (advance int, token []byte, err error) {
		for i, j := 0, 0; i < len(data); i++ {
			if data[i] == '\000' {
				if i-j > 0 {
					return i + 1, data[j:i], nil
				}
				j = i + 1
			}
		}
		return 0, data, bufio.ErrFinalToken
	}

	s := bufio.NewScanner(bufio.NewReader(f))
	s.Split(nulSplitter)
	for s.Scan() {
		path := s.Text()
		if !isAppleDoubleFilename(path) {
			continue
		}
		t, err := isAppleDoubleFormat(path)
		if err != nil {
			continue
		}
		if !t {
			continue
		}
		fmt.Fprintf(os.Stdout, "%s\000", path)
	}

	err := s.Err()
	if err != nil {
		fmt.Fprintln(os.Stderr, "error: ", err)
	}
}

func main() {
	consumeFromFile(os.Stdin)
}
