package main

import (
	"fmt"
	"log"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"

	"whatsapp-panel/internal/config"
	"whatsapp-panel/internal/handlers"
	"whatsapp-panel/internal/services/whatsapp"
	"whatsapp-panel/internal/storage"
)

func main() {
	// Carregar configurações
	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatalf("Erro ao carregar configurações: %v", err)
	}

	// Inicializar banco de dados
	db, err := storage.NewDatabase(cfg.DatabasePath)
	if err != nil {
		log.Fatalf("Erro ao inicializar banco de dados: %v", err)
	}
	defer db.Close()

	// Inicializar gerenciador de clientes WhatsApp
	waManager := whatsapp.NewManager(db)

	// Inicializar handlers
	sessionHandler := handlers.NewSessionHandler(waManager, db)
	whatsappHandler := handlers.NewWhatsAppHandler(waManager, db)

	// Configurar servidor Gin
	router := gin.Default()

	// Configurar modo de depuração do Gin
	gin.SetMode(gin.DebugMode)

	// Configurar CORS
	router.Use(cors.Default())

	// Configurar carregamento de templates - CORREÇÃO DEFINITIVA
	// Carregar templates da pasta unified para evitar conflitos
	router.LoadHTMLGlob("web/templates/unified/*.html")

	// Servir arquivos estáticos
	router.Static("/assets", "./web/assets")

	// Definir rotas
	router.GET("/", whatsappHandler.Index)
	router.GET("/sessions", sessionHandler.GetSessions)
	// Rota para gerar modal de QR code para conexão
	router.GET("/qrcode", sessionHandler.GenerateQRCode)
	router.GET("/qrcode/raw", sessionHandler.GenerateQRCodeRaw)
	router.GET("/connection-status", sessionHandler.CheckConnection)
	router.DELETE("/sessions/:id", sessionHandler.DeleteSession)

	// Iniciar servidor
	addr := fmt.Sprintf(":%s", cfg.Port)
	log.Printf("Servidor iniciado em http://localhost%s", addr)
	if err := router.Run(addr); err != nil {
		log.Fatalf("Erro ao iniciar servidor: %v", err)
	}
}
