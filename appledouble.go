package main

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
)

const VERSION = "0.0.3"

type options struct {
	help    bool
	version bool
	input0  bool
	inputn  bool
	files   []string
	output0 bool
	outputn bool
	quiet   bool
}

type testInfo struct {
	filenameOk bool
	contentOk  bool
	contentErr error
}

func parseArgs() (options, error) {
	opts := options{
		help:    false,
		version: false,
		input0:  false,
		inputn:  false,
		files:   []string{},
		output0: false,
		outputn: false,
		quiet:   false,
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
				} else if arg == "--quiet" {
					opts.quiet = true
				} else {
					return opts, fmt.Errorf("unrecognized option: %s", arg)
				}
			} else if strings.HasPrefix(arg, "-") {
				if arg == "-h" {
					opts.help = true
				} else if arg == "-v" {
					opts.version = true
				} else if arg == "-q" {
					opts.quiet = true
				} else if arg == "-0" {
					opts.input0 = true
				} else if arg == "-n" {
					opts.inputn = true
				} else if arg == "-print0" {
					opts.output0 = true
				} else if arg == "-printn" {
					opts.outputn = true
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

	if opts.inputn && opts.input0 {
		return opts, fmt.Errorf("options -0 and -n cannot be used together")
	}
	if opts.outputn && opts.output0 {
		return opts, fmt.Errorf("options -print0 and -printn cannot be used together")
	}
	if !opts.input0 && !opts.inputn {
		opts.inputn = true
	}
	if !opts.output0 && !opts.outputn {
		opts.outputn = true
	}

	return opts, nil
}

func help() {
	fmt.Println("usage: appledouble [options] [--] [file ...]")
	fmt.Println("   -h, --help     Display this help text")
	fmt.Println("   -v, --version  Display version information")
	fmt.Println("   -q, --quiet    Do not print errors to stderr")
	fmt.Println("   -0             Accept NUL delimitered input from stdin (compatible with find's -print0)")
	fmt.Println("   -n             Accept newline delimitered input from stdin")
	fmt.Println("   -print0        Output NUL delimitered results to stdout (compatible with xarg's -0)")
	fmt.Println("   -printn        Output newline delimited results to stdout")
	fmt.Println("   --             Interpret all following arguments as files, not options")
}

func version() {
	fmt.Printf("appledouble %s\n", VERSION)
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
	info.contentOk = ok

	return info
}

func outputResult(path string, info testInfo, opts options) {
	if opts.output0 {
		// output positives to stdout (NUL delimited)
		if info.filenameOk && info.contentOk {
			fmt.Printf("%s\000", path)
		}
	} else {
		// output positives to stdout (newline delimited)
		if info.filenameOk && info.contentOk {
			fmt.Printf("%s\n", path)
		}
	}
	// output errors to stderr (always newline delimited)
	if info.contentErr != nil && !opts.quiet {
		fmt.Fprintf(os.Stderr, "error: %s: %s\n", path, info.contentErr)
	}
}

func consumeFilesFromCommandLine(opts options) error {
	for _, path := range opts.files {
		outputResult(path, testFile(path, opts), opts)
	}
	return nil
}

func consumeFilesFromStdin(opts options) error {
	s := bufio.NewScanner(bufio.NewReader(os.Stdin))

	if opts.input0 {
		// split on NUL, skipping empty strings
		s.Split(func(data []byte, atEOF bool) (advance int, token []byte, err error) {
			for i := 0; i < len(data); i++ {
				if data[i] == '\000' {
					if i == 0 {
						return i + 1, nil, nil // skip empty token
					}
					return i + 1, data[:i], nil // non empty token
				}
			}
			if atEOF {
				return 0, data, bufio.ErrFinalToken // final token
			}
			return 0, nil, nil // no delimiter; retry with more data
		})
	}

	for s.Scan() {
		path := s.Text()
		outputResult(path, testFile(path, opts), opts)
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

	// if any files are specified as arguments, ignore anything piped in
	// (same behavior as tools like grep)
	if len(opts.files) > 0 {
		err := consumeFilesFromCommandLine(opts)
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
		}
	} else {
		err := consumeFilesFromStdin(opts)
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
		}
	}
	os.Exit(0)
}
