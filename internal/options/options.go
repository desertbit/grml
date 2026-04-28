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

type Options struct {
	Bools   map[string]bool
	Choices map[string]*Choice
}

type Choice struct {
	Active  string
	Options []string
	UserSet bool // true once the user explicitly picked via 'options set'
}

func New() *Options {
	return &Options{
		Bools:   make(map[string]bool),
		Choices: make(map[string]*Choice),
	}
}

func (o *Options) Restore(p *Options) {
	// If the value exists in the previous options, then restore it.
	for k, _ := range o.Bools {
		if v, ok := p.Bools[k]; ok {
			o.Bools[k] = v
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
