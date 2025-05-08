package handlers

import (
	"log"
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/afv/whatsapp-panel/internal/services/whatsapp"
	"github.com/afv/whatsapp-panel/internal/storage"
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
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Erro ao carregar sessões"})
		return
	}

	// Adicionar log para verificar as sessões retornadas
	log.Printf("Sessões retornadas: %+v", sessions)

	c.HTML(http.StatusOK, "index.html", gin.H{
		"Sessions": sessions,
	})
}
