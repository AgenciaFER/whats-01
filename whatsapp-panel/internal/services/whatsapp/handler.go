package whatsapp

import (
	"fmt"
	"sync/atomic"

	"whatsapp-panel/internal/storage"

	"go.mau.fi/whatsmeow/types/events"
)

type EventHandler struct {
	DB        *storage.Database
	SessionID string
	Stats     struct {
		MessageCount int64
		Contacts     int
		Groups       int
	}
}

func NewEventHandler(db *storage.Database, sessionID string) *EventHandler {
	return &EventHandler{
		DB:        db,
		SessionID: sessionID,
	}
}

func (h *EventHandler) Handle(evt interface{}) {
	switch v := evt.(type) {
	case *events.Connected:
		fmt.Printf("[%s] ✅ Conectado ao WhatsApp\n", h.SessionID)
		// Quando conectado, apenas atualiza o status
		h.DB.SaveSession(h.SessionID, "WhatsApp", "", "")

	case *events.PairSuccess:
		fmt.Printf("[%s] 🔗 Pareamento bem sucedido!\n", h.SessionID)
		phoneNumber := v.ID.User
		h.DB.SaveSession(h.SessionID, "WhatsApp", v.ID.String(), phoneNumber)

	case *events.ClientOutdated:
		fmt.Println("Cliente desatualizado")

	case *events.Disconnected:
		fmt.Printf("Desconectado do WhatsApp\n")

	case *events.Message:
		// Incrementa o contador de mensagens
		atomic.AddInt64(&h.Stats.MessageCount, 1)
		fmt.Printf("Mensagem recebida de %s\n", v.Info.Sender.String())
		// Atualiza estatísticas após cada mensagem
		h.updateStats()

	case *events.LoggedOut:
		fmt.Println("Deslogado do WhatsApp")
		// Remove a sessão quando deslogado
		h.DB.DeleteSession(h.SessionID)

	default:
		fmt.Printf("Evento não tratado: %T\n", v)
	}
}

func (h *EventHandler) updateStats() {
	h.DB.UpdateSessionStats(
		h.SessionID,
		h.Stats.Contacts,
		h.Stats.Groups,
		int(atomic.LoadInt64(&h.Stats.MessageCount)),
	)
}
