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
	"sort"

	"github.com/desertbit/columnize"
	"github.com/desertbit/grml/spec"

	"github.com/fatih/color"
)

func fatalErr(err error) {
	printError(err)
	os.Exit(1)
}

func printError(err error) {
	color.Set(color.FgRed, color.Bold, color.Underline)
	fmt.Print("error:")
	color.Unset()
	fmt.Printf(" %v\n", err)
}

func printDone() {
	color.Set(color.FgGreen, color.Bold, color.Underline)
	fmt.Println("done")
	color.Unset()
}

func printTarget(t string) {
	color.Set(color.FgYellow, color.Bold, color.Underline)
	fmt.Print("target:")
	color.Unset()
	color.Set(color.FgYellow, color.Bold)
	fmt.Printf(" %v\n", t)
	color.Unset()
}

func printTargetsList(s *spec.Spec) {
	if len(s.Targets) == 0 {
		fmt.Println("no targets available")
		return
	}

	// Obtain all groups.
	var groups []string
GroupLoop:
	for _, t := range s.Targets {
		for _, g := range groups {
			if g == t.HelpGroup {
				continue GroupLoop
			}
		}
		groups = append(groups, t.HelpGroup)
	}

	sort.Strings(groups)

	config := columnize.DefaultConfig()
	config.Delim = "|"
	config.Glue = "    "

	for _, g := range groups {
		if len(g) == 0 {
			fmt.Println()
		} else {
			fmt.Printf("\n%s:\n\n", g)
		}

		var names []string
		for name := range s.Targets {
			names = append(names, name)
		}
		sort.Strings(names)

		var output []string
		for _, name := range names {
			t := s.Targets[name]
			if t.HelpGroup == g {
				output = append(output, fmt.Sprintf("%s | %s", name, t.Help))
			}
		}
		fmt.Printf("%s\n\n", columnize.Format(output, config))
	}
}
