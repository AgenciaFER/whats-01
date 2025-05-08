package config

import (
	"os"
	"path/filepath"

	"github.com/joho/godotenv"
)

// Config contém todas as configurações da aplicação
type Config struct {
	Port         string
	DatabasePath string
	StoreDir     string
}

// LoadConfig carrega as configurações do ambiente
func LoadConfig() (*Config, error) {
	// Carregar variáveis de ambiente do arquivo .env se existir
	godotenv.Load()

	// Configurar diretório para armazenamento
	storeDir := os.Getenv("STORE_DIR")
	if storeDir == "" {
		userHomeDir, err := os.UserHomeDir()
		if err != nil {
			return nil, err
		}
		storeDir = filepath.Join(userHomeDir, ".whatsapp-panel")
	}

	// Criar diretório se não existir
	if _, err := os.Stat(storeDir); os.IsNotExist(err) {
		if err := os.MkdirAll(storeDir, 0755); err != nil {
			return nil, err
		}
	}

	// Configurar caminho do banco de dados
	dbPath := os.Getenv("DB_PATH")
	if dbPath == "" {
		dbPath = filepath.Join(storeDir, "whatsapp.db")
	}

	// Obter porta do servidor
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	return &Config{
		Port:         port,
		DatabasePath: dbPath,
		StoreDir:     storeDir,
	}, nil
}
