# AppleDouble

Assists in the removal of AppleDouble files from non-macOS filesystems.  Intended to be used in conjunction with `find`, `appledouble` filters for only actual AppleDouble files (by checking their contents), thereby reducing the likelihood of a false positive over using `find` alone.

Specifically, `appledouble` filters for files that:
1. have names prefixed with "`._`".
2. have, as the first 4 bytes of content, the magic sentinel value `0x00051607`.

## Examples

Remove all AppleDouble files on the filesystem (on macOS, omit the `-r` xargs option):
```
find / -type f -name '._*' -print0 | appledouble -0 -print0 | xargs -0 -r rm
```

One might combine this with the following for a comprehensive means of purging all the crappy files macOS deems necessary to liberally spray all over your nice clean non-macOS filesystem (run regularly from cron!):
```
find / -type d -name '.TemporaryItems' -print0 | xargs -r -0 rm
find / -type f -name '.DS_Store' -print0 | xargs -r -0 rm
find / -type f -name '._*' -print0 | appledouble -0 -print0 | xargs -r -0 rm
```

## Usage

```
$ appledouble --help
usage: appledouble [options] [--] [file ...]
   -h, --help     Display this help text
   -v, --version  Display version information
   -q, --quiet    Do not print errors to stderr
   -0             Accept NUL delimitered input from stdin (compatible with find's -print0)
   -n             Accept newline delimitered input from stdin
   -print0        Output NUL delimitered results to stdout (compatible with xarg's -0)
   -printn        Output newline delimited results to stdout
   --             Interpret all following arguments as files, not options
```

You can separately control how filenames are delimited in both the input and output, to suit the tools or situations you're using `appledouble` in. Failure to use the correct delimiting will result in `appledouble` failing to understand the input, or the other tools failing to understand the output of `appledouble`.

Input:
* **-n** : delimit with `"\n"` (default). Specify when typing in the console, or when using find's `-print` option.
* **-0** : delimit with `"\0"` (nul). Specify when using find's `-print0` option.

Output:
* **-printn** : delimit with `"\n"` (default). For readable console output, or when using xargs *without* its `-0` option.
* **-print0** : delimit with `"\0"` (nul). Specify when using with xarg's `-0` option.

> **IMPORTANT**: you really should be using `"\0"` delimiting throughout when in scripts, to ensure odd filenames don't result in surprise failures.  This means: using find's `-print0`, appledouble's `-0` and `-print0`, and xarg's `-0` options, as per the example above.

## Installation

For the time being, you need the Go compiler (v1.11+) to compile from source.  With that installed:
```
$ git clone https://github.com/mlilley/appledouble.git
$ cd appledouble
$ make
$ sudo make install
```
This places the `appledouble` binary into `/usr/local/bin` (edit the Makefile if you want to put it somewhere else). Ensure whatever install location you use is in the path.

## Warning

Not all AppleDouble files are completely useless and only good for deletion. Although I personally have not encountered situations where I need any of the data in these files, your situation may be different. You should carefully look into whether removal of some or all of these files will impact you.  

Please log an issue if you know of cases where it's important to keep these files.  We might be able to add a (configurable?) exclusion.

## License

MIT