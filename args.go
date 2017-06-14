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

	"github.com/fatih/color"
)

func printHelp() {
	fmt.Println("grml {FLAGS} [TARGET] ...")
	fmt.Println("\nFLAGS:")
	fmt.Println("  -d | --directory   set an alternative root directory path")
	fmt.Println("  -l | --list        print a list of all defined targets")
	fmt.Println("  -v | --verbose     enable verbose execution mode")
	fmt.Println("  -h | --help        print this help text")
	fmt.Println("  --no-color         disable color output")
	fmt.Println("")
}

func parseArgs(ctx *Context) (err error) {
	args := os.Args

	// Remove the program name.
	if len(args) > 0 {
		args = args[1:]
	}

	// Helper.
	getNext := func(i int) (string, error) {
		i++
		if i >= len(args) {
			return "", fmt.Errorf("no value specified")
		}
		return args[i], nil
	}

	// Flags.
	var lastFlagIndex int
	for i, f := range args {
		if !strings.HasPrefix(f, "-") {
			break
		}

		lastFlagIndex++

		switch f {
		case "-h", "--help":
			printHelp()
			os.Exit(0)
		case "-l", "--list":
			ctx.OnlyPrintAllTargets = true
		case "-d", "--directory":
			ctx.RootPath, err = getNext(i)
			if err != nil {
				return
			}
		case "-v", "--verbose":
			ctx.Verbose = true
		case "--no-color":
			color.NoColor = true
		default:
			err = fmt.Errorf("invalid flag: %s", f)
			return
		}
	}

	// Passed targets.
	ctx.Targets = args[lastFlagIndex:]

	return
}
