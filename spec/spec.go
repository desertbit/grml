/*
 *  Grumble - A simple build automation tool written in Go
 *  Copyright (C) 2016  Roland Singer <roland.singer[at]desertbit.com>
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
	"io/ioutil"
	"strings"

	yaml "gopkg.in/yaml.v2"
)

//############//
//### Spec ###//
//############//

// Spec defines a grumble build file.
type Spec struct {
	Env     map[string]string
	Targets map[string]*Target
}

// EnvToSlice maps the environment variables to an os.exec Env slice.
func (s Spec) EnvToSlice() (env []string) {
	env = make([]string, len(s.Env))
	i := 0

	for k, v := range s.Env {
		env[i] = fmt.Sprintf("%s=%s", k, v)
		i++
	}
	return
}

//######################//
//### Spec - Private ###//
//######################//

func (s *Spec) evaluateVars(str string) string {
	for key, value := range s.Env {
		key = fmt.Sprintf("${%s}", key)
		str = strings.Replace(str, key, value, -1)
	}

	return str
}

// targetWithOutput returns the target which creates the given output.
func (s Spec) targetWithOutput(o string) *Target {
	for _, t := range s.Targets {
		for _, to := range t.Outputs {
			if to == o {
				return t
			}
		}
	}
	return nil
}

//##############//
//### Public ###//
//##############//

// ParseSpec parses a grumble build file.
func ParseSpec(path string) (s *Spec, err error) {
	data, err := ioutil.ReadFile(path)
	if err != nil {
		return
	}

	s = new(Spec)
	err = yaml.Unmarshal(data, s)
	if err != nil {
		return
	}

	// Evaluate the environment variables.
	for key, value := range s.Env {
		s.Env[key] = s.evaluateVars(value)
	}

	// Initialize the private target values.
	for name, t := range s.Targets {
		t.init(name, s)
	}

	return
}
