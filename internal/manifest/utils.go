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

package manifest

import (
	"path/filepath"
	"regexp"
	"strconv"
)

var numberRegex = regexp.MustCompile(`(\d+)`)

// naturalLess compares two strings in a natural way.
func naturalLess(a, b string) bool {
	aBase := filepath.Base(a)
	bBase := filepath.Base(b)

	// Split into text and numeric segments
	aParts := numberRegex.Split(aBase, -1)
	bParts := numberRegex.Split(bBase, -1)
	aNums := numberRegex.FindAllString(aBase, -1)
	bNums := numberRegex.FindAllString(bBase, -1)

	minLen := len(aParts)
	if len(bParts) < minLen {
		minLen = len(bParts)
	}

	for i := 0; i < minLen; i++ {
		// Compare text segments first
		if aParts[i] != bParts[i] {
			return aParts[i] < bParts[i]
		}

		// If we have numeric segments at this position, compare them
		if i < len(aNums) && i < len(bNums) {
			aNum, _ := strconv.Atoi(aNums[i])
			bNum, _ := strconv.Atoi(bNums[i])
			if aNum != bNum {
				return aNum < bNum
			}
		} else if i < len(aNums) {
			return false
		} else if i < len(bNums) {
			return true
		}
	}

	return aBase < bBase
}
