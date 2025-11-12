#!/usr/bin/env bash

set -e

PROJECT_ROOT="$(cd "$(dirname "$0")/.." && pwd)"
MAIN_GO_FILE="cmd/web-server/main.go"
SWAGGER_DOCS_DIR="$PROJECT_ROOT/docs"

echo "🔧 Generating Swagger documentation..."

# Проверяем наличие swag
if ! command -v swag &> /dev/null; then
    echo "❌ Error: swag not found"
    echo "Please install swag by running: go install github.com/swaggo/swag/cmd/swag@latest"
    exit 1
fi

echo "⚙️ Using swag: $(which swag)"

# Проверяем наличие главного Go файла
if [ ! -f "$MAIN_GO_FILE" ]; then
    echo "❌ Error: Main Go file not found at $MAIN_GO_FILE"
    exit 1
fi

# Создаем директорию для сгенерированного кода
mkdir -p "$SWAGGER_DOCS_DIR"

echo "📦 Generating Swagger docs from: $MAIN_GO_FILE"

# Генерируем Swagger документацию
cd "$PROJECT_ROOT" && swag init -g "$MAIN_GO_FILE" --output "docs"

echo "✅ Swagger documentation generation completed!"
