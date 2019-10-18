# inu - use dogbin from your terminal

> Use dogbin/hastebin from your cli

[![GoDoc](https://godoc.org/github.com/dogbin/inu/dogbin?status.svg)](https://godoc.org/github.com/dogbin/inu/dogbin)
[![Go Report Card](https://goreportcard.com/badge/github.com/dogbin/inu)](https://goreportcard.com/report/github.com/dogbin/inu)
[![Build Status](https://travis-ci.com/dogbin/inu.svg?branch=master)](https://travis-ci.com/dogbin/inu)
[![Code Coverage](https://img.shields.io/codecov/c/github/dogbin/inu)](https://codecov.io/gh/dogbin/inu/)

## Installation:

Ensure you have [Go](https://golang.org) installed, and run:

```bash
deletescape@nortia:~$ go get github.com/dogbin/inu
```

## Usage:

```bash
deletescape@nortia:~$ inu help
NAME:
   inu - Use dogbin/hastebin right from your terminal

USAGE:
   inu [global options] command [command options] [arguments...]

VERSION:
   0.1.0

AUTHOR:
   Till Kottmann <me@deletescape.ch>

COMMANDS:
   put, up, p, u,   Create a new paste
   get, show, s     Obtains the contents of a paste
   help, h          Shows a list of commands or help for one command

GLOBAL OPTIONS:
   --server value, -r value  The dogbin/hastebin server to use (default: "del.dog") [%DOGBIN_SERVER%] [~/.inu/server]
   --slug value, -s value    The slug to use instead of the server generated one [haste doesn't support this]
   --file value, -f value    A file to upload to dogbin
   --json, -j                Outputs the result as JSON
   --copy, -c                Additionally puts the created URL in your clipboard
   --help, -h                show help
   --version, -v             print the version

COPYRIGHT:
   (c) 2019 Till Kottmann
```

You can also simply pipe into the `inu` command like this:

```bash
deletescape@nortia:~$ echo "Awesome test for README" | inu
https://del.dog/nodabaqisa
```

## Library use

To access dogbin in your Go app you can import the `github.com/dogbin/inu/dogbin` package. For now use this cli as a reference until a proper godoc exists.
