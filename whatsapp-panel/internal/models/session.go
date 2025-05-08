package models

import (
	"time"
)

// Session representa uma sessão do WhatsApp
type Session struct {
	ID          string    `json:"id"`
	Name        string    `json:"name"`
	JID         string    `json:"jid"`
	PhoneNumber string    `json:"phone_number"`
	ConnectedAt time.Time `json:"connected_at"`
	LastActive  time.Time `json:"last_active"`
	Status      string    `json:"status"` // connected, disconnected, etc.
	Stats       Stats     `json:"stats"`
	CreatedAt   time.Time `json:"created_at"`
}

// Stats contém estatísticas de uma sessão do WhatsApp
type Stats struct {
	Contacts      int   `json:"contacts"`
	Groups        int   `json:"groups"`
	Conversations int   `json:"conversations"`
	MessageCount  int64 `json:"message_count"`
}

// SessionList é uma lista de sessões para facilitar o uso em templates
type SessionList []Session

// Status possíveis para uma sessão
const (
	StatusConnected    = "connected"
	StatusDisconnected = "disconnected"
	StatusLoggedOut    = "logged_out"
	StatusPaired       = "paired"
	StatusPending      = "pending"
)

// NewSession cria uma nova sessão
func NewSession(id, name, jid, phoneNumber string) *Session {
	now := time.Now()
	return &Session{
		ID:          id,
		Name:        name,
		JID:         jid,
		PhoneNumber: phoneNumber,
		ConnectedAt: now,
		LastActive:  now,
		Status:      StatusPending,
		Stats: Stats{
			Contacts:      0,
			Groups:        0,
			Conversations: 0,
			MessageCount:  0,
		},
		CreatedAt: now,
	}
}

// UpdateStats atualiza as estatísticas da sessão
func (s *Session) UpdateStats(contacts, groups, conversations int) {
	s.Stats.Contacts = contacts
	s.Stats.Groups = groups
	s.Stats.Conversations = conversations
	s.LastActive = time.Now()
}

// IncrementMessageCount incrementa o contador de mensagens
func (s *Session) IncrementMessageCount() {
	s.Stats.MessageCount++
	s.LastActive = time.Now()
}

// UpdateStatus atualiza o status da sessão
func (s *Session) UpdateStatus(status string) {
	s.Status = status
	s.LastActive = time.Now()
}

// IsActive retorna true se a sessão estiver ativa
func (s *Session) IsActive() bool {
	return s.Status == StatusConnected || s.Status == StatusPaired
}

// FormatPhoneNumber retorna o número de telefone formatado
func (s *Session) FormatPhoneNumber() string {
	if s.PhoneNumber == "" {
		return "Não disponível"
	}
	return s.PhoneNumber
}

// GetDisplayName retorna o nome de exibição da sessão
func (s *Session) GetDisplayName() string {
	if s.Name != "" {
		return s.Name
	}
	return "WhatsApp (" + s.FormatPhoneNumber() + ")"
}
