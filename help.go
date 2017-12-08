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
	"sort"

	"github.com/abiosoft/ishell"
	"github.com/desertbit/columnize"
)

type HelpCommands []*HelpCommand

type HelpCommand struct {
	Name     string
	Help     string
	Commands HelpCommands
}

var HelpMap struct {
	Builtins HelpCommands
	Commands HelpCommands
}

func sortHelpMap() {
	sort.Slice(HelpMap.Builtins, func(i, j int) bool {
		return HelpMap.Builtins[i].Name < HelpMap.Builtins[j].Name
	})

	sort.Slice(HelpMap.Commands, func(i, j int) bool {
		return HelpMap.Commands[i].Name < HelpMap.Commands[j].Name
	})

	// TODO: sort all sub commands as soon as used.
	// Right now only the first level of sub commands is used.
	for _, c := range HelpMap.Commands {
		sort.Slice(c.Commands, func(i, j int) bool {
			return c.Commands[i].Name < c.Commands[j].Name
		})
	}
}

func init() {
	// Columnize options.
	config := columnize.DefaultConfig()
	config.Delim = "|"
	config.Glue = "  "
	config.Prefix = "  "

	// Add the main shell help command.
	addBuiltinCmd(&ishell.Cmd{
		Name: "help",
		Help: "display help",
		Func: func(c *ishell.Context) {
			printGRML()
			shell.Println()

			var output []string
			for _, c := range HelpMap.Builtins {
				output = append(output, fmt.Sprintf("%s: | %v", c.Name, c.Help))
			}

			printColorln("Builtins:")
			printColorln("=========\n")
			shell.Printf("%s\n\n", columnize.Format(output, config))

			output = nil
			for _, c := range HelpMap.Commands {
				output = append(output, fmt.Sprintf("%s: | %v", c.Name, c.Help))
			}

			printColorln("Commands:")
			printColorln("=========\n")
			if len(output) > 0 {
				shell.Printf("%s\n\n", columnize.Format(output, config))
			}

			printColorln("Sub Commands:")
			printColorln("=============\n")

			// Only print the first level of sub commands.
			for _, c := range HelpMap.Commands {
				if len(c.Commands) == 0 {
					continue
				}

				output = nil
				for _, sc := range c.Commands {
					output = append(output, fmt.Sprintf("%s: | %v", sc.Name, sc.Help))
				}

				printColor(c.Name + ":")
				shell.Printf("\n%s\n\n", columnize.Format(output, config))
			}
		},
	})
}
