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

package spec

import (
	"fmt"
	"strings"
)

//###############//
//### Command ###//
//###############//

type Commands map[string]*Command

// Command defines a build command.
type Command struct {
	Aliases     []string   `yaml:"aliases"`
	Help        string     `yaml:"help"`
	Deps        []*Command `yaml:"-"`
	DepsStrings []string   `yaml:"deps"`
	Exec        string     `yaml:"exec"`
	Commands    Commands   `yaml:"commands"`

	name     string
	fullName string
	spec     *Spec
}

// Name returns the command's name.
func (c *Command) Name() string {
	return c.name
}

// FullName returns the command's full name including parents.
func (c *Command) FullName() string {
	return c.fullName
}

// Spec returns the command's spec.
func (c *Command) Spec() *Spec {
	return c.spec
}

//#########################//
//### Command - Private ###//
//#########################//

func (c *Command) init(parentFullName, name string, spec *Spec) (err error) {
	c.name = name
	c.spec = spec

	if len(parentFullName) > 0 {
		c.fullName = parentFullName + "." + name
	} else {
		c.fullName = name
	}

	// Evaluate the variables.
	c.Help = c.spec.evaluateVars(c.Help)

	// Initialize the sub commands.
	for name, sc := range c.Commands {
		err = sc.init(c.fullName, name, spec)
		if err != nil {
			return fmt.Errorf("command '%s': %v", name, err)
		}
	}

	return nil
}

func (c *Command) linkDeps() (err error) {
	var parent Commands
	var cur *Command

	for _, d := range c.DepsStrings {
		if len(d) == 0 {
			return fmt.Errorf("empty dependency value")
		}

		if strings.HasPrefix(d, ".") {
			d = strings.TrimPrefix(d, ".")
			parent = c.Commands
		} else {
			parent = c.spec.Commands
		}

		split := strings.Split(d, ".")
		if len(split) == 0 {
			return fmt.Errorf("invalid dependency value")
		}

		cur = nil
		for _, name := range split {
			cur = parent[name]
			if cur == nil {
				return fmt.Errorf("invalid dependency value: %v", d)
			}
			parent = cur.Commands
		}

		if cur == nil {
			return fmt.Errorf("invalid dependency value: %v", d)
		}

		c.Deps = append(c.Deps, cur)
	}

	// Finally link all dependencies for all sub commands.
	for name, sc := range c.Commands {
		err = sc.linkDeps()
		if err != nil {
			return fmt.Errorf("command '%s': %v", name, err)
		}
	}

	return nil
}
