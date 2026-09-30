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
	"strconv"
	"strings"

	"github.com/desertbit/grml/internal/cmd"
)

// appFlags lists the long names of the grml app flags, including grumble's
// builtin ones. Options must not use these names (see checkOptionNames).
// appValueFlags lists those which take their value as a separate argument
// (e.g. '-d path'). Both must be kept in sync with the flags in Run.
var (
	appFlags = map[string]bool{
		"help": true, "nocolor": true,
		"directory": true, "file": true, "verbose": true,
	}
	appValueFlags = map[string]bool{
		"-d": true, "--directory": true,
		"-f": true, "--file": true,
	}
)

// checkOptionNames returns an error if an option in any scope is named like
// a grml app flag, because it could not be set from the command line.
func (a *app) checkOptionNames() error {
	scopes := make([]string, 0, len(a.options))
	for sp := range a.options {
		scopes = append(scopes, sp)
	}
	sort.Strings(scopes)

	for _, sp := range scopes {
		opts := a.options[sp]
		names := make([]string, 0, len(opts.Bools)+len(opts.Choices))
		for name := range opts.Bools {
			names = append(names, name)
		}
		for name := range opts.Choices {
			names = append(names, name)
		}
		sort.Strings(names)

		for _, name := range names {
			if !appFlags[name] {
				continue
			}
			if sp == "" {
				return fmt.Errorf("option '%s' conflicts with the grml flag --%s", name, name)
			}
			return fmt.Errorf("option '%s' of '%s' conflicts with the grml flag --%s",
				name, strings.ReplaceAll(sp, ".", " "), name)
		}
	}
	return nil
}

// optionFlag is an option override given on the command line as
// '--name=value' (or '--name' for bool options).
type optionFlag struct {
	name     string
	value    string
	hasValue bool
	pos      int // index in the remaining args at which the flag appeared
}

// extractOptionFlags removes option flags from args and returns the remaining
// args together with the extracted flags. An option flag is a '--' token whose
// name starts with a letter and is not one of the grml app flags, which are
// left for grumble to parse (checkOptionNames rules out options named alike). Everything after '--' is left untouched.
func extractOptionFlags(args []string) (rest []string, flags []optionFlag) {
	for i, a := range args {
		if a == "--" {
			return append(rest, args[i:]...), flags
		}
		if !strings.HasPrefix(a, "--") || len(a) < 3 || !isLetter(a[2]) {
			rest = append(rest, a)
			continue
		}

		f := optionFlag{name: a[2:], pos: len(rest)}
		if p := strings.Index(f.name, "="); p >= 0 {
			f.name, f.value, f.hasValue = f.name[:p], f.name[p+1:], true
		}
		if appFlags[f.name] {
			rest = append(rest, a)
			continue
		}
		flags = append(flags, f)
	}
	return
}

func isLetter(b byte) bool {
	return (b >= 'a' && b <= 'z') || (b >= 'A' && b <= 'Z')
}

// applyOptionFlags sets the option values given on the command line.
// args are the command line args with the option flags already removed.
// Each flag is resolved against the command path formed by the command words
// preceding it: flags before any command apply to the root scope, flags after
// e.g. 'release' may also name options of the release scope. Inner scopes
// shadow outer ones, matching the option env layering at exec time.
func (a *app) applyOptionFlags(args []string, flags []optionFlag) error {
	// Skip the grml app flags to find the first command word.
	start := 0
	for start < len(args) && strings.HasPrefix(args[start], "-") {
		if args[start] == "--" {
			start++
			break
		}
		if appValueFlags[args[start]] {
			start++
		}
		start++
	}
	if start > len(args) {
		start = len(args)
	}

	for _, f := range flags {
		var path string
		if f.pos > start {
			path = commandPath(a.commands, args[start:f.pos])
		}
		if err := a.setOption(path, f); err != nil {
			return err
		}
	}
	return nil
}

// commandPath resolves words to the path of the deepest matching command.
// Resolution stops at the first word which is not a (sub) command, because
// the remaining words are the command's positional args.
func commandPath(cs cmd.Commands, words []string) (path string) {
Loop:
	for _, w := range words {
		for _, c := range cs {
			if c.Name() == w || containsString(c.Alias(), w) {
				path = c.Path()
				cs = c.SubCommands()
				continue Loop
			}
		}
		return
	}
	return
}

func containsString(list []string, s string) bool {
	for _, v := range list {
		if v == s {
			return true
		}
	}
	return false
}

// setOption applies f to the innermost option scope visible from cmdPath
// that declares an option with f's name.
func (a *app) setOption(cmdPath string, f optionFlag) error {
	scopes := a.activeOptionScopes(cmdPath)
	for i := len(scopes) - 1; i >= 0; i-- {
		opts := a.options[scopes[i]]

		if o, ok := opts.Bools[f.name]; ok {
			if !f.hasValue {
				o.Value = true
				return nil
			}
			v, err := strconv.ParseBool(f.value)
			if err != nil {
				return fmt.Errorf("option --%s: invalid bool value '%s'", f.name, f.value)
			}
			o.Value = v
			return nil
		}

		if o, ok := opts.Choices[f.name]; ok {
			if !f.hasValue {
				return fmt.Errorf("option --%s: missing value, use --%s=<value> (one of: %s)",
					f.name, f.name, strings.Join(o.Options, ", "))
			}
			if !containsString(o.Options, f.value) {
				return fmt.Errorf("option --%s: invalid value '%s' (one of: %s)",
					f.name, f.value, strings.Join(o.Options, ", "))
			}
			o.Active = f.value
			o.UserSet = true // Keep the value across reloads.
			return nil
		}
	}

	if cmdPath == "" {
		return fmt.Errorf("unknown option: --%s", f.name)
	}
	return fmt.Errorf("unknown option for '%s': --%s", strings.ReplaceAll(cmdPath, ".", " "), f.name)
}
