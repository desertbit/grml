#!/bin/bash

# Sourced only when running commands inside this include's subtree.
# Sees the per-include env (DESTBIN, RELEASE_NOTE) plus root env.

release_banner() {
    echo "=== ${RELEASE_NOTE} ==="
}
