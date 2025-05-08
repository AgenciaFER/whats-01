package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"

	"github.com/afv/whatsapp-panel/internal/config"
	"github.com/afv/whatsapp-panel/internal/handlers"
	"github.com/afv/whatsapp-panel/internal/services/whatsapp"
	"github.com/afv/whatsapp-panel/internal/storage"
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

	// Configurar carregamento de templates
	router.LoadHTMLFiles(
		"web/templates/index.html",
		"web/templates/qrcode.html",
		"web/templates/sessions.html",
		"web/templates/session_card.html",
		"web/templates/test.html",
	)

	// Configurar CORS
	router.Use(cors.Default())

	// Configurar arquivos estáticos
	router.Static("/assets", "web/assets")

	// Configurar rotas
	router.GET("/", whatsappHandler.Index)
	router.GET("/sessions", sessionHandler.GetSessions)
	router.GET("/qrcode", sessionHandler.GenerateQRCode)
	router.DELETE("/sessions/:id", sessionHandler.DeleteSession)

	// Adicionar rota de teste para verificar o carregamento de templates
	router.GET("/test", func(c *gin.Context) {
		log.Println("Acessando rota /test")
		c.HTML(http.StatusOK, "test.html", nil)
		log.Println("Template renderizado com sucesso")
	})

	// Iniciar servidor
	addr := fmt.Sprintf(":%s", cfg.Port)
	log.Printf("Servidor iniciado em http://localhost%s", addr)
	if err := router.Run(addr); err != nil {
		log.Fatalf("Erro ao iniciar servidor: %v", err)
	}
}
