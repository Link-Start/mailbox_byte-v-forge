#!/usr/bin/env sh
set -eu

ROOT="$(CDPATH= cd -- "$(dirname -- "$0")/.." && pwd)"
PATH="$(go env GOPATH)/bin:$PATH"

rm -rf "$ROOT/services/mailbox-api/pb" "$ROOT/services/mailbox-api/internal/contracts"
mkdir -p "$ROOT/services/mailbox-api/pb" "$ROOT/services/mailbox-api/internal/contracts"

protoc -I "$ROOT/proto" \
  --go_out="$ROOT/services/mailbox-api" \
  --go_opt=module=mailboxapi \
  "$ROOT/proto/byte/v/forge/contracts/common/v1/common.proto" \
  "$ROOT/proto/byte/v/forge/contracts/common/v1/eventbus.proto" \
  "$ROOT/proto/byte/v/forge/contracts/mailbox/v1/mailbox.proto" \
  "$ROOT/proto/byte/v/forge/contracts/observability/v1/hotstream.proto" \
  "$ROOT/proto/byte/v/forge/contracts/browserautomation/v1/browser_automation.proto"

protoc -I "$ROOT/proto" \
  --go-grpc_out="$ROOT/services/mailbox-api" \
  --go-grpc_opt=module=mailboxapi \
  "$ROOT/proto/byte/v/forge/contracts/browserautomation/v1/browser_automation.proto"

protoc -I "$ROOT/proto" \
  --go_out="$ROOT/services/mailbox-api/pb" \
  --go-grpc_out="$ROOT/services/mailbox-api/pb" \
  "$ROOT/proto/email.proto" \
  "$ROOT/proto/mailbox_register.proto" \
  "$ROOT/proto/mailbox_commands.proto" \
  "$ROOT/proto/mailbox_service.proto"
