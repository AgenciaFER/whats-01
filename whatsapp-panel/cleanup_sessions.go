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
	sessionDir := "storage/sessions"

	// Verificar se diretório existe
	info, err := os.Stat(sessionDir)
	if os.IsNotExist(err) {
		fmt.Println("Diretório de sessões não existe, criando...")
		if err := os.MkdirAll(sessionDir, 0755); err != nil {
			log.Fatalf("Erro ao criar diretório: %v", err)
		}
		fmt.Println("Diretório criado com sucesso.")
		return
	}

	if !info.IsDir() {
		log.Fatalf("O caminho %s não é um diretório", sessionDir)
	}

	// Ler arquivos no diretório
	files, err := os.ReadDir(sessionDir)
	if err != nil {
		log.Fatalf("Erro ao ler diretório: %v", err)
	}

	if len(files) == 0 {
		fmt.Println("Não há sessões para limpar.")
		return
	}

	fmt.Printf("Encontradas %d sessões para remover.\n", len(files))

	// Remover cada arquivo
	for _, file := range files {
		if file.IsDir() {
			continue
		}

		// Verificar se é um arquivo de sessão (.db)
		if filepath.Ext(file.Name()) != ".db" {
			continue
		}

		fullPath := filepath.Join(sessionDir, file.Name())
		if err := os.Remove(fullPath); err != nil {
			fmt.Printf("Erro ao remover %s: %v\n", fullPath, err)
		} else {
			fmt.Printf("Sessão removida: %s\n", file.Name())
		}
	}

	fmt.Println("Limpeza concluída.")
}
