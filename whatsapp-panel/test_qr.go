//go:build ignore
// +build ignore

package main

import (
	"context"
	"fmt"
	"os"
	"time"

	"whatsapp-panel/internal/services/whatsapp"
	"whatsapp-panel/internal/storage"
)

func main() {
	// Criar banco de dados de teste
	db, err := storage.NewDatabase("test_qr.db")
	if err != nil {
		fmt.Println("Erro ao abrir banco de dados:", err)
		os.Exit(1)
	}
	defer db.Close()

	// Criar Manager do WhatsApp
	mgr := whatsapp.NewManager(db)

	// Criar novo cliente
	client, err := mgr.NewClient()
	if err != nil {
		fmt.Println("Erro ao criar cliente WhatsApp:", err)
		os.Exit(1)
	}

	// Contexto de 2 minutos para gerar QR
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()

	// Obter canal de QR Code
	qrChan, err := client.GetQRChannel(ctx)
	if err != nil {
		fmt.Println("Erro ao obter canal de QR Code:", err)
		os.Exit(1)
	}

	// Ler primeiro código do canal
	select {
	case code := <-qrChan:
		fmt.Println("QR Code capturado (base64):")
		fmt.Println(code)
	case <-ctx.Done():
		fmt.Println("Timeout ao aguardar QR Code")
		os.Exit(1)
	}
}
