/*
 *  grml - A simple build automation tool written in Go
 *  Copyright (C) 2017  Roland Singer <roland.singer[at]desertbit.com>
 *
 *  This program is free software: you can redistribute it and/or modify
 *  it under the terms of the GNU General Public License as published by
 *  the Free Software Foundation, either version 3 of the License, or
 *  (at your option) any later version.
 *
 *  This program is distributed in the hope that it will be useful,
 *  but WITHOUT ANY WARRANTY; without even the implied warranty of
 *  MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
 *  GNU General Public License for more details.
 *
 *  You should have received a copy of the GNU General Public License
 *  along with this program.  If not, see <http://www.gnu.org/licenses/>.
 */

package main

import (
	"fmt"
	"os"
	"strings"
)

type Args struct {
	Verbose   bool
	PrintHelp bool
	NoColor   bool
	RootPath  string
	Tail      []string
}

func parseArgs() (a *Args, err error) {
	a = &Args{}
	args := os.Args

	// Remove the program name.
	if len(args) > 0 {
		args = args[1:]
	}

	var i int
	var lastFlagIndex int

	// Helper.
	getNext := func() (string, error) {
		i++
		if i >= len(args) {
			return "", fmt.Errorf("no value specified")
		}
		lastFlagIndex++
		return args[i], nil
	}

	// Parse the Flags.
	for ; i < len(args); i++ {
		f := args[i]

		if !strings.HasPrefix(f, "-") {
			break
		}

		lastFlagIndex++

		switch f {
		case "-h", "--help":
			a.PrintHelp = true
		case "-d", "--directory":
			a.RootPath, err = getNext()
			if err != nil {
				return
			}
		case "-v", "--verbose":
			a.Verbose = true
		case "--no-color":
			a.NoColor = true
		default:
			err = fmt.Errorf("invalid flag: %s", f)
			return
		}
	}

	// Add all unparsed strings to the tail slice.
	a.Tail = args[lastFlagIndex:]

	return
}
