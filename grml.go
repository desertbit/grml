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
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/desertbit/grml/spec"
	"github.com/desertbit/topsort"
)

const (
	specFilename = "grml.yaml"
)

func main() {
	var err error

	// Our build context.
	ctx := &Context{}

	// Set the initial root path to the current working dir.
	ctx.RootPath, err = os.Getwd()
	if err != nil {
		fatalErr(fmt.Errorf("failed to obtain the current working directory: %v", err))
	}

	// Parse the command line arguments.
	err = parseArgs(ctx)
	if err != nil {
		fatalErr(err)
	}

	// Get the absolute path.
	ctx.RootPath, err = filepath.Abs(ctx.RootPath)
	if err != nil {
		fatalErr(err)
	}

	// Prepare the environment variables.
	// Inherit the current process environment.
	env := make(map[string]string)
	for _, v := range os.Environ() {
		p := strings.Index(v, "=")
		if p > 0 {
			env[v[0:p]] = v[p+1:]
		}
	}
	env["ROOT"] = ctx.RootPath

	// Read the specification file.
	specPath := filepath.Join(ctx.RootPath, specFilename)
	spec, err := spec.ParseSpec(specPath, env)
	if err != nil {
		fatalErr(fmt.Errorf("spec file: %v", err))
	}

	// Set the default target if no targets were passed.
	if len(ctx.Targets) == 0 {
		dt := spec.DefaultTarget()
		if dt != nil {
			ctx.Targets = []string{dt.Name()}
		}
	}

	// Print all targets if required.
	if ctx.OnlyPrintAllTargets || len(ctx.Targets) == 0 {
		printTargetsList(spec)
		return
	}

	// Check if the passed targets are valid.
	for _, t := range ctx.Targets {
		tt := spec.Targets[t]
		if tt == nil {
			fatalErr(fmt.Errorf("target does not exists: %s", t))
		}
	}

	// Run the targets.
	err = runTargets(ctx, spec)
	if err != nil {
		fatalErr(err)
	}

	printDone()
}

// runTargets runs the specified targets.
// Targets are sorted by their dependencies.
func runTargets(c *Context, s *spec.Spec) error {
	graph := topsort.NewGraph()

	// Add all graph nodes.
	for name := range s.Targets {
		graph.AddNode(name)
	}

	// Set the edges (dependencies).
	for name, t := range s.Targets {
		for _, d := range t.Deps {
			graph.AddEdge(name, d)
		}
	}

	// Sort the targets and run them.
	for _, tn := range c.Targets {
		t := s.Targets[tn]
		if t == nil || !graph.ContainsNode(tn) {
			return fmt.Errorf("target does not exists: %s", tn)
		}

		// Do the topological sort for each specified build target.
		sorted, err := graph.TopSort(tn)
		if err != nil {
			return err
		}

		for _, st := range sorted {
			tt := s.Targets[st]
			if tt == nil {
				return fmt.Errorf("target does not exists: %s", st)
			}

			err = runTarget(c, tt)
			if err != nil {
				return err
			}
		}
	}

	return nil
}

func runTarget(c *Context, t *spec.Target) error {
	required, err := c.RunRequired(t)
	if err != nil {
		return err
	} else if !required {
		return nil
	}

	// Log.
	printTarget(t.Name())

	// Our process environment.
	env := t.Spec().ExecEnv()

	// Go go go.
	err = execCommand(t.Run, c, env)
	if err != nil {
		return err
	}

	// Remember the successfully run target.
	c.DoneTargets = append(c.DoneTargets, t.Name())

	return nil
}

func execCommand(cmdStr string, c *Context, env []string) error {
	if len(cmdStr) == 0 {
		return nil
	}

	// Prepend the shell attribute to exit immediately on error.
	attr := "set -e\n"

	// Enable verbose mode if set.
	if c.Verbose {
		attr += "set -x\n"
	}

	cmd := exec.Command("sh", "-c", attr+cmdStr)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin
	cmd.Dir = c.RootPath
	cmd.Env = env
	return cmd.Run()
}
