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

// grmlBuiltins is a POSIX-compatible shell snippet sourced before every
// command's exec body. It defines helpers under the grml_* namespace.
//
// grml_option <name> : exit 0 if the named option equals "true"
// grml_option <name> <value> : exit 0 if the named option equals <value>
// grml_if <name> <if-str> <else-str> : print if-str when the option is true, else else-str
// grml_if <name> <value> <if-str> <else-str> : print if-str when option equals value, else else-str
const grmlBuiltins = `
grml_option() {
    case $# in
        1) eval "[ \"\${$1-}\" = true ]" ;;
        2) eval "[ \"\${$1-}\" = \"\$2\" ]" ;;
        *) echo "grml_option: usage: grml_option <name> [value]" >&2; return 2 ;;
    esac
}

grml_if() {
    case $# in
        3)
            if grml_option "$1"; then echo "$2"; else echo "$3"; fi
            ;;
        4)
            if grml_option "$1" "$2"; then echo "$3"; else echo "$4"; fi
            ;;
        *)
            echo "grml_if: usage: grml_if <name> [<value>] <if-str> <else-str>" >&2
            return 2
            ;;
    esac
}
`
