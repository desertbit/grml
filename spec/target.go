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

package spec

import (
	"fmt"
	"path/filepath"
)

//##############//
//### Target ###//
//##############//

// Target defines a build target.
type Target struct {
	Help      string   `yaml:"help"`
	HelpGroup string   `yaml:"help-group"`
	Run       string   `yaml:"run"`
	Default   bool     `yaml:"default"`
	Deps      []string `yaml:"deps"`
	Outputs   []string `yaml:"output"`

	name string
	spec *Spec
}

// Name returns the target's name.
func (t *Target) Name() string {
	return t.name
}

// Spec returns the target's spec.
func (t *Target) Spec() *Spec {
	return t.spec
}

//########################//
//### Target - Private ###//
//########################//

func (t *Target) init(name string, spec *Spec) error {
	t.name = name
	t.spec = spec

	for _, d := range t.Deps {
		if len(d) == 0 {
			return fmt.Errorf("empty dependency value")
		}
	}
	for _, o := range t.Outputs {
		if len(o) == 0 {
			return fmt.Errorf("empty output value")
		}
	}

	// Evaluate the variables.
	for i := 0; i < len(t.Deps); i++ {
		t.Deps[i] = t.spec.evaluateVars(t.Deps[i])
	}
	for i := 0; i < len(t.Outputs); i++ {
		t.Outputs[i] = filepath.Clean(t.spec.evaluateVars(t.Outputs[i]))
	}

	return nil
}
