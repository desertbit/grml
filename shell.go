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
	"os/signal"

	"github.com/abiosoft/ishell"
	"github.com/chzyer/readline"
	"github.com/fatih/color"
)

var (
	shell = ishell.NewWithConfig(&readline.Config{
		Prompt: color.HiYellowString("grml » "),
	})
)

func init() {
	addBuiltinCmd(&ishell.Cmd{
		Name: "exit",
		Help: "exit the shell",
		Func: func(c *ishell.Context) {
			c.Stop()
		},
	})

	addBuiltinCmd(&ishell.Cmd{
		Name: "clear",
		Help: "clear the screen",
		Func: func(c *ishell.Context) {
			err := c.ClearScreen()
			if err != nil {
				c.Err(err)
			}
		},
	})
}

func addBuiltinCmd(cmd *ishell.Cmd) {
	shell.AddCmd(cmd)
	HelpMap.Builtins = append(HelpMap.Builtins, &HelpCommand{
		Name: cmd.Name,
		Help: cmd.Help,
	})
}

func addCmd(cmd *ishell.Cmd) {
	shell.AddCmd(cmd)
	addCmdRecursive(cmd, &HelpMap.Commands)
}

func addCmdRecursive(cmd *ishell.Cmd, parentCmds *HelpCommands) {
	helpCmd := &HelpCommand{
		Name: cmd.Name,
		Help: cmd.Help,
	}
	*parentCmds = append(*parentCmds, helpCmd)

	children := cmd.Children()
	for _, subCmd := range children {
		addCmdRecursive(subCmd, &helpCmd.Commands)
	}
}

func runShell(args *Args) (err error) {
	// Check if a command chould be executed in non-interactive mode.
	if len(args.Tail) > 0 {
		return shell.Process(args.Tail...)
	}

	// TODO: Improve this.
	// Ignore interrupt signals, because ishell will handle the interrupts anyway.
	// and the interrupt signals will be passed through automatically to all
	// client processes. They will exit, but the shell will pop up and stay alive.
	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, os.Interrupt)
	go func() {
		for {
			<-signalChan
		}
	}()

	sortHelpMap()
	printGRML()

	shell.Run()
	shell.Close()
	return
}
