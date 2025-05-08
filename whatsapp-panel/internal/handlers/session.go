package handlers

import (
	"context"
	"encoding/base64"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/skip2/go-qrcode"

	"github.com/afv/whatsapp-panel/internal/services/whatsapp"
	"github.com/afv/whatsapp-panel/internal/storage"
)

type SessionHandler struct {
	WAClientManager *whatsapp.Manager
	DB              *storage.Database
}

func NewSessionHandler(manager *whatsapp.Manager, db *storage.Database) *SessionHandler {
	return &SessionHandler{
		WAClientManager: manager,
		DB:              db,
	}
}

func (h *SessionHandler) GetSessions(c *gin.Context) {
	sessions, err := h.DB.GetAllSessions()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Erro ao buscar sessões"})
		return
	}
	c.JSON(http.StatusOK, sessions)
}

func (h *SessionHandler) GenerateQRCode(c *gin.Context) {
	client, err := h.WAClientManager.NewClient()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Erro ao criar cliente"})
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()

	qrChan, err := client.GetQRChannel(ctx)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Erro ao obter QR Code"})
		return
	}

	select {
	case qrCode := <-qrChan:
		qrImage, err := qrcode.Encode(qrCode, qrcode.Medium, 256)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Erro ao gerar QR Code"})
			return
		}
		c.JSON(http.StatusOK, gin.H{"qrcode": base64.StdEncoding.EncodeToString(qrImage)})
	case <-ctx.Done():
		c.JSON(http.StatusRequestTimeout, gin.H{"error": "Timeout ao gerar QR Code"})
	}
}

func (h *SessionHandler) DeleteSession(c *gin.Context) {
	sessionID := c.Param("id")
	if sessionID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ID da sessão não fornecido"})
		return
	}

	h.WAClientManager.RemoveClient(sessionID)
	if err := h.DB.DeleteSession(sessionID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Erro ao deletar sessão"})
		return
	}

	c.Status(http.StatusNoContent)
}
