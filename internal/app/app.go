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
	defaultManifestFilename = "grml.yaml"
)

type app struct {
	*grumble.App

	fgColor      *color.Color
	verbose      bool
	rootPath     string
	manifestPath string

	env      map[string]string
	manifest *manifest.Manifest
	options  map[string]*options.Options // keyed by scope path; "" is root scope
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
				f.String("d", "directory", ".", "set the root directory path")
				f.String("f", "file", defaultManifestFilename, "set an alternative grml file (relative to the root directory)")
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

		// Get the absolute path.
		a.rootPath, err = filepath.Abs(a.rootPath)
		if err != nil {
			return err
		}

		// Resolve the manifest file path. Absolute paths are taken as-is;
		// relative paths are resolved against the root directory.
		a.manifestPath = filepath.Join(a.rootPath, flags.String("file"))

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
			a.printOptions("")
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

	// Prepare options (per-scope).
	a.options, err = a.manifest.ParseOptions()
	if err != nil {
		return fmt.Errorf("failed to parse options: %v", err)
	}
	// Attach the root scope's options UI at the top level.
	if _, ok := a.options[""]; ok {
		a.attachOptions(a.AddCommand, "")
	}

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

	// Restore each scope's options independently. Scopes that no longer
	// exist after the reload are silently dropped.
	for sp, o := range a.options {
		if old, ok := oldOpts[sp]; ok {
			o.Restore(old)
		}
	}
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
			Help:    a.evalVar(a.cmdEnv(c), c.Help()), // Help messages may contain scoped variables.
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

		// If this command declared its own options scope, attach the options
		// UI under it (e.g. 'labrat options check', 'labrat options set foo').
		if _, ok := a.options[c.Path()]; ok {
			a.attachOptions(gc.AddCommand, c.Path())
		}

		// Add this grumble command to the parent.
		parentAddCmd(gc)
	}
}

// evalVar interpolates ${VAR} references in str using the provided env map.
// Options are not included; pass them via the env at the call site if needed.
func (a *app) evalVar(env map[string]string, str string) string {
	for key, value := range env {
		key = fmt.Sprintf("${%s}", key)
		str = strings.Replace(str, key, value, -1)
	}
	return str
}

// cmdEnv layers a command's scope chain on top of the root env.
func (a *app) cmdEnv(c *cmd.Command) map[string]string {
	env := a.env
	for _, scope := range c.Envs() {
		env = manifest.EvalEnvSlice(scope, env)
	}
	return env
}

// execEnv returns the execute process environment variables for c,
// with c's scope chain layered on top of the root env.
func (a *app) execEnv(c *cmd.Command) (env []string) {
	// Environment variables (root + scoped).
	for k, v := range a.cmdEnv(c) {
		env = append(env, fmt.Sprintf("%s=%s", k, v))
	}

	// Layer options across applicable scopes. Walk outermost (root) to
	// innermost (c's path); inner scopes shadow outer for same-named options.
	bools := make(map[string]bool)
	choices := make(map[string]string)
	for _, sp := range a.activeOptionScopes(c.Path()) {
		opts := a.options[sp]
		for k, o := range opts.Bools {
			bools[k] = o.Value
		}
		for k, o := range opts.Choices {
			choices[k] = o.Active
		}
	}
	for k, v := range bools {
		env = append(env, fmt.Sprintf("%s=%v", k, v))
	}
	for k, v := range choices {
		env = append(env, fmt.Sprintf("%s=%s", k, v))
	}
	return
}

// activeOptionScopes returns the scope paths that contribute options to a
// command at cmdPath, ordered outermost (root) first to innermost (cmdPath).
// Only scopes that actually have an Options entry are included.
func (a *app) activeOptionScopes(cmdPath string) []string {
	var scopes []string
	if _, ok := a.options[""]; ok {
		scopes = append(scopes, "")
	}
	if cmdPath == "" {
		return scopes
	}
	parts := strings.Split(cmdPath, ".")
	for i := 1; i <= len(parts); i++ {
		sp := strings.Join(parts[:i], ".")
		if _, ok := a.options[sp]; ok {
			scopes = append(scopes, sp)
		}
	}
	return scopes
}
