#!/usr/bin/env bash
set -euo pipefail

readonly OAPI_CODEGEN_VERSION=v2.5.1
readonly REDOCLY_VERSION=2.46.2

go_binary=${GO_BINARY:-$(command -v go || true)}
if [[ -z "$go_binary" && -x /usr/local/go/bin/go ]]; then
  go_binary=/usr/local/go/bin/go
fi
if [[ -z "$go_binary" ]]; then
  echo "go is required" >&2
  exit 127
fi

repo_root=$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)
cd "$repo_root"

npx --yes "@redocly/cli@${REDOCLY_VERSION}" lint api/openapi.yaml
"$go_binary" run "github.com/oapi-codegen/oapi-codegen/v2/cmd/oapi-codegen@${OAPI_CODEGEN_VERSION}" \
  --config api/oapi-codegen.yaml api/openapi.yaml

git diff --exit-code -- webapi/generated.go
