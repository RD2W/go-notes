#!/usr/bin/env bash

set -e

PROJECT_ROOT="$(cd "$(dirname "$0")/.." && pwd)"
PROTO_DIR="$PROJECT_ROOT/api/proto"
GO_OUT_DIR="$PROJECT_ROOT/pkg/proto"
PROTOC_GEN_GO="$(go env GOPATH)/bin/protoc-gen-go"
PROTOC_GEN_GO_GRPC="$(go env GOPATH)/bin/protoc-gen-go-grpc"

echo "🔧 Generating Go code from protobuf definitions..."

# Проверяем наличие плагинов
if [ ! -f "$PROTOC_GEN_GO" ]; then
    echo "❌ Error: protoc-gen-go not found at $PROTOC_GEN_GO"
    echo "Please run: go install google.golang.org/protobuf/cmd/protoc-gen-go@latest"
    exit 1
fi

if [ ! -f "$PROTOC_GEN_GO_GRPC" ]; then
    echo "❌ Error: protoc-gen-go-grpc not found at $PROTOC_GEN_GO_GRPC"
    echo "Please run: go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest"
    exit 1
fi

echo "⚙️ Using protoc-gen-go: $PROTOC_GEN_GO"
echo "⚙️ Using protoc-gen-go-grpc: $PROTOC_GEN_GO_GRPC"

# Создаем директорию для сгенерированного кода
mkdir -p "$GO_OUT_DIR"

# Генерируем код для каждого protobuf файла
find "$PROTO_DIR" -name "*.proto" -print0 | while IFS= read -r -d '' proto_file; do
    echo "📦 Generating: $proto_file"

    protoc --go_out="$GO_OUT_DIR" \
           --go-grpc_out="$GO_OUT_DIR" \
           --go_opt=paths=source_relative \
           --go-grpc_opt=paths=source_relative \
           -I="$PROTO_DIR" \
           "$proto_file"
done

echo "✅ Protobuf code generation completed!"
