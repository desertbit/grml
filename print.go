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

	"github.com/fatih/color"
)

func setNoColor(b bool) {
	color.NoColor = b
}

func fatalErr(err error) {
	printError(err)
	os.Exit(1)
}

func printError(err error) {
	color.Set(color.FgRed, color.Bold)
	shell.Print("error:")
	color.Unset()
	shell.Printf(" %v\n", err)
}

func printGRML() {
	color.Set(color.FgYellow, color.Bold)
	defer color.Unset()

	shell.Println("               _ ")
	shell.Println(" ___ ___ _____| |")
	shell.Println("| . |  _|     | |")
	shell.Println("|_  |_| |_|_|_|_|")
	shell.Println("|___|            ")
	shell.Println("")
}

func printFlagsHelp() {
	printGRML()
	shell.Printf("%v {FLAGS} [COMMAND]\n", color.HiYellowString("grml"))

	shell.Println("\nFlags:")
	shell.Println("  -d | --directory   set an alternative root directory path")
	shell.Println("  -v | --verbose     enable verbose execution mode")
	shell.Println("  -h | --help        print this help text")
	shell.Println("  --no-color         disable color output")
	shell.Println("")
}

func printColor(s string) {
	color.Set(color.FgYellow)
	shell.Print(s)
	color.Unset()
}

func printColorln(s string) {
	printColor(s + "\n")
}
