# grml — A simple build automation tool written in Go

`grml` is a Makefile alternative. Build targets are declared in a `grml.yaml` file at the project root, using [YAML](http://yaml.org/) syntax. Running `grml` opens an interactive shell that exposes those targets as commands; passing a target on the command line runs it once and exits.

A worked example lives in [sample/](sample/) — `cd sample && grml`.

[![asciicast](https://asciinema.org/a/460524.svg)](https://asciinema.org/a/460524)

## Installation

### From source
    go install github.com/desertbit/grml@latest

### Prebuilt binaries
https://github.com/desertbit/grml/releases

## Usage

```
grml [flags] [command...]
```

| Flag              | Description                                                          |
|:------------------|:---------------------------------------------------------------------|
| `-d, --directory` | root directory containing the grml file (default: current directory) |
| `-f, --file`      | grml file relative to the root (default: `grml.yaml`)                |
| `-v, --verbose`   | trace each shell command as it runs (`set -x`)                       |

The `-f` flag lets you keep multiple manifests side by side — e.g. `grml.yaml` for in-container work and `grml.host.yaml` for tasks that must run on the host.

Without a target, `grml` drops into an interactive shell with tab completion. Built-in commands:

| Command              | Description                                          |
|:---------------------|:-----------------------------------------------------|
| `reload`             | re-read the grml file (preserves option values)      |
| `verbose <bool>`     | toggle verbose mode at runtime                       |
| `options`            | print current option values                          |
| `options check`      | toggle bool options interactively                    |
| `options set <name>` | pick a value for a choice option                     |

## Manifest reference

### Top-level keys

| Key           | Description                                                              |
|:--------------|:-------------------------------------------------------------------------|
| `version`     | manifest schema version, currently `2` (required)                        |
| `project`     | project name, exposed as `${PROJECT}` (required)                         |
| `env`         | ordered map of environment variables, supporting `${VAR}` interpolation  |
| `options`     | user-tweakable options: bools (check) or lists of strings (single choice) |
| `interpreter` | `sh` (default) or `bash`                                                  |
| `import`      | shell files sourced before every exec body                                |
| `commands`    | command tree                                                              |

### Per-command keys

| Key        | Description                                                                |
|:-----------|:---------------------------------------------------------------------------|
| `help`     | help text (supports `${VAR}` interpolation from env)                       |
| `alias`    | list of alternative names                                                  |
| `args`     | positional arguments, exposed as env vars of the same name                 |
| `deps`     | other commands to run first; dotted path, leading `.` resolves relative to the current command |
| `exec`     | shell body to run                                                          |
| `commands` | nested sub-commands                                                        |
| `include`  | load the rest of this command's definition from another YAML file          |

### Implicit environment variables

The process environment is inherited and the following are always set:

| Variable  | Value                                                        |
|:----------|:-------------------------------------------------------------|
| `ROOT`    | absolute path to the root directory containing the grml file |
| `PROJECT` | project name from the manifest                               |
| `NUMCPU`  | number of CPU cores                                          |

Each option is also exported: bools as `true`/`false`, choices as the active value. Each `args` entry is exported when the command runs.

### Variable interpolation

`${VAR}` is expanded by `grml` inside `env` values, `import` paths, and `help` strings. Inside `exec` bodies, expansion is performed by the shell at runtime — env vars, options, args, and any other shell-visible variables are all available there.
