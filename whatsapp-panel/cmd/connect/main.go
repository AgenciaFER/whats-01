package main

import (
	"context"
	"fmt"
	"os"
	"time"

	"whatsapp-panel/internal/services/whatsapp"
	"whatsapp-panel/internal/storage"

	qrcode "github.com/skip2/go-qrcode"
)

func main() {
	// Inicializar banco de dados para sessão de conexão
	db, err := storage.NewDatabase("connect.db")
	if err != nil {
		fmt.Println("Erro ao abrir banco de dados:", err)
		os.Exit(1)
	}
	defer db.Close()

	// Criar gerenciador e cliente WhatsApp
	mgr := whatsapp.NewManager(db)
	client, err := mgr.NewClient()
	if err != nil {
		fmt.Println("Erro ao criar cliente WhatsApp:", err)
		os.Exit(1)
	}
	// Debug: Printar todos os eventos do cliente
	client.WAClient.AddEventHandler(func(evt interface{}) {
		fmt.Printf("[EVENT_DEBUG] %T: %+v\n", evt, evt)
	})

	// Contexto para geração de QR Code (2 minutos)
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()

	// Obter canal de QR Code
	qrChan, err := client.GetQRChannel(ctx)
	if err != nil {
		fmt.Println("Erro ao obter canal de QR Code:", err)
		os.Exit(1)
	}

	// Obter e salvar QR Code em imagem PNG
	if code, ok := <-qrChan; ok {
		err := qrcode.WriteFile(code, qrcode.Medium, 256, "qr.png")
		if err != nil {
			fmt.Println("Erro ao salvar QR Code em arquivo:", err)
			os.Exit(1)
		}
		fmt.Println("QR Code salvo em qr.png. Abra a imagem e escaneie usando o WhatsApp.")
	} else {
		fmt.Println("Erro: não foi possível obter o QR Code")
		os.Exit(1)
	}

	// Aguardar confirmação do pareamento
	maxWait := 30 * time.Second
	start := time.Now()
	for {
		if client.Connected {
			fmt.Println("\nConectado com sucesso!")
			break
		}
		if time.Since(start) > maxWait {
			fmt.Println("\nTimeout ao conectar. Verifique se o QR foi escaneado corretamente.")
			os.Exit(1)
		}
		time.Sleep(1 * time.Second)
	}

	// Sessão estabelecida, finalize
	fmt.Println("Sessão de WhatsApp ativa. Execute o comando de sincronização quando desejar.")
}
