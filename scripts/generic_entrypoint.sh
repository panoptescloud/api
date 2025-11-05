#!/bin/sh

# This script is just a generic entrypoint that can be used in most containers
# allowing us to pass any generic command through to it and execute. Typically
# useful for containers where the entrypoint is a specific command already,
# and we may wanna run something else.
$@