/*
 *  Grumble - A simple build automation tool written in Go
 *  Copyright (C) 2016  Roland Singer <roland.singer[at]desertbit.com>
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
	"os"
	"path/filepath"
	"strings"

	"github.com/desertbit/grml/context"
	"github.com/desertbit/grml/spec"
)

const (
	specFilename = "grml.yaml"
)

func main() {
	var err error

	// Our build context.
	ctx := &context.Context{
		Env:         os.Environ(), // Inherit the current process environment.
		PrintTarget: printTarget,
	}

	// Set the initial base path to the current working dir.
	ctx.BasePath, err = os.Getwd()
	if err != nil {
		fatalErr(fmt.Errorf("failed to obtain the current working directory: %v", err))
	}

	// Parse the command line arguments.
	err = parseArgs(ctx)
	if err != nil {
		fatalErr(err)
	}

	// Get the absolute path.
	ctx.BasePath, err = filepath.Abs(ctx.BasePath)
	if err != nil {
		fatalErr(err)
	}

	// Set the PWD environment variable.
	ctx.Env = append(ctx.Env, "PWD="+ctx.BasePath)

	// Transform the context environment variables to a map.
	ctxEnv := make(map[string]string)
	for _, e := range ctx.Env {
		p := strings.Index(e, "=")
		if p > 0 {
			ctxEnv[e[0:p]] = e[p+1:]
		}
	}

	// Read the specification file.
	specPath := filepath.Join(ctx.BasePath, specFilename)
	spec, err := spec.ParseSpec(specPath, ctxEnv)
	if err != nil {
		fatalErr(fmt.Errorf("spec file: %v", err))
	}

	// Merge the environment variables.
	ctx.Env = append(ctx.Env, spec.EnvToSlice()...)

	// Set the default target if no targets were passed.
	if len(ctx.Targets) == 0 {
		dt := spec.DefaultTarget()
		if dt != nil {
			ctx.Targets = append(ctx.Targets, dt.Name())
		}
	}

	// Print all targets if required.
	if ctx.OnlyPrintAllTargets || len(ctx.Targets) == 0 {
		printTargets(spec)
		return
	}

	// Check if the passed targets are valid.
	for _, t := range ctx.Targets {
		tt := spec.Targets[t]
		if tt == nil {
			fatalErr(fmt.Errorf("target does not exists: %s", t))
		}
	}

	// Run the targets.
	err = ctx.RunTargets(spec)
	if err != nil {
		fatalErr(err)
	}

	printDone()
}
