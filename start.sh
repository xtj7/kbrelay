#!/bin/bash
DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" >/dev/null 2>&1 && pwd )"
stty -echo
sudo go run $DIR/kbrelay.go "$@"
stty echo