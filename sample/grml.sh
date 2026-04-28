#!/bin/bash

# Sourced before every exec body via the manifest's 'import:' list.
# Set up shared directories and helper functions here.

mkdir -p "${BUILDDIR}" "${BINDIR}"

go_build() {
    local opts=()
    if grml_option debug; then
        opts+=(-gcflags="all=-N -l")
    fi

    go build "${opts[@]}" -o "${BINDIR}/${DESTBIN}"
}
