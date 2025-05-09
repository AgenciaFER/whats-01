//go:build ignore
// +build ignore

package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	_ "github.com/mattn/go-sqlite3"
	"go.mau.fi/whatsmeow"
	"go.mau.fi/whatsmeow/store/sqlstore"
	waLog "go.mau.fi/whatsmeow/util/log"
)

func main() {
	// Configurar logger
	logger := waLog.Stdout("whatsmeow", "DEBUG", true)
	logger.Infof("Iniciando teste de conexão com WhatsApp")

	// Limpar DB anterior de teste (se existir)
	os.Remove("./test_connection.db")

	// Criar store
	container, err := sqlstore.New("sqlite3", "file:./test_connection.db?_foreign_keys=on", logger)
	if err != nil {
		logger.Errorf("Erro ao criar store: %v", err)
		return
	}

	// Criar dispositivo
	deviceStore := container.NewDevice()
	client := whatsmeow.NewClient(deviceStore, logger)

	// Setup interrupção para fechamento limpo
	c := make(chan os.Signal)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-c
		client.Disconnect()
		os.Exit(0)
	}()

	// Verificar se precisa de QR code
	if client.Store.ID == nil {
		// Precisa parear
		logger.Infof("Dispositivo novo, aguardando QR code...")
		qrChan, _ := client.GetQRChannel(context.Background())
		err = client.Connect()
		if err != nil {
			logger.Errorf("Erro ao conectar: %v", err)
			return
		}

		for evt := range qrChan {
			if evt.Event == "code" {
				// Exibir QR code
				logger.Infof("QR code recebido: %s", evt.Code)
				logger.Infof("Escaneie este código com seu aplicativo WhatsApp")
				logger.Infof("Por favor, aguarde...")
			} else {
				logger.Infof("Evento QR: %v", evt)
			}
		}
	} else {
		// Já está pareado
		logger.Infof("Dispositivo já pareado, tentando reconectar...")
		err = client.Connect()
		if err != nil {
			logger.Errorf("Erro ao conectar: %v", err)
			return
		}
	}

	// Aguardar eventos
	logger.Infof("Conectado! Pressione Ctrl+C para sair")
	for {
		if !client.IsConnected() {
			logger.Warnf("Conexão perdida!")
			break
		}
		time.Sleep(1 * time.Second)
		fmt.Print(".")
	}
}
