/*
 *  Grumble - A simple build automation tool written in Go
 *  Copyright (C) 2016  Roland Singer <roland.singer[at]desertbit.com>
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

package context

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/desertbit/grml/spec"
	"github.com/desertbit/topsort"
)

//###############//
//### Context ###//
//###############//

// Context defines a grumble build context.
type Context struct {
	// BasePath to the directory containing the source files.
	BasePath string

	// Targets to build.
	Targets []string

	// The exec target process environment.
	Env []string

	// OnlyPrintAllTargets should print all available targets defined in the spec and exit.
	OnlyPrintAllTargets bool

	// Enable verbose execution mode.
	Verbose bool

	// PrintTarget is called during each target run.
	PrintTarget func(t string)

	// doneTargets defines a slice containing already run targets.
	doneTargets []string
}

// RunTargets runs the specified targets.
// Targets are sorted by their dependencies.
func (c *Context) RunTargets(s *spec.Spec) error {
	graph := topsort.NewGraph()

	// Add all graph nodes.
	for name := range s.Targets {
		graph.AddNode(name)
	}

	// Set the edges (dependencies).
	for name, t := range s.Targets {
		deps, err := t.DepTargets()
		if err != nil {
			return err
		}

		for _, d := range deps {
			graph.AddEdge(name, d.Name())
		}
	}

	// Sort the targets and run them.
	for _, t := range c.Targets {
		if !graph.ContainsNode(t) {
			return fmt.Errorf("target does not exists: %s", t)
		}

		// Do the topological sort for each specified build target.
		sorted, err := graph.TopSort(t)
		if err != nil {
			return err
		}

		for _, st := range sorted {
			tt := s.Targets[st]
			if tt == nil {
				return fmt.Errorf("target does not exists: %s", st)
			}

			err = c.runTarget(tt)
			if err != nil {
				return err
			}
		}
	}

	return nil
}

//###############//
//### Private ###//
//###############//

func (c *Context) runTarget(t *spec.Target) error {
	required, err := c.targetRunRequired(t)
	if err != nil {
		return err
	} else if !required {
		return nil
	}

	// Log if the print function is defined.
	if c.PrintTarget != nil {
		c.PrintTarget(t.Name())
	}

	// Go.
	err = c.run(t.Run)
	if err != nil {
		return err
	}

	// Remember the successfully run target.
	c.doneTargets = append(c.doneTargets, t.Name())

	return nil
}

// targetRunRequired checks if the target should be run
func (c *Context) targetRunRequired(t *spec.Target) (bool, error) {
	// Skip if already run.
	name := t.Name()
	for _, dt := range c.doneTargets {
		if dt == name {
			return false, nil
		}
	}

	// Force a run if a dependency had been run.
	deps, err := t.DepTargets()
	if err != nil {
		return false, err
	}
	for _, d := range deps {
		dName := d.Name()
		for _, dt := range c.doneTargets {
			if dt == dName {
				return true, nil
			}
		}
	}

	// Always run if no outputs are defined.
	if len(t.Outputs) == 0 {
		return true, nil
	}

	// Only run if an output file does not exists.
	for _, o := range t.Outputs {
		_, err := os.Stat(filepath.Join(c.BasePath, o))
		if err != nil {
			if !os.IsNotExist(err) {
				return false, err
			}
			return true, nil
		}
	}

	return false, nil
}

// run a command string in the build context.
func (c *Context) run(cmdStr string) error {
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
	cmd.Dir = c.BasePath
	cmd.Env = c.Env
	return cmd.Run()
}
