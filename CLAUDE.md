# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project

`grml` is a Makefile-alternative build automation tool. Targets are declared in a `grml.yaml` file at the project root; running `grml` opens an interactive shell (built on `desertbit/grumble`) that exposes those targets as commands.

## Common commands

```sh
# Build & run locally
go build ./...
go run .                          # opens the grml shell against ./grml.yaml
go run . -d /path/to/project      # run against a different project
go run . <target>                 # one-shot: run a target then exit

# Cross-compile release binaries (uses Docker, see ./grml.yaml)
go run . build                    # builds linux-amd64 and win-amd64 into ./bin

# Module hygiene
go mod tidy
```

There is no test suite in this repo — no `_test.go` files exist. Don't claim "tests pass" in summaries; there is nothing to run.

The repo's own `grml.yaml` declares `version: 1`, but `manifest.Version` is `2` and `Parse` rejects mismatches. The repo file is effectively stale for self-hosting; treat `sample/grml.yaml` as the canonical example of current syntax (note: it also says `version: 1` and would need to be `2` to actually run).

## Architecture

Entry point `grml.go` calls `internal/app.Run()`. From there, the code splits into four internal packages with a clear pipeline:

1. **`internal/manifest`** — YAML schema + parser. `Parse()` reads `grml.yaml`, validates the version against `manifest.Version`, then recursively resolves `include:` directives (a command can pull its definition from another YAML file via `parseIncludes`). The schema supports nested `commands`, `deps`, `args`, `env`, `options`, an `interpreter` (`sh` or `bash`), and `import` (shell files sourced before each exec).

2. **`internal/options`** — runtime-mutable user options. Two kinds: `Bools` (toggleable check options) and `Choices` (pick-one-of-N). `Restore()` carries values across a `reload` so options survive re-reading the manifest. For choices, `UserSet` tracks whether the user explicitly picked via `options set`; `Restore` only carries the prior active when `UserSet` is true, so editing the YAML to prepend a new option lets the new file's first item become active (issue #10).

3. **`internal/cmd`** — flattens the manifest's nested command tree into `cmd.Commands` with dotted paths (e.g. `build.linux-amd64`) and resolves `deps:` strings into pointers to other `*Command`s. Dep path syntax: `foo.bar` is absolute from root; leading `.` is relative to the command's own path (`.bar` → `selfPath.bar`); leading `~.` is relative to the command's `origin` — the path of the nearest enclosing `include` point (or root if none) — set during `addCommands` whenever `mc.Include != ""`. The `~` form requires the dot (e.g. `~.bar`); bare `~bar` errors out. **Constraint enforced here:** dependency commands cannot have `args`.

4. **`internal/app`** — wires everything to `grumble`. `app.load()` (in `app.go`) is the central reload step: clears commands, parses the manifest, builds env (process env + `ROOT`/`PROJECT`/`NUMCPU` + manifest `env` with `${VAR}` interpolation via `EvalEnv`), parses options, then registers grumble commands recursively in `registerCommands`.

### Execution model (`internal/app/exec.go`)

When a command runs:
- An `execContext` tracks already-run targets so a dep shared by multiple commands runs at most once per top-level invocation.
- Deps run depth-first before the command's own `exec` block.
- `runShellCommand` constructs a script: `set -e` (always), then the `grml_*` shell builtins from `internal/app/builtins.go`, then `set -x` (when verbose) — builtins are emitted *before* `set -x` on purpose so their definitions don't pollute the trace, then `. "${ROOT}/<file>"` for each manifest `import`, then the user's `exec` body. This is piped to `sh -c` or `bash -c` based on `manifest.Interpreter`. Builtins must stay POSIX-compatible because `sh` is the default interpreter.
- The child process inherits `os.Environ()` plus manifest env, plus each option as `KEY=value` (bools as `true`/`false`, choices as the active string), plus per-invocation `args`. Args are placed first in the env slice so later entries can shadow them — relevant if a target arg name collides with an env var.

### Variable interpolation — two distinct passes

- **`manifest.EvalEnvSlice`** interpolates `${VAR}` for a single env scope: each entry expands against earlier entries in the same scope, then against the parent env. `Manifest.EvalEnv` delegates to it for the root scope. `env:` is a `yaml.MapSlice` to preserve declaration order.
- **`app.evalVar(env, str)`** interpolates `${VAR}` in help strings and import paths. Takes the env map explicitly: callers pass `a.env` for root-only contexts (imports) and `a.cmdEnv(c)` for per-command contexts (help strings). **Options are not available here** — only env values. If you see a help string referencing an option name, it will not expand.

The `exec` body itself is *not* pre-interpolated by Go — variables are expanded by the shell at runtime via the `env` passed to `exec.Command`.

### Per-include env

An `include`d subgrml file can declare an `env:` block at its top. Those values are scoped to the commands defined inside that file, layered over root env, and not visible to commands outside the file.

Implementation note: the `Env` field lives on `manifest.Command`, so technically *any* command (not just the include point) can carry `env:`, and scopes nest. The user-facing framing is per-include because that's the intended use case. The scope chain is built statically during cmd flattening (`cmd.addCommands` carries `parentEnvs []yaml.MapSlice`) and stored on each `cmd.Command` as `envs` — outermost first, command's own scope last. `app.cmdEnv(c)` layers this chain on top of `a.env` at use time via `manifest.EvalEnvSlice`. Scoping is by command location, not by call chain: when a dep crosses scope boundaries, each command runs with *its own* layered env. The root manifest's `env:` is merged into `a.env` directly since it applies universally.

### Built-in commands added at load

`reload`, `verbose <bool>`, `options`, `options check`, `options set <name>` are injected by `app.load()` and `initOptions()`; they're grouped under "Builtins:" in help. User commands from the manifest live in the default group.
