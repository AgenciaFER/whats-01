package main

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

func main() {
	// Diretório base dos templates
	templateDir := "web/templates/unified"

	// Criar diretório de backup se não existir
	backupDir := filepath.Join(templateDir, "backup")
	os.MkdirAll(backupDir, 0755)

	// Processar todos os templates
	files, err := os.ReadDir(templateDir)
	if err != nil {
		fmt.Printf("Erro ao ler diretório de templates: %v\n", err)
		return
	}

	for _, file := range files {
		if file.IsDir() || !strings.HasSuffix(file.Name(), ".html") {
			continue
		}

		// Ler arquivo original
		filePath := filepath.Join(templateDir, file.Name())
		content, err := os.ReadFile(filePath)
		if err != nil {
			fmt.Printf("Erro ao ler arquivo %s: %v\n", filePath, err)
			continue
		}

		// Fazer backup
		backupPath := filepath.Join(backupDir, file.Name())
		os.WriteFile(backupPath, content, 0644)

		// Corrigir conteúdo
		contentStr := string(content)

		// 1. Remover definições aninhadas
		contentStr = removeNestedDefines(contentStr)

		// 2. Simplificar definições restantes
		contentStr = simplifyDefines(contentStr)

		// Escrever conteúdo corrigido
		err = os.WriteFile(filePath, []byte(contentStr), 0644)
		if err != nil {
			fmt.Printf("Erro ao escrever arquivo %s: %v\n", filePath, err)
			continue
		}

		fmt.Printf("Corrigido: %s\n", filePath)
	}

	fmt.Println("Correção de templates concluída com sucesso!")
}

// Remove defines aninhados que podem causar problemas
func removeNestedDefines(content string) string {
	// Identificar e corrigir padrões problemáticos comuns
	result := content

	// 1. Remover definições aninhadas do tipo {{ define "X" }}{{ template "Y" . }}{{ end }}
	nestedDefinePattern := `{{ ?define "([^"]+)" ?}}{{ ?template "([^"]+)" [^ ]* ?}}{{ ?end ?}}`
	result = regexp.MustCompile(nestedDefinePattern).ReplaceAllString(result, "")

	return result
}

// Simplifica as definições de template
func simplifyDefines(content string) string {
	result := content

	// 1. Converter {{ define "X" }} para {{ template "X" . }} quando necessário
	selfReferencePattern := `{{ ?define "([^"]+)" ?}}{{ ?template "(\1)" [^ ]* ?}}`
	result = regexp.MustCompile(selfReferencePattern).ReplaceAllString(result, "{{ template \"$1\" . }}")

	return result
}
