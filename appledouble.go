package main

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
)

const VERSION = "0.0.2"

type options struct {
	help    bool
	debug   bool
	version bool
	files   []string
}

type testInfo struct {
	filenameOk bool
	contentOk  bool
	contentErr error
}

func parseArgs() (options, error) {
	opts := options{
		help:    false,
		debug:   false,
		version: false,
		files:   []string{},
	}

	endOpts := false

	for x := 1; x < len(os.Args); x++ {
		arg := os.Args[x]

		if !endOpts {
			if arg == "--" {
				endOpts = true
			} else if strings.HasPrefix(arg, "--") {
				if arg == "--help" {
					opts.help = true
				} else if arg == "--version" {
					opts.version = true
				} else if arg == "--debug" {
					opts.debug = true
				} else {
					return opts, fmt.Errorf("unrecognized option: %s", arg)
				}
			} else if strings.HasPrefix(arg, "-") {
				if arg == "-h" {
					opts.help = true
				} else if arg == "-v" {
					opts.version = true
				} else if arg == "-d" {
					opts.debug = true
				} else {
					return opts, fmt.Errorf("unrecognized option: %s", arg)
				}
			} else {
				opts.files = append(opts.files, os.Args[x])
			}
		} else {
			opts.files = append(opts.files, os.Args[x])
		}
	}

	return opts, nil
}

func help() {
	fmt.Println("usage: appledouble [options] [file ...]")
	fmt.Println("   -h, --help     Display this help text")
	fmt.Println("   -v, --version  Display version information")
	fmt.Println("   -d, --debug    Display debug information when reading from stdin")
}

func version() {
	fmt.Println("appledouble ", VERSION)
}

func testFileName(path string) bool {
	return strings.HasPrefix(filepath.Base(path), "._")
}

func testFileContent(path string) (bool, error) {
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

func testFile(path string, opts options) testInfo {
	info := testInfo{}

	ok := testFileName(path)
	if !ok {
		return info
	}
	info.filenameOk = true

	ok, err := testFileContent(path)
	if err != nil {
		info.contentErr = err
		return info
	}
	info.contentOk = true

	return info
}

// Outputs results suitable for console (result per line, filename:result format, positives and errors only)
func outputForConsole(path string, info testInfo) {
	if info.filenameOk && info.contentOk {
		fmt.Println(path, ": AppleDouble")
	} else if info.contentErr != nil {
		fmt.Println(path, ": error - ", info.contentErr)
	}
}

// Outputs reuslts suitable for passing to xargs (nul delimited, positives only, errors to stderr)
func outputDelimited(path string, info testInfo) {
	if info.filenameOk && info.contentOk {
		fmt.Print(path, "\000")
	} else if info.contentErr != nil {
		fmt.Fprintln(os.Stderr, path, ": error - ", info.contentErr)
	}
}

// Consumes input files from options (command line args)
func consumeOptions(opts options) error {
	for _, path := range opts.files {
		outputForConsole(path, testFile(path, opts))
	}
	return nil
}

// Consumes input files from stdin (delimited by NULs)
func consumeStdin(opts options) error {
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

	s := bufio.NewScanner(bufio.NewReader(os.Stdin))
	s.Split(nulSplitter)

	for s.Scan() {
		path := s.Text()
		outputDelimited(path, testFile(path, opts))
	}

	err := s.Err()
	return err
}

func main() {
	opts, err := parseArgs()
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	if opts.help {
		help()
		os.Exit(0)
	}

	if opts.version {
		version()
		os.Exit(0)
	}

	if len(opts.files) > 0 {
		err := consumeOptions(opts)
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
		}
	} else {
		err := consumeStdin(opts)
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
		}
	}
}
