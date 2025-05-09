//go:build exportcontacts
// +build exportcontacts

package main

import (
	"context"
	"encoding/csv"
	"fmt"
	"os"
	"time"

	"whatsapp-panel/internal/services/whatsapp"
	"whatsapp-panel/internal/storage"

	qrterminal "github.com/mdp/qrterminal/v3"
	"go.mau.fi/whatsmeow/store/sqlstore"
	"go.mau.fi/whatsmeow/types"
)

func main() {
	// Inicializar banco de dados (usa armazenamento existente)
	db, err := storage.NewDatabase("contacts.db")
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

	// Contexto para QR Code
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()

	// Obter canal de QR Code
	qrChan, err := client.GetQRChannel(ctx)
	if err != nil {
		fmt.Println("Erro ao obter canal de QR Code:", err)
		os.Exit(1)
	}

	// Mostrar QR Code no terminal (receber apenas uma vez)
	fmt.Println("Escaneie este QR Code no WhatsApp:")
	if code, ok := <-qrChan; ok {
		qrterminal.GenerateHalfBlock(code, qrterminal.L, os.Stdout)
	} else {
		fmt.Println("Erro: não foi possível obter o QR Code")
		os.Exit(1)
	}

	// Aguardar confirmação de pareamento com status percentual
	maxWait := 30 * time.Second
	start := time.Now()
	for {
		elapsed := time.Since(start)
		percent := int(elapsed * 100 / maxWait)
		if percent > 100 {
			percent = 100
		}
		fmt.Printf("\rConectando: %d%%", percent)
		if client.Connected {
			fmt.Println("\rConectado com sucesso!   ")
			break
		}
		if elapsed > maxWait {
			fmt.Println("\nTimeout ao conectar. Verifique se o QR foi escaneado corretamente.")
			os.Exit(1)
		}
		time.Sleep(500 * time.Millisecond)
	}

	// Coletar contatos via SQLStore
	devid := *client.WAClient.Store.ID
	sqlClient := sqlstore.NewSQLStore(client.Store, devid)
	contactsMap, err := sqlClient.GetAllContacts()
	if err != nil {
		fmt.Println("Erro ao obter contatos:", err)
		os.Exit(1)
	}
	// Preparar progresso de exportação
	total := len(contactsMap)
	sliceJIDs := make([]types.JID, 0, total)
	for jid := range contactsMap {
		sliceJIDs = append(sliceJIDs, jid)
	}

	fmt.Printf("Exportando %d contatos...\n", total)

	// Salvar em CSV
	file, err := os.Create("contatos.csv")
	if err != nil {
		fmt.Println("Erro ao criar arquivo CSV:", err)
		os.Exit(1)
	}
	defer file.Close()

	writer := csv.NewWriter(file)
	defer writer.Flush()

	// Cabeçalho
	writer.Write([]string{"JID", "PushName", "Number"})

	// Escrever contatos com progresso
	for i, jid := range sliceJIDs {
		info := contactsMap[jid]
		number := jid.User + "@" + jid.Server
		writer.Write([]string{jid.String(), info.PushName, number})
		percent := (i + 1) * 100 / total
		fmt.Printf("\rExportando: %d/%d (%d%%)", i+1, total, percent)
	}
	fmt.Println("\nExportação concluída. \nContatos salvos em contatos.csv")
}
