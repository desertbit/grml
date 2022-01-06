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
	Env           map[string]string      `yaml:"-"`
	EnvMap        yaml.MapSlice          `yaml:"env"`
	Options       map[string]interface{} `yaml:"options"`
	ChoiceOptions ChoiceOptions          `yaml:"-"`
	CheckOptions  CheckOptions           `yaml:"-"`
	Commands      Commands               `yaml:"commands"`
}

// ExecEnv returns the execute process environment variables.
func (s Spec) ExecEnv() (env []string) {
	// Environment variables.
	for k, v := range s.Env {
		env = append(env, fmt.Sprintf("%s=%s", k, v))
	}

	// Check options.
	for k, o := range s.CheckOptions {
		env = append(env, fmt.Sprintf("%s=%v", k, o))
	}

	// Choice options.
	for k, o := range s.ChoiceOptions {
		env = append(env, fmt.Sprintf("%s=%v", k, o.Set))
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

func (s *Spec) prepareOptions() (err error) {
	for name, i := range s.Options {
		switch v := i.(type) {
		case bool:
			if _, ok := s.CheckOptions[name]; ok {
				return fmt.Errorf("option already set: %v", name)
			}

			s.CheckOptions[name] = v

		case []interface{}:
			if _, ok := s.ChoiceOptions[name]; ok {
				return fmt.Errorf("option already set: %v", name)
			} else if len(v) == 0 {
				return fmt.Errorf("invalid option: %v", name)
			}

			list := make([]string, len(v))
			for i, iv := range v {
				list[i] = fmt.Sprintf("%v", iv)
			}

			s.ChoiceOptions[name] = &ChoiceOption{
				Set:     list[0],
				Options: list,
			}

		default:
			return fmt.Errorf("invalid option: %v: %v", name, i)
		}
	}
	return
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

	s = &Spec{
		ChoiceOptions: make(ChoiceOptions),
		CheckOptions:  make(CheckOptions),
	}
	err = yaml.UnmarshalStrict(data, s)
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

	// Initialize the commands.
	for name, c := range s.Commands {
		err = c.init("", name, s)
		if err != nil {
			err = fmt.Errorf("command '%s': %v", name, err)
			return
		}
	}

	// Finally link all dependencies.
	for name, c := range s.Commands {
		err = c.linkDeps()
		if err != nil {
			err = fmt.Errorf("command '%s': %v", name, err)
			return
		}
	}

	// Initialize the options.
	err = s.prepareOptions()
	if err != nil {
		return
	}

	return
}
