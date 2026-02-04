#!/usr/bin/env bash
set -euo pipefail

go test ./...
(cd examples/basic && go build ./...)
