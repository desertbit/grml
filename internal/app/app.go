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

package app

import (
	"fmt"
	"os"
	"os/signal"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"

	"github.com/desertbit/grml/internal/cmd"
	"github.com/desertbit/grml/internal/manifest"
	"github.com/desertbit/grml/internal/options"
	"github.com/desertbit/grumble"
	"github.com/fatih/color"
)

const (
	manifestFilename = "grml.yaml"
)

type app struct {
	*grumble.App

	fgColor      *color.Color
	verbose      bool
	rootPath     string
	manifestPath string

	env      map[string]string
	manifest *manifest.Manifest
	options  *options.Options
	commands cmd.Commands
}

// Run the application.
func Run() {
	a := &app{
		App: grumble.New(&grumble.Config{
			Name:                  "grml",
			Description:           fmt.Sprintf("A simple build automation tool written in Go (version: %v)", manifest.Version),
			Prompt:                "grml » ",
			PromptColor:           color.New(color.FgYellow, color.Bold),
			HelpHeadlineColor:     color.New(color.FgYellow),
			HelpHeadlineUnderline: true,
			HelpSubCommands:       true,

			Flags: func(f *grumble.Flags) {
				f.String("d", "directory", "", "set an alternative root directory path")
				f.Bool("v", "verbose", false, "enable verbose execution mode")
			},
		}),

		fgColor: color.New(color.FgYellow),
		env:     make(map[string]string),
	}

	a.SetPrintASCIILogo(func(gapp *grumble.App) {
		a.printGRML()
	})

	a.OnShell(func(gapp *grumble.App) error {
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

	a.OnInit(func(gapp *grumble.App, flags grumble.FlagMap) (err error) {
		// Initialize global flag values.
		a.verbose = flags.Bool("verbose")
		a.rootPath = flags.String("directory")
		a.setNoColor(gapp.Config().NoColor)

		// Set the initial root path to the current working dir if not set through flags.
		if len(a.rootPath) == 0 {
			a.rootPath, err = os.Getwd()
			if err != nil {
				return fmt.Errorf("failed to obtain the current working directory: %v", err)
			}
		}

		// Get the absolute path.
		a.rootPath, err = filepath.Abs(a.rootPath)
		if err != nil {
			return err
		}
		a.manifestPath = filepath.Join(a.rootPath, manifestFilename)

		// Load the manifest.
		return a.load()
	})

	grumble.Main(a.App)
}

func (a *app) load() (err error) {
	// Remove previous commands first.
	a.Commands().RemoveAll()

	// Add built-int commands.
	a.AddCommand(&grumble.Command{
		Name: "reload",
		Help: "reload the grml file and keep the current options",
		Run: func(c *grumble.Context) (err error) {
			err = a.reload()
			if err != nil {
				return
			}
			a.Println("parsed grml file and reloaded successfully")
			a.printOptions()
			return
		},
	})
	a.AddCommand(&grumble.Command{
		Name: "verbose",
		Help: "set the verbose execution mode",
		Args: func(a *grumble.Args) {
			a.Bool("verbose", "enable or disable the mode")
		},
		Run: func(c *grumble.Context) (err error) {
			a.verbose = c.Args.Bool("verbose")
			return
		},
	})

	// Read the grml file.
	a.manifest, err = manifest.Parse(a.manifestPath)
	if err != nil {
		return fmt.Errorf("grml file: %v", err)
	}

	// Set the updated prompt.
	if a.manifest.Project != "" {
		a.SetPrompt(fmt.Sprintf("grml %s » ", color.New(color.FgWhite, color.Faint).Sprint(a.manifest.Project)))
	}

	// Prepare options
	a.options, err = a.manifest.ParseOptions()
	if err != nil {
		return fmt.Errorf("failed to parse options")
	}
	a.initOptions()

	// Prepare the environment.
	// Inherit the current process environment.
	for _, v := range os.Environ() {
		p := strings.Index(v, "=")
		if p > 0 {
			a.env[v[0:p]] = v[p+1:]
		}
	}
	a.env["PROJECT"] = a.manifest.Project
	a.env["ROOT"] = a.rootPath
	a.env["NUMCPU"] = strconv.Itoa(runtime.NumCPU())
	a.env = a.manifest.EvalEnv(a.env) // Add values from manifest.

	// Group all commands to the builtin group (help message).
	cmds := a.Commands().All()
	for _, c := range cmds {
		c.HelpGroup = "Builtins:"
	}

	// Prepare the commands.
	a.commands, err = cmd.ParseManifest(a.manifest)
	if err != nil {
		return
	}

	// Register the commands to grumble.
	a.registerCommands(a.AddCommand, a.commands)

	return
}

func (a *app) reload() (err error) {
	// Store current options.
	oldOpts := a.options

	// Reset some required values.
	a.env = make(map[string]string)

	// Load the new grml file.
	err = a.load()
	if err != nil {
		return
	}

	// Restore as many options as possible.
	a.options.Restore(oldOpts)
	return
}

func (a *app) registerCommands(parentAddCmd func(cmd *grumble.Command), cs cmd.Commands) {
	for _, c := range cs {
		var (
			localCmd = c // Catch the variable locally for run.
		)
		gc := &grumble.Command{
			Name:    c.Name(),
			Aliases: c.Alias(),
			Help:    a.evalVar(c.Help()), // Help messages may contain variables.
			Args: func(ga *grumble.Args) {
				for _, arg := range localCmd.Args() {
					ga.String(arg, "_")
				}
			},
			Run: func(c *grumble.Context) error {
				var args map[string]string
				if localCmd.HasArgs() {
					args = make(map[string]string)
					for _, arg := range localCmd.Args() {
						args[arg] = c.Args.String(arg)
					}
				}
				return a.exec(localCmd, args)
			},
		}

		// Add sub commands to this grumble command.
		if c.HasSubCommands() {
			a.registerCommands(gc.AddCommand, c.SubCommands())
		}

		// Add this grumble command to the parent.
		parentAddCmd(gc)
	}
}

// Hint: options are not included.
func (a *app) evalVar(str string) string {
	for key, value := range a.env {
		key = fmt.Sprintf("${%s}", key)
		str = strings.Replace(str, key, value, -1)
	}
	return str
}

// execEnv returns the execute process environment variables.
func (a *app) execEnv() (env []string) {
	// Environment variables.
	for k, v := range a.env {
		env = append(env, fmt.Sprintf("%s=%s", k, v))
	}

	// Check options.
	for k, o := range a.options.Bools {
		env = append(env, fmt.Sprintf("%s=%v", k, o))
	}

	// Choice options.
	for k, o := range a.options.Choices {
		env = append(env, fmt.Sprintf("%s=%v", k, o.Active))
	}
	return
}
