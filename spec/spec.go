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
	"io/ioutil"
	"strings"

	yaml "gopkg.in/yaml.v2"
)

//############//
//### Spec ###//
//############//

// Spec defines a grml build file.
type Spec struct {
	//Options map[string]interface{} TODO
	Env     map[string]string  `yaml:"-"`
	EnvMap  yaml.MapSlice      `yaml:"env"`
	Targets map[string]*Target `yaml:"targets"`
}

// ExecEnv returns the execute process environment variables.
func (s Spec) ExecEnv() (env []string) {
	env = make([]string, len(s.Env))
	i := 0

	for k, v := range s.Env {
		env[i] = fmt.Sprintf("%s=%s", k, v)
		i++
	}
	return
}

// DefaultTarget returns the default run target if specified.
// Otherwise nil is returned.
func (s Spec) DefaultTarget() *Target {
	for _, t := range s.Targets {
		if t.Default {
			return t
		}
	}
	return nil
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

//##############//
//### Public ###//
//##############//

// ParseSpec parses a grml build file.
// Pass a preset environment map, which will be added to the final spec's environment.
func ParseSpec(path string, env map[string]string) (s *Spec, err error) {
	// Parse the spec.
	data, err := ioutil.ReadFile(path)
	if err != nil {
		return
	}

	s = new(Spec)
	err = yaml.Unmarshal(data, s)
	if err != nil {
		return
	}

	// Prepare and evaluate the environment variables.
	s.Env = make(map[string]string)
	for _, i := range s.EnvMap {
		key := fmt.Sprintf("%v", i.Key)
		value := fmt.Sprintf("%v", i.Value)

		for k, v := range s.Env {
			value = strings.Replace(value, fmt.Sprintf("${%s}", k), v, -1)
		}
		for k, v := range env {
			value = strings.Replace(value, fmt.Sprintf("${%s}", k), v, -1)
		}

		s.Env[key] = value
	}

	// Merge the environments.
	for k, v := range env {
		if _, ok := s.Env[k]; !ok {
			s.Env[k] = v
		}
	}

	// Initialize the private target values.
	for name, t := range s.Targets {
		err = t.init(name, s)
		if err != nil {
			err = fmt.Errorf("target '%s': %v", name, err)
			return
		}
	}

	return
}
