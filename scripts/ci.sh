#!/usr/bin/env bash
set -euo pipefail

go test ./...
(cd examples/basic && go build -o /tmp/nope-example-basic .)
(cd examples/facade && go build -o /tmp/nope-example-facade .)
