package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	_ "github.com/mattn/go-sqlite3"
	"github.com/mdp/qrterminal/v3"
	"go.mau.fi/whatsmeow"
	"go.mau.fi/whatsmeow/store/sqlstore"
	waLog "go.mau.fi/whatsmeow/util/log"
)

func main() {
	// Configurar logger com nível DEBUG para ver todos os detalhes
	logger := waLog.Stdout("whatsmeow", "DEBUG", true)
	logger.Infof("=== TESTE DE QR CODE NO TERMINAL ===")

	// Remover DB anterior se existir
	os.Remove("./terminal_qr_test.db")

	// Criar armazenamento SQLite para a sessão
	container, err := sqlstore.New("sqlite3", "file:./terminal_qr_test.db?_foreign_keys=on&_busy_timeout=30000", logger)
	if err != nil {
		logger.Errorf("ERRO ao criar store: %v", err)
		return
	}

	// Criar novo dispositivo
	deviceStore := container.NewDevice()

	// Criar cliente WhatsApp
	client := whatsmeow.NewClient(deviceStore, logger)

	// Configurar canal para capturar sinais de interrupção (Ctrl+C)
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-c
		fmt.Println("\nDesconectando...")
		client.Disconnect()
		os.Exit(0)
	}()

	// Mostrar informações do dispositivo
	logger.Infof("ID do dispositivo: %v", deviceStore.ID)

	// Obter canal QR antes de conectar
	qrChan, err := client.GetQRChannel(context.Background())
	if err != nil {
		logger.Errorf("ERRO ao obter canal QR: %v", err)
		return
	}

	// Conectar ao WhatsApp
	err = client.Connect()
	if err != nil {
		logger.Errorf("ERRO ao conectar: %v", err)
		return
	}

	logger.Infof("Aguardando código QR...")

	for evt := range qrChan {
		if evt.Event == "code" {
			// Configurar o gerador de QR para terminal
			config := qrterminal.Config{
				Level:     qrterminal.L,
				Writer:    os.Stdout,
				BlackChar: qrterminal.BLACK,
				WhiteChar: qrterminal.WHITE,
				QuietZone: 1,
			}

			// Mostrar QR code no terminal
			fmt.Println("\n==== ESCANEIE ESTE QR CODE COM SEU WHATSAPP ====\n")
			qrterminal.GenerateWithConfig(evt.Code, config)
			fmt.Println("\n=================================================")
			fmt.Println("Aguardando que o código seja escaneado...")
		} else if evt.Event == "success" {
			logger.Infof("QR CODE ESCANEADO COM SUCESSO!")
		}
	}

	// Se chegou aqui, o QR foi escaneado ou houve timeout
	if client.IsConnected() {
		logger.Infof("CONECTADO AO WHATSAPP!")

		// Mostrar informações do telefone
		pinfo := client.Store.PushName
		if pinfo == "" {
			pinfo = "Desconhecido"
		}
		logger.Infof("Conectado como: %s", pinfo)

		// Manter conectado até Ctrl+C
		logger.Infof("Pressione Ctrl+C para sair")
		for {
			if !client.IsConnected() {
				logger.Warnf("Conexão perdida!")
				break
			}
			time.Sleep(1 * time.Second)
		}
	} else {
		logger.Warnf("Não foi possível conectar ao WhatsApp")
	}
}
