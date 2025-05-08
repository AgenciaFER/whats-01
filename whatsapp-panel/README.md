# WhatsApp Panel

Painel de gerenciamento de múltiplas sessões do WhatsApp usando a biblioteca whatsmeow.

## Funcionalidades

- ✨ Interface web moderna e responsiva
- 🔄 Gerenciamento de múltiplas sessões
- 📊 Estatísticas em tempo real
- 🔒 Desconexão segura de sessões
- 📱 Conexão via QR Code
- 💾 Persistência de dados com SQLite

## Requisitos

- Go 1.20 ou superior
- SQLite 3
- Navegador moderno com suporte a JavaScript

## Instalação

1. Clone o repositório:
```bash
git clone https://github.com/seu-usuario/whatsapp-panel
cd whatsapp-panel
```

2. Instale as dependências:
```bash
go mod download
```

3. Configure as variáveis de ambiente:
```bash
cp env.example .env
```
Edite o arquivo `.env` com suas configurações.

## Uso

1. Inicie o servidor:
```bash
go run cmd/server/main.go
```

2. Acesse o painel no navegador:
```
http://localhost:8080
```

3. Clique em "Conectar WhatsApp" para adicionar uma nova sessão.

4. Escaneie o QR Code com seu WhatsApp para conectar.

## Estrutura do Projeto

```
whatsapp-panel/
├── cmd/
│   └── server/
│       └── main.go          # Ponto de entrada da aplicação
├── internal/
│   ├── config/             # Configuração da aplicação
│   ├── handlers/           # Handlers HTTP
│   ├── models/            # Modelos de dados
│   ├── services/          # Lógica de negócio
│   └── storage/           # Camada de persistência
└── web/                   # Interface web
```

## Desenvolvimento

Para executar em modo de desenvolvimento:

```bash
DEBUG=true go run cmd/server/main.go
```

## Licença

Este projeto está licenciado sob a licença MIT - veja o arquivo [LICENSE](LICENSE) para mais detalhes.