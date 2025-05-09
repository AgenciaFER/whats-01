#!/bin/bash

# Obter o caminho absoluto do diretório do projeto
PROJECT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
SCRIPT_PATH="$PROJECT_DIR/cleanup_sessions.go"

# Verificar se o go está instalado
if ! command -v go &> /dev/null; then
    echo "Go não está instalado. Por favor, instale o Go antes de continuar."
    exit 1
fi

# Compilar o script de limpeza para ter um binário
go build -o "$PROJECT_DIR/cleanup_sessions" "$SCRIPT_PATH"
if [ $? -ne 0 ]; then
    echo "Erro ao compilar o script de limpeza."
    exit 1
fi

# Tornar o script executável
chmod +x "$PROJECT_DIR/cleanup_sessions"

# Criar entrada crontab para executar a cada 12 horas
(crontab -l 2>/dev/null || echo "") | grep -v "$PROJECT_DIR/cleanup_sessions" | cat - <(echo "0 */12 * * * cd $PROJECT_DIR && $PROJECT_DIR/cleanup_sessions >> $PROJECT_DIR/cleanup.log 2>&1") | crontab -

echo "Tarefa de limpeza configurada para executar a cada 12 horas."
echo "Para verificar, execute 'crontab -l'"
