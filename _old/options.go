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
	"strings"

	"github.com/desertbit/columnize"
	"github.com/desertbit/grumble"
	"gopkg.in/AlecAivazis/survey.v1"
)

func initOptions() {
	// Columnize options.
	config := columnize.DefaultConfig()
	config.Delim = "|"
	config.Glue = "  "
	config.Prefix = "  "

	// Options Command.
	cmd := &grumble.Command{
		Name: "options",
		Help: "print & handle options",
		Run: func(c *grumble.Context) error {
			fmt.Println()

			// Print all check options.
			var output []string
			for name, o := range global.Spec.CheckOptions {
				output = append(output, fmt.Sprintf("%s: | %v", name, o))
			}

			if len(output) > 0 {
				printColorln("Check Options:\n")
				fmt.Printf("%s\n\n", columnize.Format(output, config))
			}

			// Print all choice options.
			output = nil
			for name, o := range global.Spec.ChoiceOptions {
				output = append(output, fmt.Sprintf("%s: | %v", name, o.Set))
			}

			if len(output) > 0 {
				printColorln("Choice Options:\n")
				fmt.Printf("%s\n\n", columnize.Format(output, config))
			}
			return nil
		},
	}

	// Check command.
	cmd.AddCommand(&grumble.Command{
		Name: "check",
		Help: "select options",
		Run: func(c *grumble.Context) error {
			l := len(global.Spec.CheckOptions)
			if l == 0 {
				return fmt.Errorf("no check options available")
			}

			options := make([]string, l)
			var defaults []string

			i := 0
			for name, o := range global.Spec.CheckOptions {
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
						global.Spec.CheckOptions[o] = true
						continue Loop
					}
				}
				global.Spec.CheckOptions[o] = false
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
			for name := range global.Spec.ChoiceOptions {
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

			o := global.Spec.ChoiceOptions[c.Args.String("option")]
			if o == nil {
				return fmt.Errorf("invalid choice option: does not exists")
			}

			prompt := &survey.Select{
				Message: "Select Option:",
				Options: o.Options,
			}
			survey.AskOne(prompt, &o.Set, nil)
			return nil
		},
	})

	app.AddCommand(cmd)
}
