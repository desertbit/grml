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
	"os"
	"path/filepath"

	"github.com/desertbit/grml/spec"
)

//###############//
//### Context ###//
//###############//

// Context defines a grml build context.
type Context struct {
	// RootPath to the directory containing the source files.
	RootPath string

	// Targets to build.
	Targets []string

	// DoneTargets defines a slice containing already run targets.
	DoneTargets []string

	// OnlyPrintAllTargets should print all available targets defined in the spec and exit.
	OnlyPrintAllTargets bool

	// Enable verbose execution mode.
	Verbose bool
}

// RunRequired checks if the target should be run.
func (c *Context) RunRequired(t *spec.Target) (bool, error) {
	// Skip if already run.
	name := t.Name()
	for _, dt := range c.DoneTargets {
		if dt == name {
			return false, nil
		}
	}

	// Force a run if a dependency had been run.
	for _, d := range t.Deps {
		for _, dt := range c.DoneTargets {
			if dt == d {
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
		_, err := os.Stat(filepath.Join(c.RootPath, o))
		if err != nil {
			if !os.IsNotExist(err) {
				return false, err
			}
			return true, nil
		}
	}

	return false, nil
}
