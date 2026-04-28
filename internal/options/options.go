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

package options

import "fmt"

// Options is a single scope's option set. Names are unique within an
// Options instance but may collide across instances — each scope (root
// manifest and each include point that declares 'options:') owns its own
// Options, so two subgrmls can each have a 'debug' option without clashing.
type Options struct {
	Bools   map[string]*Bool
	Choices map[string]*Choice
}

type Bool struct {
	Value bool
}

type Choice struct {
	Active  string
	Options []string
	UserSet bool // true once the user explicitly picked via 'options set'
}

func New() *Options {
	return &Options{
		Bools:   make(map[string]*Bool),
		Choices: make(map[string]*Choice),
	}
}

// Add merges raw option entries (decoded from YAML) into o. Returns an error
// if a name collides with an option already present in this scope.
func (o *Options) Add(raw map[string]interface{}) error {
	for name, i := range raw {
		if _, exists := o.Bools[name]; exists {
			return fmt.Errorf("duplicate option: %v", name)
		}
		if _, exists := o.Choices[name]; exists {
			return fmt.Errorf("duplicate option: %v", name)
		}

		switch v := i.(type) {
		case bool:
			o.Bools[name] = &Bool{Value: v}

		case []interface{}:
			if len(v) == 0 {
				return fmt.Errorf("invalid option: %v", name)
			}
			list := make([]string, len(v))
			for j, iv := range v {
				list[j] = fmt.Sprintf("%v", iv)
			}
			o.Choices[name] = &Choice{
				Active:  list[0],
				Options: list,
			}

		default:
			return fmt.Errorf("invalid option: %v: %v", name, i)
		}
	}
	return nil
}

func (o *Options) Restore(p *Options) {
	// Carry forward bool values when the option still exists in the new
	// configuration.
	for k, v := range o.Bools {
		if pv, ok := p.Bools[k]; ok {
			v.Value = pv.Value
		}
	}

Loop:
	for k, v := range o.Choices {
		pv, ok := p.Choices[k]
		// Skip if the user never explicitly picked: let the new YAML default
		// (the first item) win, otherwise prepending a new option to the
		// list in the file would have no visible effect after reload.
		if !ok || !pv.UserSet {
			continue
		}

		// Ensure the active value exists
		for _, s := range v.Options {
			if pv.Active == s {
				v.Active = s
				v.UserSet = true
				continue Loop
			}
		}
	}
}
