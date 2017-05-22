# Grumble - A simple build automation tool written in Go

Grumble is a simple Makefile alternative. Build targets are defined in a `GRUMBLE`
file located in the project's root directory.
This file uses the [YAML](http://yaml.org/) syntax.

A minimal `GRUMBLE` file can be expressed as follows:

```yaml
targets:
    app:
        help: build the app
        run: |
            gb build all
```

The build is triggered with:

```
$ grumble app
```

The run section is called in a shell (sh) process. All sh expressions (`if`, `elif`, ...) are valid.

## Dependencies

For each target multiple outputs can be defined. These outputs can be used as dependencies
for other build targets. Targets are skipped if no build is required.

```yaml
targets:
    app:
        help:    build the app
        default: true
        deps:
            - build/resources
        run: |
            gb build all

    resources:
        help: build the resources
        output:
            - build/resources
        run: |
            mkdir -p build
            touch build/resources
```

## Variables

Environment variables can be defined in the **env** section. These variables are passed to all run target processes.

```yaml
env:
    version: 1.0.0

targets:
    app:
        help:    build the app
        default: true
        run: |
            echo "$version"
```

Variables are accessible with the `${}` selector in the **env**, **deps** and **output** section.

```yaml
env:
    version:    1.0.0
    buildDir:   build/
    destBin:    app-${version}

targets:
    app:
        help: build the app
        deps:
            - ${buildDir}/resources
        run: |
            echo "building app ${destBin}"
            gb build all

    resources:
        help: build the resources
        output:
            - ${buildDir}/resources
        run: |
            mkdir -p ${buildDir}
            touch ${buildDir}/resources
```

## Final Example

```yaml
env:
    version:    1.0.0
    buildDir:   build/
    destBin:    app-${version}

targets:
    app:
        help:    build the app
        default: true
        deps:
            - ${buildDir}/resources
            - ${buildDir}/db
        run: |
            echo "building app ${destBin}"
            gb build all

    resources:
        help: build the resources
        deps:
            - ${buildDir}/images
        output:
            - ${buildDir}/resources
        run: |
            mkdir -p ${buildDir}
            touch ${buildDir}/resources

    images:
        help: build the image resources
        output:
            - ${buildDir}/images
        run: |
            mkdir -p ${buildDir}
            touch ${buildDir}/images

    db:
        help: build the database files
        output:
            - ${buildDir}/db
        run: |
            mkdir -p ${buildDir}
            touch ${buildDir}/db
```
