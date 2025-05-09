// Este arquivo contém as correções para o arquivo internal/handlers/session.go
// Foco na função GenerateQRCode que apresenta problemas

package handlers

import (
	"context"
	"encoding/base64"
	"log"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/skip2/go-qrcode"
)

// Esta é a função corrigida para gerar o QR code
func (h *SessionHandler) GenerateQRCode(c *gin.Context) {
	log.Println("[GenerateQRCode] handler chamado")
	client, err := h.WAClientManager.NewClient()
	if err != nil {
		log.Printf("[ERROR] Erro ao criar cliente: %v", err)
		c.HTML(http.StatusInternalServerError, "error", gin.H{
			"Error": "Erro ao criar cliente: " + err.Error(),
		})
		return
	}

	// Contexto com timeout mais longo - 2 minutos
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()

	log.Printf("[QRCode] Aguardando QR code para sessão %s...", client.ID)
	qrChan, err := client.GetQRChannel(ctx)
	if err != nil {
		log.Printf("[ERROR] Erro ao obter QR code: %v", err)
		h.WAClientManager.RemoveClient(client.ID)
		c.HTML(http.StatusInternalServerError, "error", gin.H{
			"Error": "Erro ao obter QR Code: " + err.Error(),
		})
		return
	}

	// Iniciar a conexão em uma goroutine separada
	go func() {
		log.Printf("[Connection] Iniciando conexão para sessão %s", client.ID)
		// Esperar um pouco para dar tempo do QR ser escaneado
		time.Sleep(1 * time.Second)
		if err := client.WAClient.Connect(); err != nil {
			log.Printf("[ERROR] Erro ao conectar cliente %s: %v", client.ID, err)
			return
		}
	}()

	// Esperar pelo QR code
	select {
	case qrCode := <-qrChan:
		log.Printf("[QRCode] Recebido QR code para sessão %s", client.ID)
		qrImage, err := qrcode.Encode(qrCode, qrcode.Medium, 256)
		if err != nil {
			log.Printf("[ERROR] Erro ao gerar QR code: %v", err)
			h.WAClientManager.RemoveClient(client.ID)
			c.HTML(http.StatusInternalServerError, "error", gin.H{
				"Error": "Erro ao gerar QR Code: " + err.Error(),
			})
			return
		}

		// Renderizar o template com o QR code
		c.HTML(http.StatusOK, "qrcode", gin.H{
			"QRCode":    base64.StdEncoding.EncodeToString(qrImage),
			"SessionID": client.ID,
		})

	case <-ctx.Done():
		log.Printf("[ERROR] Timeout aguardando QR code para sessão %s", client.ID)
		h.WAClientManager.RemoveClient(client.ID)
		c.HTML(http.StatusRequestTimeout, "error", gin.H{
			"Error": "Timeout ao gerar QR Code",
		})
	}
}

// Esta é a função corrigida para verificar a conexão
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

	// Verificar se o cliente está conectado
	connected := client.Connected
	log.Printf("[CheckConnection] connected=%v", connected)

	// Se estiver conectado, salvar a sessão no banco de dados
	if connected {
		phoneNumber := ""
		if client.WAClient.Store.ID != nil {
			jid := client.WAClient.Store.ID.String()
			phoneNumber = client.WAClient.Store.ID.User
			log.Printf("[CheckConnection] Salvando sessão no banco de dados: JID=%s, Phone=%s", jid, phoneNumber)
			// Salvar a sessão no banco de dados
			if err := h.DB.SaveSession(sessionID, "WhatsApp", jid, phoneNumber); err != nil {
				log.Printf("[ERROR] Erro ao salvar sessão: %v", err)
			}
		}
	}

	c.JSON(http.StatusOK, gin.H{"connected": connected})
}
