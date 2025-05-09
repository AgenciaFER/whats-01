package whatsapp

import (
	"fmt"

	"whatsapp-panel/internal/storage"

	"go.mau.fi/whatsmeow/types/events"
)

type EventHandler struct {
	DB        *storage.Database
	SessionID string
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
		fmt.Println("Conectado ao WhatsApp")
	case *events.Disconnected:
		fmt.Printf("Desconectado\n")
	case *events.Message:
		fmt.Printf("Mensagem recebida de %s\n", v.Info.Sender.String())
	default:
		fmt.Printf("Evento não tratado: %T\n", v)
	}
}
