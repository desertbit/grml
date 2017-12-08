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

	"github.com/abiosoft/ishell"
	"github.com/desertbit/columnize"
)

func init() {
	// Columnize options.
	config := columnize.DefaultConfig()
	config.Delim = "|"
	config.Glue = "  "
	config.Prefix = "  "

	// Options Command.
	cmd := &ishell.Cmd{
		Name: "options",
		Help: "print & handle options",
		Func: func(c *ishell.Context) {
			fmt.Println()

			// Print all check options.
			var output []string
			for name, o := range global.Spec.CheckOptions {
				output = append(output, fmt.Sprintf("%s: | %v", name, o))
			}

			if len(output) > 0 {
				printColorln("Check Options:\n")
				shell.Printf("%s\n\n", columnize.Format(output, config))
			}

			// Print all choice options.
			output = nil
			for name, o := range global.Spec.ChoiceOptions {
				output = append(output, fmt.Sprintf("%s: | %v", name, o.Options[o.Set]))
			}

			if len(output) > 0 {
				printColorln("Choice Options:\n")
				shell.Printf("%s\n\n", columnize.Format(output, config))
			}
		},
	}

	// Check command.
	cmd.AddCmd(&ishell.Cmd{
		Name: "check",
		Help: "select options",
		Func: func(c *ishell.Context) {
			l := len(global.Spec.CheckOptions)
			if l == 0 {
				shell.Println("no check options available")
				return
			}

			options := make([]string, l)
			var selected []int

			i := 0
			for name, o := range global.Spec.CheckOptions {
				options[i] = name
				if o {
					selected = append(selected, i)
				}
				i++
			}

			selected = c.Checklist(options, "Select Options:", selected)

		Loop:
			for i = 0; i < l; i++ {
				name := options[i]
				for _, s := range selected {
					if i == s {
						global.Spec.CheckOptions[name] = true
						continue Loop
					}
				}
				global.Spec.CheckOptions[name] = false
			}

			shell.Println()
		},
	})

	// Set command.
	cmd.AddCmd(&ishell.Cmd{
		Name: "set",
		Help: "set a specific choice option",
		Completer: func([]string) []string {
			var words []string
			for name := range global.Spec.ChoiceOptions {
				words = append(words, name)
			}
			return words
		},
		Func: func(c *ishell.Context) {
			if len(c.Args) != 1 {
				shell.Println("invalid args: one choice option required")
				return
			}

			name := c.Args[0]
			o := global.Spec.ChoiceOptions[name]
			if o == nil {
				shell.Println("invalid choice option: does not exists")
				return
			}

			o.Set = c.MultiChoice(o.Options, "Select Option:")
			shell.Println()
		},
	})

	addBuiltinCmd(cmd)
}
