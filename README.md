# AppleDouble

Assists in the removal of AppleDouble files from non-mac filesystems.

Intended to be used in conjunction with `find`, `appledouble` filters the results of `find` for only those files who's content indicates they are actually AppleDouble formatted files, thereby reducing the likelihood of a false positive over using `find` alone.

Two specific conditions are used when filtering:
1. the file's name is prefixed with "`._`".
2. the file's first 4 bytes match the AppleDouble magic sentinal 0x00051607.

Currently, `appledouble` supports only a NUL delimited file list as input, as produced by `find`'s `-print0` option (conveniently enough).

## Usage

Linux:
```
find / -type f -name '._*' -print0 | appledouble | xargs -r -0 rm
```

Mac (where xargs does not support `-r`):
```
find / -type f -name '._*' -print0 | appledouble | xargs -0 rm
```

One might combine the above with additional commands to complete the picture:
```
find / -type f -name '.DS_Store' | xargs -r -0 rm
find / -type d -name '.TemporaryItems' | xargs -r -0 rm
find / -type f -name '._' -print0 | appledouble | xargs -r -0 rm
```

## Warning

Not all AppleDouble files are completely useless and only good for deletion. Although I personally have not encountered situations where I need any of the data in these files, your situation may be different. You should carefully look into whether removal of some or all of these files will impact you.  

Please log an issue if you know of cases where it's important to keep these files.  We might be able to add a (configurable?) exclusion.

## Installation

Binaries for various systems may become available if I get around to setting it up.  Until then, you'll need the Go compiler (v1.11+) to compile from source:
```
$ git clone https://github.com/mlilley/appledouble.git
$ cd appledouble
$ make
$ sudo make install
```
Then put the built `appledouble` binary somewhere and ensure that somewhere is in your `$PATH`.

## License

MIT