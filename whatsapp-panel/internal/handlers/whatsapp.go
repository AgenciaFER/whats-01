package handlers

import (
	"log"
	"net/http"

	"github.com/gin-gonic/gin"

	"whatsapp-panel/internal/services/whatsapp"
	"whatsapp-panel/internal/storage"
)

type WhatsAppHandler struct {
	WAClientManager *whatsapp.Manager
	DB              *storage.Database
}

func NewWhatsAppHandler(manager *whatsapp.Manager, db *storage.Database) *WhatsAppHandler {
	return &WhatsAppHandler{
		WAClientManager: manager,
		DB:              db,
	}
}

func (h *WhatsAppHandler) Index(c *gin.Context) {
	sessions, err := h.DB.GetAllSessions()
	if err != nil {
		log.Printf("Erro ao carregar sessões: %v", err)
		c.HTML(http.StatusOK, "error.html", gin.H{
			"Error": "Erro ao carregar sessões. Por favor, tente novamente.",
		})
		return
	}

	c.HTML(http.StatusOK, "index.html", gin.H{
		"Sessions": sessions,
		"Title":    "Painel Principal",
	})
}
