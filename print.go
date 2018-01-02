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
	"fmt"

	"github.com/fatih/color"
)

func setNoColor(b bool) {
	color.NoColor = b
}

func printGRML() {
	fmt.Println("               _ ")
	fmt.Println(" ___ ___ _____| |")
	fmt.Println("| . |  _|     | |")
	fmt.Println("|_  |_| |_|_|_|_|")
	fmt.Println("|___|            ")
	fmt.Println("")
}

func printColor(s string) {
	color.Set(color.FgYellow)
	fmt.Print(s)
	color.Unset()
}

func printColorln(s string) {
	printColor(s + "\n")
}
