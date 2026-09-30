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
	"reflect"
	"testing"

	"github.com/desertbit/grml/internal/cmd"
	"github.com/desertbit/grml/internal/manifest"
)

func TestExtractOptionFlags(t *testing.T) {
	rest, flags := extractOptionFlags([]string{
		"-v", "-d", "dir", "--nocolor", "--debug=true", "--file=x.yaml", "-x=1",
		"release", "--dryrun", "--channel=", "--5", "tag", "--", "--keep=me",
	})

	wantRest := []string{"-v", "-d", "dir", "--nocolor", "--file=x.yaml", "-x=1", "release", "--5", "tag", "--", "--keep=me"}
	if !reflect.DeepEqual(rest, wantRest) {
		t.Errorf("rest = %q, want %q", rest, wantRest)
	}

	wantFlags := []optionFlag{
		{name: "debug", value: "true", hasValue: true, pos: 4},
		{name: "dryrun", pos: 7},
		{name: "channel", value: "", hasValue: true, pos: 7},
	}
	if !reflect.DeepEqual(flags, wantFlags) {
		t.Errorf("flags = %+v, want %+v", flags, wantFlags)
	}
}

func newSampleApp(t *testing.T) *app {
	t.Helper()
	m, err := manifest.Parse("../../sample/grml.yaml")
	if err != nil {
		t.Fatal(err)
	}
	a := &app{manifest: m}
	a.options, err = m.ParseOptions()
	if err != nil {
		t.Fatal(err)
	}
	a.commands, err = cmd.ParseManifest(m)
	if err != nil {
		t.Fatal(err)
	}
	return a
}

func TestApplyOptionFlags(t *testing.T) {
	cases := []struct {
		name    string
		args    []string
		wantErr bool
		check   func(t *testing.T, a *app)
	}{
		{
			name: "root options before command",
			args: []string{"--debug=true", "--runopts=mars", "deploy", "host", "user"},
			check: func(t *testing.T, a *app) {
				if !a.options[""].Bools["debug"].Value {
					t.Error("debug not set")
				}
				if o := a.options[""].Choices["runopts"]; o.Active != "mars" || !o.UserSet {
					t.Errorf("runopts = %+v", o)
				}
			},
		},
		{
			name: "scoped options after include command",
			args: []string{"--debug=true", "release", "--dryrun=true", "--channel=beta", "tag"},
			check: func(t *testing.T, a *app) {
				if !a.options[""].Bools["debug"].Value {
					t.Error("debug not set")
				}
				if !a.options["release"].Bools["dryrun"].Value {
					t.Error("dryrun not set")
				}
				if a.options["release"].Choices["channel"].Active != "beta" {
					t.Error("channel not set")
				}
			},
		},
		{
			name: "outer options visible inside scope, bare bool, app flags skipped",
			args: []string{"-v", "-d", ".", "release", "tag", "--runopts=moon", "--dryrun"},
			check: func(t *testing.T, a *app) {
				if a.options[""].Choices["runopts"].Active != "moon" {
					t.Error("runopts not set")
				}
				if !a.options["release"].Bools["dryrun"].Value {
					t.Error("dryrun not set")
				}
			},
		},
		{name: "scoped option before its command", args: []string{"--dryrun=true", "release", "tag"}, wantErr: true},
		{name: "scoped option after positional args", args: []string{"deploy", "release", "--dryrun"}, wantErr: true},
		{name: "invalid choice", args: []string{"release", "--channel=foo"}, wantErr: true},
		{name: "choice without value", args: []string{"--runopts"}, wantErr: true},
		{name: "invalid bool", args: []string{"--debug=nah"}, wantErr: true},
		{name: "unknown option", args: []string{"--nope=1"}, wantErr: true},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			a := newSampleApp(t)
			rest, flags := extractOptionFlags(tc.args)
			err := a.applyOptionFlags(rest, flags)
			if tc.wantErr {
				if err == nil {
					t.Fatal("expected error")
				}
				return
			}
			if err != nil {
				t.Fatal(err)
			}
			tc.check(t, a)
		})
	}
}

func TestCheckOptionNames(t *testing.T) {
	a := newSampleApp(t)
	if err := a.checkOptionNames(); err != nil {
		t.Fatalf("sample options: %v", err)
	}

	for _, raw := range []map[string]interface{}{
		{"verbose": true},
		{"file": []interface{}{"a", "b"}},
		{"help": false},
	} {
		a := newSampleApp(t)
		if err := a.options["release"].Add(raw); err != nil {
			t.Fatal(err)
		}
		if err := a.checkOptionNames(); err == nil {
			t.Errorf("options %v: expected error", raw)
		}
	}
}
