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

package cmd

import (
	"fmt"
	"strings"

	"github.com/desertbit/grml/internal/manifest"
)

type Commands []*Command

type Command struct {
	name string
	path string
	mc   *manifest.Command
	cmds Commands
	deps Commands
}

// Name returns the command's name.
func (c *Command) Name() string {
	return c.name
}

// Path returns the command's full name including parents.
func (c *Command) Path() string {
	return c.path
}

func (c *Command) Alias() []string {
	return c.mc.Alias
}

func (c *Command) Help() string {
	return c.mc.Help
}

func (c *Command) ExecString() string {
	return c.mc.Exec
}

func (c *Command) SubCommands() Commands {
	return c.cmds
}

func (c *Command) HasSubCommands() bool {
	return len(c.cmds) > 0
}

func (c *Command) Deps() Commands {
	return c.deps
}

func ParseManifest(m *manifest.Manifest) (cmds Commands, err error) {
	cmds = make(Commands, 0, m.Commands.Count())

	// Add the commands from the manifest.
	addCommands("", &cmds, m.Commands)

	// Link the dependencies now.
	err = linkDeps(cmds, cmds)
	return
}

func addCommands(parentPath string, cmds *Commands, mcs manifest.Commands) {
	for name, mc := range mcs {
		c := &Command{
			name: name,
			mc:   mc,
			cmds: make(Commands, 0, mc.Commands.Count()),
		}
		*cmds = append(*cmds, c)

		if len(parentPath) == 0 {
			c.path = name
		} else {
			c.path = parentPath + "." + name
		}

		// Add sub commands.
		addCommands(c.path, &c.cmds, mc.Commands)
	}
}

func linkDeps(root, cmds Commands) (err error) {
	var (
		dep *Command
	)
	for _, c := range cmds {
		// Link dependencies for the command.
		for _, d := range c.mc.Deps {
			if len(d) == 0 {
				return fmt.Errorf("command '%s': empty dependency value", c.path)
			}

			dep, err = getCommandByPath(root, d)
			if err != nil {
				return fmt.Errorf("command '%s': invalid dependency value: %w", c.path, err)
			}
			c.deps = append(c.deps, dep)
		}

		// Link all dependencies for all sub commands.
		err = linkDeps(root, c.cmds)
		if err != nil {
			return
		}
	}
	return
}

func getCommandByPath(root Commands, path string) (*Command, error) {
	split := strings.Split(path, ".")
	if len(split) == 0 {
		return nil, fmt.Errorf("invalid command path value")
	}

	var (
		cur    *Command
		parent = root
	)
	for _, name := range split {
		for _, c := range parent {
			if c.name == name {
				cur = c
				break
			}
		}
		if cur == nil {
			return nil, fmt.Errorf("command not found by path: %v", path)
		}
		parent = cur.cmds
	}
	if cur == nil {
		return nil, fmt.Errorf("command not found by path: %v", path)
	}
	return cur, nil
}
