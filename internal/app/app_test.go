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
	"path/filepath"
	"testing"
)

// TestCompletePath drives the path completer against the in-tree sample
// directory, mirroring what tab completion produces when typing args at
// the grml prompt.
func TestCompletePath(t *testing.T) {
	base, err := filepath.Abs("../../sample")
	if err != nil {
		t.Fatal(err)
	}

	cases := []struct {
		name        string
		prefix      string
		mustContain []string
		mustExclude []string
	}{
		{
			name:        "empty prefix lists root entries; dirs end with slash",
			prefix:      "",
			mustContain: []string{"grml.yaml", "grml.host.yaml", "grml.sh", "go.mod", "sample.go", "commands/"},
		},
		{
			name:        "letter prefix narrows by name",
			prefix:      "g",
			mustContain: []string{"grml.yaml", "grml.host.yaml", "grml.sh", "go.mod"},
			mustExclude: []string{"sample.go", "commands/"},
		},
		{
			name:        "trailing slash descends into directory",
			prefix:      "commands/",
			mustContain: []string{"commands/release.yaml", "commands/release.sh", "commands/notes.txt"},
		},
		{
			name:        "nested partial match",
			prefix:      "commands/notes",
			mustContain: []string{"commands/notes.txt"},
			mustExclude: []string{"commands/release.yaml"},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := completePath(tc.prefix, base)
			set := make(map[string]bool, len(got))
			for _, g := range got {
				set[g] = true
			}
			for _, w := range tc.mustContain {
				if !set[w] {
					t.Errorf("prefix=%q: missing %q in result %v", tc.prefix, w, got)
				}
			}
			for _, w := range tc.mustExclude {
				if set[w] {
					t.Errorf("prefix=%q: unexpected %q in result %v", tc.prefix, w, got)
				}
			}
		})
	}
}
