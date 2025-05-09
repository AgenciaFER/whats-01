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

// GetStats retorna estatísticas gerais do sistema
func (h *WhatsAppHandler) GetStats(c *gin.Context) {
	sessions, err := h.DB.GetAllSessions()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Erro ao carregar estatísticas"})
		return
	}

	var stats struct {
		TotalSessions  int `json:"total_sessions"`
		ActiveSessions int `json:"active_sessions"`
		TotalContacts  int `json:"total_contacts"`
		TotalGroups    int `json:"total_groups"`
		TotalMessages  int `json:"total_messages"`
	}

	for _, session := range sessions {
		stats.TotalSessions++
		if status, ok := session["status"].(string); ok && status == "connected" {
			stats.ActiveSessions++
		}
		if sessionStats, ok := session["Stats"].(map[string]interface{}); ok {
			if contacts, ok := sessionStats["Contacts"].(int); ok {
				stats.TotalContacts += contacts
			}
			if groups, ok := sessionStats["Groups"].(int); ok {
				stats.TotalGroups += groups
			}
			if messages, ok := sessionStats["MessageCount"].(int); ok {
				stats.TotalMessages += messages
			}
		}
	}

	c.JSON(http.StatusOK, stats)
}

// GetSessionInfo retorna informações detalhadas de uma sessão específica
func (h *WhatsAppHandler) GetSessionInfo(c *gin.Context) {
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

// DisconnectSession desconecta uma sessão do WhatsApp
func (h *WhatsAppHandler) DisconnectSession(c *gin.Context) {
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

	client.Disconnect()
	c.Status(http.StatusNoContent)
}

// SendMessage envia uma mensagem de texto para um número
func (h *WhatsAppHandler) SendMessage(c *gin.Context) {
	sessionID := c.Param("id")
	if sessionID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ID da sessão não fornecido"})
		return
	}

	var req struct {
		PhoneNumber string `json:"phone_number" binding:"required"`
		Message     string `json:"message" binding:"required"`
	}

	if err := c.BindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Dados inválidos",
			"details": err.Error(),
		})
		return
	}

	h.WAClientManager.Mutex.Lock()
	client, exists := h.WAClientManager.Clients[sessionID]
	h.WAClientManager.Mutex.Unlock()

	if !exists {
		c.JSON(http.StatusNotFound, gin.H{"error": "Sessão não encontrada"})
		return
	}

	// Enviar a mensagem
	err := client.SendTextMessage(req.PhoneNumber, req.Message)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Erro ao enviar mensagem",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Mensagem enviada com sucesso",
	})
}

// GetMessageForm retorna o formulário para envio de mensagens
func (h *WhatsAppHandler) GetMessageForm(c *gin.Context) {
	sessionID := c.Param("id")
	if sessionID == "" {
		c.HTML(http.StatusBadRequest, "error.html", gin.H{
			"Error": "ID da sessão não fornecido",
		})
		return
	}

	h.WAClientManager.Mutex.Lock()
	_, exists := h.WAClientManager.Clients[sessionID]
	h.WAClientManager.Mutex.Unlock()

	if !exists {
		c.HTML(http.StatusNotFound, "error.html", gin.H{
			"Error": "Sessão não encontrada",
		})
		return
	}

	c.HTML(http.StatusOK, "message_form.html", gin.H{
		"SessionID": sessionID,
	})
}
