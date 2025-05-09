package main

import (
	"fmt"
	"html/template"
	"log"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-contrib/sessions"
	"github.com/gin-contrib/sessions/cookie"
	"github.com/gin-gonic/gin"

	"whatsapp-panel/internal/config"
	"whatsapp-panel/internal/handlers"
	"whatsapp-panel/internal/services/whatsapp"
	"whatsapp-panel/internal/storage"
)

// Função auxiliar para criar mapas no template
func dict(values ...interface{}) map[string]interface{} {
	m := make(map[string]interface{})
	for i := 0; i < len(values); i += 2 {
		m[values[i].(string)] = values[i+1]
	}
	return m
}

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
	authHandler := handlers.NewAuthHandler()
	sessionHandler := handlers.NewSessionHandler(waManager, db)
	whatsappHandler := handlers.NewWhatsAppHandler(waManager, db)

	// Configurar servidor Gin
	if cfg.Debug {
		gin.SetMode(gin.DebugMode)
	} else {
		gin.SetMode(gin.ReleaseMode)
	}
	router := gin.Default()

	// Configurar funções auxiliares para templates
	router.SetFuncMap(template.FuncMap{
		"dict": dict,
	})

	// Configurar CORS
	router.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"*"},
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Accept", "Authorization"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
	}))

	// Configurar sessões
	store := cookie.NewStore([]byte("secret"))
	router.Use(sessions.Sessions("whatsapp-session", store))

	// Configurar carregamento de templates - CORREÇÃO DEFINITIVA
	router.LoadHTMLGlob("web/templates/unified/*.html")

	// Servir arquivos estáticos
	router.Static("/assets", "./web/assets")

	// Grupo de rotas principais
	mainRoutes := router.Group("/")
	mainRoutes.Use(authHandler.AuthMiddleware())
	{
		mainRoutes.GET("/", whatsappHandler.Index)
		mainRoutes.GET("/stats", whatsappHandler.GetStats)
	}

	// Grupo de rotas para sessões
	sessionRoutes := router.Group("/sessions")
	sessionRoutes.Use(authHandler.AuthMiddleware())
	{
		sessionRoutes.GET("/", sessionHandler.GetSessionsHTML)
		sessionRoutes.GET("/list", sessionHandler.GetSessions)
		sessionRoutes.GET("/:id", sessionHandler.GetSessionInfo)
		sessionRoutes.DELETE("/:id", sessionHandler.DeleteSession)
		sessionRoutes.POST("/:id/disconnect", whatsappHandler.DisconnectSession)
		// Adicionar rotas para envio de mensagens
		sessionRoutes.GET("/:id/message", whatsappHandler.GetMessageForm)
		sessionRoutes.POST("/:id/message", whatsappHandler.SendMessage)
	}

	// Grupo de rotas para QR Code
	// Grupo de rotas para QR Code
	qrRoutes := router.Group("/qrcode")
	qrRoutes.Use(authHandler.AuthMiddleware())
	{
		qrRoutes.GET("/", sessionHandler.GenerateQRCode)
		qrRoutes.GET("/raw", func(c *gin.Context) {
			// Adicionar cabeçalhos para prevenir caching
			c.Header("Cache-Control", "no-store, no-cache, must-revalidate")
			c.Header("Pragma", "no-cache")
			c.Header("Expires", "0")
			sessionHandler.GenerateQRCodeRaw(c)
		})
	}

	// Rotas de status
	router.GET("/connection-status", func(c *gin.Context) {
		// Adicionar cabeçalhos para prevenir caching
		c.Header("Cache-Control", "no-store, no-cache, must-revalidate")
		c.Header("Pragma", "no-cache")
		c.Header("Expires", "0")
		sessionHandler.CheckConnection(c)
	})

	router.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok"})
	})

	// Iniciar servidor
	addr := fmt.Sprintf(":%s", cfg.Port)
	log.Printf("Servidor iniciado em http://localhost%s", addr)
	if err := router.Run(addr); err != nil {
		log.Fatalf("Erro ao iniciar servidor: %v", err)
	}
}
