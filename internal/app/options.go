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
	"sort"
	"strings"

	"github.com/desertbit/columnize"
	"github.com/desertbit/grumble"
	"gopkg.in/AlecAivazis/survey.v1"
)

// TODO: better option command. maybe a single option command to set all options at once in a tui like menu?

func (a *app) initOptions() {
	// Options Command.
	cmd := &grumble.Command{
		Name: "options",
		Help: "print & handle options",
		Run: func(c *grumble.Context) error {
			a.printOptions()
			return nil
		},
	}

	// Check command.
	cmd.AddCommand(&grumble.Command{
		Name: "check",
		Help: "select options",
		Run: func(c *grumble.Context) error {
			l := len(a.options.Bools)
			if l == 0 {
				return fmt.Errorf("no check options available")
			}

			options := make([]string, l)
			var defaults []string

			i := 0
			for name, o := range a.options.Bools {
				options[i] = name
				if o {
					defaults = append(defaults, name)
				}
				i++
			}

			var selected []string
			prompt := &survey.MultiSelect{
				Message: "Select Options:",
				Options: options,
				Default: defaults,
			}
			survey.AskOne(prompt, &selected, nil)

		Loop:
			for _, o := range options {
				for _, s := range selected {
					if s == o {
						a.options.Bools[o] = true
						continue Loop
					}
				}
				a.options.Bools[o] = false
			}
			return nil
		},
	})

	// Set command.
	cmd.AddCommand(&grumble.Command{
		Name: "set",
		Help: "set a specific choice option",
		Args: func(a *grumble.Args) {
			a.String("option", "name of option")
		},
		Completer: func(prefix string, args []string) []string {
			var words []string
			for name := range a.options.Choices {
				if strings.HasPrefix(name, prefix) {
					words = append(words, name)
				}
			}
			return words
		},
		Run: func(c *grumble.Context) error {
			if len(c.Args) != 1 {
				return fmt.Errorf("invalid args: one choice option required")
			}

			o := a.options.Choices[c.Args.String("option")]
			if o == nil {
				return fmt.Errorf("invalid choice option: does not exists")
			}

			prompt := &survey.Select{
				Message: "Select Option:",
				Options: o.Options,
			}
			survey.AskOne(prompt, &o.Active, nil)
			return nil
		},
	})

	a.AddCommand(cmd)
}

func (a *app) printOptions() {
	fmt.Println()

	var (
		output []string
		keys   []string
	)

	// Columnize options.
	config := columnize.DefaultConfig()
	config.Delim = "|"
	config.Glue = "  "
	config.Prefix = "  "

	// Print all check options sorted.
	for k, _ := range a.options.Bools {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	for _, k := range keys {
		v := a.options.Bools[k]
		output = append(output, fmt.Sprintf("%s: | %v", k, v))

	}
	if len(output) > 0 {
		a.printColorln("Check Options:\n")
		fmt.Printf("%s\n\n", columnize.Format(output, config))
	}

	// Print all choice options sorted.
	output = nil
	keys = nil
	for k, _ := range a.options.Choices {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	for _, k := range keys {
		o := a.options.Choices[k]
		output = append(output, fmt.Sprintf("%s: | %v", k, o.Active))
	}

	if len(output) > 0 {
		a.printColorln("Choice Options:\n")
		fmt.Printf("%s\n\n", columnize.Format(output, config))
	}
}
