//go:build ignore
// +build ignore

package main

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
)

func main() {
	// Limpar todas as sessions e sessions_old
	dirs := []string{"storage/sessions", "storage/sessions_old"}

	for _, dir := range dirs {
		// Verificar se diretório existe
		info, err := os.Stat(dir)
		if os.IsNotExist(err) {
			fmt.Printf("Diretório %s não existe, criando...\n", dir)
			if err := os.MkdirAll(dir, 0755); err != nil {
				log.Fatalf("Erro ao criar diretório %s: %v", dir, err)
			}
			fmt.Printf("Diretório %s criado com sucesso.\n", dir)
			continue
		}

		if !info.IsDir() {
			log.Fatalf("O caminho %s não é um diretório", dir)
		}

		// Ler arquivos no diretório
		files, err := os.ReadDir(dir)
		if err != nil {
			log.Fatalf("Erro ao ler diretório %s: %v", dir, err)
		}

		if len(files) == 0 {
			fmt.Printf("Não há sessões para limpar em %s.\n", dir)
			continue
		}

		fmt.Printf("Encontradas %d sessões para remover em %s.\n", len(files), dir)

		// Remover cada arquivo
		for _, file := range files {
			if file.IsDir() {
				continue
			}

			// Verificar se é um arquivo de sessão (.db)
			if filepath.Ext(file.Name()) != ".db" {
				continue
			}

			fullPath := filepath.Join(dir, file.Name())
			if err := os.Remove(fullPath); err != nil {
				fmt.Printf("Erro ao remover %s: %v\n", fullPath, err)
			} else {
				fmt.Printf("Sessão removida: %s\n", file.Name())
			}
		}
	}

	// Limpar também outros arquivos de banco de dados WhatsApp
	for _, file := range []string{"connect.db", "contacts.db", "whatsapp.db"} {
		if _, err := os.Stat(file); err == nil {
			if err := os.Remove(file); err != nil {
				fmt.Printf("Erro ao remover %s: %v\n", file, err)
			} else {
				fmt.Printf("Arquivo removido: %s\n", file)
			}
		}
	}

	fmt.Println("Limpeza completa concluída.")
}
