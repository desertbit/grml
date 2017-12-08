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
	"path/filepath"
	"runtime"
	"strconv"
	"strings"

	"github.com/abiosoft/ishell"
	"github.com/desertbit/grml/spec"
)

const (
	specFilename = "grml.yaml"
)

var global struct {
	Verbose  bool
	RootPath string
	SpecPath string
	Env      map[string]string
	Spec     *spec.Spec
}

func init() {
	global.Env = make(map[string]string)
}

func main() {
	// Parse the command line arguments.
	args, err := parseArgs()
	if err != nil {
		fatalErr(err)
	}

	global.Verbose = args.Verbose
	setNoColor(args.NoColor)

	if args.PrintHelp {
		printFlagsHelp()
		os.Exit(0)
	}

	// Set the initial root path to the current working dir if not set through flags.
	if len(args.RootPath) > 0 {
		global.RootPath = args.RootPath
	} else {
		global.RootPath, err = os.Getwd()
		if err != nil {
			fatalErr(fmt.Errorf("failed to obtain the current working directory: %v", err))
		}
	}

	// Get the absolute path.
	global.RootPath, err = filepath.Abs(global.RootPath)
	if err != nil {
		fatalErr(err)
	}

	// Prepare the environment variables.
	// Inherit the current process environment.
	for _, v := range os.Environ() {
		p := strings.Index(v, "=")
		if p > 0 {
			global.Env[v[0:p]] = v[p+1:]
		}
	}
	global.Env["ROOT"] = global.RootPath
	global.Env["NUMCPU"] = strconv.Itoa(runtime.NumCPU())

	// Read the specification file.
	global.SpecPath = filepath.Join(global.RootPath, specFilename)
	global.Spec, err = spec.ParseSpec(global.SpecPath, global.Env)
	if err != nil {
		fatalErr(fmt.Errorf("spec file: %v", err))
	}

	err = addSpecCommands(global.Spec)
	if err != nil {
		fatalErr(err)
	}

	err = runShell(args)
	if err != nil {
		fatalErr(err)
	}
}

func addSpecCommands(spec *spec.Spec) (err error) {
	for name, c := range spec.Commands {
		exc := NewExecContext(c)

		sc := &ishell.Cmd{
			Name:    name,
			Aliases: c.Aliases,
			Help:    c.Help,
			Func: func(c *ishell.Context) {
				exErr := exc.Exec()
				if exErr != nil {
					printError(exErr)
				}
			},
		}

		if len(c.Commands) > 0 {
			addSubCommands(sc, c.Commands)
		}

		addCmd(sc)
	}

	return
}

func addSubCommands(parent *ishell.Cmd, commands spec.Commands) {
	for name, c := range commands {
		exc := NewExecContext(c)

		sc := &ishell.Cmd{
			Name:    name,
			Aliases: c.Aliases,
			Help:    c.Help,
			Func: func(c *ishell.Context) {
				exErr := exc.Exec()
				if exErr != nil {
					printError(exErr)
				}
			},
		}

		if len(c.Commands) > 0 {
			addSubCommands(sc, c.Commands)
		}

		parent.AddCmd(sc)
	}
}
