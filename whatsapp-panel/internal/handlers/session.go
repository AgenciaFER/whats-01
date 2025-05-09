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
		sessionID, ok := sessions[i]["ID"].(string)
		if !ok {
			continue // se não conseguir converter o ID, pula para a próxima sessão
		}

		h.WAClientManager.Mutex.Lock()
		client, exists := h.WAClientManager.Clients[sessionID]
		h.WAClientManager.Mutex.Unlock()

		if exists {
			if client.Connected {
				sessions[i]["Status"] = "connected"
			} else {
				sessions[i]["Status"] = "disconnected"
			}
		} else {
			sessions[i]["Status"] = "disconnected"
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
		log.Printf("[ERROR] Erro ao criar cliente: %v", err)
		c.HTML(http.StatusInternalServerError, "error.html", gin.H{ // nome de template com extensão
			"Error": "Erro ao criar cliente: " + err.Error(),
		})
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()

	// Iniciar tentativa de conexão em background antes de mostrar o QR
	go func() {
		log.Printf("[Connection] Iniciando conexão para sessão %s", client.ID)
		if err := client.WAClient.Connect(); err != nil {
			log.Printf("[ERROR] Erro ao conectar cliente %s: %v", client.ID, err)
		}
	}()

	log.Printf("[QRCode] Aguardando QR code para sessão %s...", client.ID)
	qrChan, err := client.GetQRChannel(ctx)
	if err != nil {
		log.Printf("[ERROR] Erro ao obter QR code: %v", err)
		h.WAClientManager.RemoveClient(client.ID)
		c.HTML(http.StatusInternalServerError, "error.html", gin.H{ // nome de template com extensão
			"Error": "Erro ao obter QR Code: " + err.Error(),
		})
		return
	}

	select {
	case qrCode := <-qrChan:
		log.Printf("[QRCode] Recebido QR code para sessão %s", client.ID)
		qrImage, err := qrcode.Encode(qrCode, qrcode.Medium, 256)
		if err != nil {
			log.Printf("[ERROR] Erro ao gerar QR code: %v", err)
			h.WAClientManager.RemoveClient(client.ID)
			c.HTML(http.StatusInternalServerError, "error.html", gin.H{ // nome de template com extensão
				"Error": "Erro ao gerar QR Code: " + err.Error(),
			})
			return
		}

		c.HTML(http.StatusOK, "qrcode.html", gin.H{ // nome de template com extensão
			"QRCode":    base64.StdEncoding.EncodeToString(qrImage),
			"SessionID": client.ID,
		})

	case <-ctx.Done():
		log.Printf("[ERROR] Timeout aguardando QR code para sessão %s", client.ID)
		h.WAClientManager.RemoveClient(client.ID)
		c.HTML(http.StatusRequestTimeout, "error.html", gin.H{ // nome de template com extensão
			"Error": "Timeout ao gerar QR Code",
		})
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

	// Iniciar tentativa de conexão em background antes de mostrar o QR
	go func() {
		log.Printf("[Connection] Iniciando conexão para sessão %s", client.ID)
		if err := client.WAClient.Connect(); err != nil {
			log.Printf("[ERROR] Erro ao conectar cliente %s: %v", client.ID, err)
		}
	}()

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

	log.Printf("[CheckConnection] client.Connected=%v", client.Connected)
	c.JSON(http.StatusOK, gin.H{"connected": client.Connected})
}

// GetSessionInfo retorna informações detalhadas de uma sessão específica
func (h *SessionHandler) GetSessionInfo(c *gin.Context) {
	sessionID := c.Param("id")
	if sessionID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ID da sessão não fornecido"})
		return
	}

	h.WAClientManager.Mutex.Lock()
	client, exists := h.WAClientManager.Clients[sessionID]
	h.WAClientManager.Mutex.Unlock()

	if !exists {
		c.JSON(http.StatusNotFound, gin.H{"error": "Sessão não encontrada"})
		return
	}

	// Buscar informações da sessão no banco de dados
	sessions, err := h.DB.GetAllSessions()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Erro ao buscar informações da sessão"})
		return
	}

	var sessionInfo map[string]interface{}
	for _, s := range sessions {
		if id, ok := s["ID"].(string); ok && id == sessionID {
			sessionInfo = s
			break
		}
	}

	if sessionInfo == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Informações da sessão não encontradas"})
		return
	}

	// Adicionar status de conexão atual
	sessionInfo["is_connected"] = client.Connected

	c.JSON(http.StatusOK, sessionInfo)
}
