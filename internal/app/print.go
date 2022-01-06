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
	"github.com/fatih/color"
)

func (a *app) setNoColor(b bool) {
	color.NoColor = b
}

func (a *app) printGRML() {
	a.Println("               _ ")
	a.Println(" ___ ___ _____| |")
	a.Println("| . |  _|     | |")
	a.Println("|_  |_| |_|_|_|_|")
	a.Println("|___|            ")
	a.Println("")
}

func (a *app) printColor(s string) {
	color.Set(color.FgYellow)
	a.Print(s)
	color.Unset()
}

func (a *app) printColorln(s string) {
	a.printColor(s + "\n")
}
