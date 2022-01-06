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
	"os/signal"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"

	"github.com/desertbit/grml/spec"
	"github.com/desertbit/grumble"
	"github.com/fatih/color"
)

const (
	specFilename = "grml.yaml"
)

var (
	app = grumble.New(&grumble.Config{
		Name:                  "grml",
		Description:           "A simple build automation tool written in Go",
		Prompt:                "grml » ",
		PromptColor:           color.New(color.FgYellow, color.Bold),
		HelpHeadlineColor:     color.New(color.FgYellow),
		HelpHeadlineUnderline: true,
		HelpSubCommands:       true,

		Flags: func(f *grumble.Flags) {
			f.String("d", "directory", "", "set an alternative root directory path")
			f.Bool("v", "verbose", false, "enable verbose execution mode")
		},
	})
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
	app.SetPrintASCIILogo(func(a *grumble.App) {
		printGRML()
	})

	app.OnShell(func(a *grumble.App) error {
		// Ignore interrupt signals, because grumble will handle the interrupts anyway.
		// and the interrupt signals will be passed through automatically to all
		// client processes. They will exit, but the shell will pop up and stay alive.
		signalChan := make(chan os.Signal, 1)
		signal.Notify(signalChan, os.Interrupt)
		go func() {
			for {
				<-signalChan
			}
		}()
		return nil
	})

	app.OnInit(func(a *grumble.App, flags grumble.FlagMap) (err error) {
		// Initialize global flag values.
		global.Verbose = flags.Bool("verbose")
		global.RootPath = flags.String("directory")
		setNoColor(app.Config().NoColor)

		// Set the initial root path to the current working dir if not set through flags.
		if len(global.RootPath) == 0 {
			global.RootPath, err = os.Getwd()
			if err != nil {
				return fmt.Errorf("failed to obtain the current working directory: %v", err)
			}
		}

		// Get the absolute path.
		global.RootPath, err = filepath.Abs(global.RootPath)
		if err != nil {
			return err
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
			return fmt.Errorf("spec file: %v", err)
		}

		// Init the option commands.
		initOptions()

		// Group all commands to the builtin group.
		cmds := app.Commands().All()
		for _, c := range cmds {
			c.HelpGroup = "Builtins:"
		}

		// Register the commands.
		addSpecCommands(global.Spec)

		return nil
	})

	grumble.Main(app)
}

func addSpecCommands(spec *spec.Spec) {
	for name, c := range spec.Commands {
		exc := NewExecContext(c)

		sc := &grumble.Command{
			Name:    name,
			Aliases: c.Aliases,
			Help:    c.Help,
			Run: func(c *grumble.Context) error {
				return exc.Exec()
			},
		}

		if len(c.Commands) > 0 {
			addSubCommands(sc, c.Commands)
		}

		app.AddCommand(sc)
	}
}

func addSubCommands(parent *grumble.Command, commands spec.Commands) {
	for name, c := range commands {
		exc := NewExecContext(c)

		sc := &grumble.Command{
			Name:    name,
			Aliases: c.Aliases,
			Help:    c.Help,
			Run: func(c *grumble.Context) error {
				return exc.Exec()
			},
		}

		if len(c.Commands) > 0 {
			addSubCommands(sc, c.Commands)
		}

		parent.AddCommand(sc)
	}
}
