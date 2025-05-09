package handlers

import (
	"context"
	"encoding/base64"
	"log"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/skip2/go-qrcode"

	"whatsapp-panel/internal/services/whatsapp"
	"whatsapp-panel/internal/storage"
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

// GetSessionsHTML renderiza a lista de sessões em formato HTML
func (h *SessionHandler) GetSessionsHTML(c *gin.Context) {
	sessions, err := h.DB.GetAllSessions()
	if err != nil {
		c.HTML(http.StatusInternalServerError, "error", gin.H{
			"Error": "Erro ao buscar sessões: " + err.Error(),
		})
		return
	}

	// Atualizar informações de conexão
	for i := range sessions {
		sessionID, ok := sessions[i]["id"].(string)
		if !ok {
			continue // se não conseguir converter o ID, pula para a próxima sessão
		}

		h.WAClientManager.Mutex.Lock()
		client, exists := h.WAClientManager.Clients[sessionID]
		h.WAClientManager.Mutex.Unlock()

		if exists {
			if client.Connected {
				sessions[i]["status"] = "connected"
			} else {
				sessions[i]["status"] = "disconnected"
			}
		} else {
			sessions[i]["status"] = "disconnected"
		}
	}

	c.HTML(http.StatusOK, "sessions", gin.H{
		"Sessions": sessions,
		"Title":    "Sessões WhatsApp",
	})
}

func (h *SessionHandler) GenerateQRCode(c *gin.Context) {
	log.Println("[GenerateQRCode] handler chamado")
	client, err := h.WAClientManager.NewClient()
	if err != nil {
		c.HTML(http.StatusInternalServerError, "error", gin.H{"Error": "Erro ao criar cliente: " + err.Error()})
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()

	log.Printf("[QRCode] Aguardando QR code para sessao %s...", client.ID)
	qrChan, err := client.GetQRChannel(ctx)
	if err != nil {
		h.WAClientManager.RemoveClient(client.ID)
		c.HTML(http.StatusInternalServerError, "error", gin.H{"Error": "Erro ao obter QR Code: " + err.Error()})
		return
	}

	select {
	case qrCode := <-qrChan:
		log.Printf("[QRCode] Recebido QR code para sessao %s", client.ID)
		qrImage, err := qrcode.Encode(qrCode, qrcode.Medium, 256)
		if err != nil {
			h.WAClientManager.RemoveClient(client.ID)
			c.HTML(http.StatusInternalServerError, "error", gin.H{"Error": "Erro ao gerar QR Code: " + err.Error()})
			return
		}
		c.HTML(http.StatusOK, "qrcode.html", gin.H{
			"QRCode":    base64.StdEncoding.EncodeToString(qrImage),
			"SessionID": client.ID,
		})
	case <-ctx.Done():
		log.Printf("[QRCode] Timeout aguardando QR code para sessao %s", client.ID)
		h.WAClientManager.RemoveClient(client.ID)
		c.HTML(http.StatusRequestTimeout, "error", gin.H{"Error": "Timeout ao gerar QR Code"})
	}
}

// GenerateQRCodeRaw retorna JSON com QR code base64 e SessionID
func (h *SessionHandler) GenerateQRCodeRaw(c *gin.Context) {
	client, err := h.WAClientManager.NewClient()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Erro ao criar cliente: " + err.Error()})
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()

	qrChan, err := client.GetQRChannel(ctx)
	if err != nil {
		h.WAClientManager.RemoveClient(client.ID)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Erro ao obter QR Code: " + err.Error()})
		return
	}

	select {
	case qrCode := <-qrChan:
		qrImage, err := qrcode.Encode(qrCode, qrcode.Medium, 256)
		if err != nil {
			h.WAClientManager.RemoveClient(client.ID)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Erro ao gerar QR Code: " + err.Error()})
			return
		}
		c.JSON(http.StatusOK, gin.H{
			"QRCode":    base64.StdEncoding.EncodeToString(qrImage),
			"SessionID": client.ID,
		})
	case <-ctx.Done():
		h.WAClientManager.RemoveClient(client.ID)
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

func (h *SessionHandler) CheckConnection(c *gin.Context) {
	sessionID := c.Query("session_id")
	log.Printf("[CheckConnection] chamada com session_id=%s", sessionID)
	if sessionID == "" {
		log.Println("[CheckConnection] session_id vazio")
		c.JSON(http.StatusBadRequest, gin.H{"error": "ID da sessão não fornecido"})
		return
	}

	h.WAClientManager.Mutex.Lock()
	client, exists := h.WAClientManager.Clients[sessionID]
	h.WAClientManager.Mutex.Unlock()
	log.Printf("[CheckConnection] exists=%v", exists)
	if !exists {
		c.JSON(http.StatusNotFound, gin.H{"error": "Sessão não encontrada"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"connected": client.Connected})
}
