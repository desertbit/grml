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

// attachOptions registers the 'options', 'options check', 'options set'
// builtins under addCmd, operating on the option scope at scopePath.
// Pass scopePath "" for the root scope (registered at the top level);
// pass a command path to attach the UI under that command's grumble subtree.
func (a *app) attachOptions(addCmd func(cmd *grumble.Command), scopePath string) {
	cmd := &grumble.Command{
		Name: "options",
		Help: "print & handle options",
		Run: func(c *grumble.Context) error {
			a.printOptions(scopePath)
			return nil
		},
	}

	cmd.AddCommand(&grumble.Command{
		Name: "check",
		Help: "select options",
		Run: func(c *grumble.Context) error {
			return a.optionsCheck(scopePath)
		},
	})

	cmd.AddCommand(&grumble.Command{
		Name: "set",
		Help: "set a specific choice option",
		Args: func(args *grumble.Args) {
			args.String("option", "name of option")
		},
		Completer: func(prefix string, args []string) []string {
			opts := a.options[scopePath]
			if opts == nil {
				return nil
			}
			var words []string
			for name := range opts.Choices {
				if strings.HasPrefix(name, prefix) {
					words = append(words, name)
				}
			}
			sort.Strings(words)
			return words
		},
		Run: func(c *grumble.Context) error {
			return a.optionsSet(scopePath, c.Args.String("option"))
		},
	})

	addCmd(cmd)
}

func (a *app) optionsCheck(scopePath string) error {
	opts := a.options[scopePath]
	if opts == nil || len(opts.Bools) == 0 {
		return fmt.Errorf("no check options available")
	}

	names := make([]string, 0, len(opts.Bools))
	var defaults []string
	for name, o := range opts.Bools {
		names = append(names, name)
		if o.Value {
			defaults = append(defaults, name)
		}
	}
	sort.Strings(names)

	var selected []string
	prompt := &survey.MultiSelect{
		Message: "Select Options:",
		Options: names,
		Default: defaults,
	}
	survey.AskOne(prompt, &selected, nil)

Loop:
	for _, name := range names {
		for _, s := range selected {
			if s == name {
				opts.Bools[name].Value = true
				continue Loop
			}
		}
		opts.Bools[name].Value = false
	}
	return nil
}

func (a *app) optionsSet(scopePath, name string) error {
	opts := a.options[scopePath]
	if opts == nil {
		return fmt.Errorf("no options in scope")
	}
	o := opts.Choices[name]
	if o == nil {
		return fmt.Errorf("invalid choice option: does not exist")
	}

	prompt := &survey.Select{
		Message: "Select Option:",
		Options: o.Options,
	}
	survey.AskOne(prompt, &o.Active, nil)
	o.UserSet = true
	return nil
}

func (a *app) printOptions(scopePath string) {
	opts := a.options[scopePath]
	if opts == nil {
		return
	}

	fmt.Println()

	config := columnize.DefaultConfig()
	config.Delim = "|"
	config.Glue = "  "
	config.Prefix = "  "

	// Bool options sorted by name.
	var (
		output []string
		keys   []string
	)
	for k := range opts.Bools {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	for _, k := range keys {
		output = append(output, fmt.Sprintf("%s: | %v", k, opts.Bools[k].Value))
	}
	if len(output) > 0 {
		a.printColorln("Check Options:\n")
		fmt.Printf("%s\n\n", columnize.Format(output, config))
	}

	// Choice options sorted by name.
	output = nil
	keys = nil
	for k := range opts.Choices {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	for _, k := range keys {
		output = append(output, fmt.Sprintf("%s: | %v", k, opts.Choices[k].Active))
	}
	if len(output) > 0 {
		a.printColorln("Choice Options:\n")
		fmt.Printf("%s\n\n", columnize.Format(output, config))
	}
}
