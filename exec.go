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
	"os"
	"os/exec"

	"github.com/desertbit/grml/spec"
)

//###################//
//### ExecContext ###//
//###################//

// ExecContext defines a grml execution context.
type ExecContext struct {
	cmd      *spec.Command
	doneCMDs map[*spec.Command]struct{}
}

func NewExecContext(cmd *spec.Command) *ExecContext {
	return &ExecContext{
		cmd: cmd,
	}
}

func (e *ExecContext) Exec() (err error) {
	// Reset.
	e.doneCMDs = make(map[*spec.Command]struct{})

	// Run the dependecny commands.
	err = e.runCommands(e.cmd.Deps)
	if err != nil {
		return
	}

	// Run the main command.
	err = e.runCommand(e.cmd)
	if err != nil {
		return
	}

	return
}

func (e *ExecContext) runCommands(commands []*spec.Command) (err error) {
	for _, dc := range commands {
		// Run the dependecny commands.
		err = e.runCommands(dc.Deps)
		if err != nil {
			return
		}

		// Run the command.
		err = e.runCommand(dc)
		if err != nil {
			return
		}
	}
	return
}

func (e *ExecContext) runCommand(cmd *spec.Command) (err error) {
	// Check if this command did not run already.
	_, ok := e.doneCMDs[cmd]
	if ok {
		return
	}

	// Log.
	printColorln("exec: " + cmd.FullName())

	// Our process environment.
	env := global.Spec.ExecEnv()

	// Go go go.
	err = e.runShellCommand(cmd.Exec, env)
	if err != nil {
		return
	}

	// Remember the successfully run target.
	e.doneCMDs[cmd] = struct{}{}

	return
}

func (e *ExecContext) runShellCommand(cmdStr string, env []string) error {
	if len(cmdStr) == 0 {
		return nil
	}

	// Prepend the shell attribute to exit immediately on error.
	attr := "set -e\n"

	// Enable verbose mode if set.
	if global.Verbose {
		attr += "set -x\n"
	}

	cmd := exec.Command("sh", "-c", attr+cmdStr)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin
	cmd.Dir = global.RootPath
	cmd.Env = env
	return cmd.Run()
}
