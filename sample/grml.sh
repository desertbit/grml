#!/bin/sh

# Ensure directories always exists.
mkdir -p "${BUILDDIR}" "${BINDIR}"

go_build() {
    local opts
    if [ "$debug" = true ]; then
        opts+="-gcflags=\"all=-N -l\""
    fi
   
    go build $opts -o "${BINDIR}/${DESTBIN}"
}
