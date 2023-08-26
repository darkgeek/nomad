// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

//go:build !windows
// +build !windows

package host

import (
	"strings"
    "bytes"
    "log"
    "os/exec"
    "strconv"

	"golang.org/x/sys/unix"
)

// uname returns the syscall like `uname -a`
func uname() string {
	u := &unix.Utsname{}
	err := unix.Uname(u)
	if err != nil {
		return err.Error()
	}

	uname := strings.Join([]string{
		nullStr(u.Machine[:]),
		nullStr(u.Nodename[:]),
		nullStr(u.Release[:]),
		nullStr(u.Sysname[:]),
		nullStr(u.Version[:]),
	}, " ")

	return uname
}

func etcHosts() string {
	return slurp("/etc/hosts")
}

func resolvConf() string {
	return slurp("/etc/resolv.conf")
}

func nullStr(bs []byte) string {
	// find the null byte
	var i int
	var b byte
	for i, b = range bs {
		if b == 0 {
			break
		}
	}

	return string(bs[:i])
}

func strToUint64(str string) uint64 {
    ui64, err := strconv.ParseUint(str, 10, 64)
    if err != nil {
        panic(err)
    }

    return ui64
}

type df struct {
    p string
    tol uint64
    avl uint64
}

func (d *df) populateDfResult() {
    cmd := exec.Command("df", "-P", "-b")

    var out bytes.Buffer
    cmd.Stdout = &out

    err := cmd.Run()

    if err != nil {
        log.Fatal(err)
    }

    outStr := out.String()
    lines := strings.Split(outStr, "\n")
    for _, ln := range lines[1:] {
        parts := strings.Split(ln, " ")
        if len(parts) == 6 && strings.Compare(d.p, parts[5]) == 0 {
            d.tol = 512*strToUint64(parts[1])
            d.avl = 512*strToUint64(parts[3])
        }
    }
}

func makeDf(path string) (*df, error) {
    return &df{p: path}, nil
}

func (d *df) total() uint64 {
	d.populateDfResult()
	return d.tol
}

func (d *df) available() uint64 {
	d.populateDfResult()
	return d.avl
}
