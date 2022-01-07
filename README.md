# grml - A simple build automation tool written in Go

grml is a simple Makefile alternative. Build targets are defined in a `grml.yaml` file located in the project's root directory.
This file uses the [YAML](http://yaml.org/) syntax.

A minimal sample can be found within the [sample](sample/grml.yaml) directory. Enter the directory with a terminal and execute `grml`.

[![asciicast](https://asciinema.org/a/I3AhdfrND2CtC4v0jODayQQKP.svg)](https://asciinema.org/a/I3AhdfrND2CtC4v0jODayQQKP)

## Installation
### From Source
    go install github.com/desertbit/grml@latest

### Prebuild Binaries
https://github.com/desertbit/grml/releases

## Specification
- Environment variables can be defined in the **env** section. These variables are passed to all run target processes.
- Variables are also accessible with the `${}` selector within **help** messages and **import** statements.
- Dependencies can be specified within the command's **deps** section.

### Additonal Environment Variables

The process environment is inherited and following additonal variables are set:

| KEY     | VALUE                                                          |
|:--------|:---------------------------------------------------------------|
| ROOT    | Path to the root build directory containing the grml.yaml file |
| PROJECT | Project name as specified within the grml file                 |
| NUMCPU  | Number of CPU cores                                            |