# grml - A simple build automation tool written in Go

grml is a simple Makefile alternative. Build targets are defined in a `grml.yaml`
file located in the project's root directory.
This file uses the [YAML](http://yaml.org/) syntax.

A minimal `grml.yaml` file can be expressed as follows:

```yaml
targets:
    app:
        help:    build the app
        default: true
        run: |
            gb build all
```

The build is triggered with:

```
$ grml app
```

or just with the default target:

```
$ grml
```


The run section is called in a shell (sh) process. All sh expressions (`if`, `elif`, ...) are valid.

## Outputs

For each target multiple outputs can be defined. Targets are skipped if the output files exist.

```yaml
targets:
    resources:
        help: build the resources
        output:
            - build/resources
        run: |
            mkdir -p build
            touch build/resources
```

## Dependencies

Dependencies can be specified within the **deps** section.

```yaml
targets:
    app:
        help:    build the app
        deps:
            - resources
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
            - resources
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

### Additonal Variables

The process environment is inherited and following additonal variables are set:

| KEY  | VALUE                                                          |
|:-----|:---------------------------------------------------------------|
| ROOT | Path to the root build directory containing the grml.yaml file |


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
            - resources
            - db
        run: |
            echo "building app ${destBin}"
            gb build all

    resources:
        help:        build the resources
        help-group:  Resources
        deps:
            - images
        output:
            - ${buildDir}/resources
        run: |
            mkdir -p ${buildDir}
            touch ${buildDir}/resources

    images:
        help:        build the image resources
        help-group:  Resources
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
