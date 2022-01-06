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
	"os"
	"os/exec"

	"github.com/desertbit/grml/internal/cmd"
)

// execContext defines a grml command execution context.
// Only use once.
type execContext struct {
	done map[*cmd.Command]struct{}
}

func newExecContext() *execContext {
	return &execContext{
		done: make(map[*cmd.Command]struct{}),
	}
}

func (a *app) exec(c *cmd.Command) (err error) {
	ctx := newExecContext()

	// Run the dependecny commands.
	err = a.execCommands(ctx, c.Deps())
	if err != nil {
		return
	}

	// Run the main command.
	err = a.execCommand(ctx, c)
	return
}

func (a *app) execCommands(ctx *execContext, commands []*cmd.Command) (err error) {
	for _, dc := range commands {
		// Run the dependecny commands.
		err = a.execCommands(ctx, dc.Deps())
		if err != nil {
			return
		}

		// Run the command.
		err = a.execCommand(ctx, dc)
		if err != nil {
			return
		}
	}
	return
}

func (a *app) execCommand(ctx *execContext, c *cmd.Command) (err error) {
	// Check if this command did not run already.
	_, ok := ctx.done[c]
	if ok {
		return
	}

	// Log.
	a.printColorln("exec: " + c.Path())

	// Go go go.
	err = a.runShellCommand(c.ExecString(), a.execEnv())
	if err != nil {
		return
	}

	// Remember the successfully run target.
	ctx.done[c] = struct{}{}
	return
}

func (a *app) runShellCommand(cmdStr string, env []string) error {
	if len(cmdStr) == 0 {
		return nil
	}

	// Prepend the shell attribute to exit immediately on error.
	attr := "set -e\n"

	// Enable verbose mode if set.
	if a.verbose {
		attr += "set -x\n"
	}

	cmd := exec.Command("sh", "-c", attr+cmdStr)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin
	cmd.Dir = a.rootPath
	cmd.Env = env
	return cmd.Run()
}
