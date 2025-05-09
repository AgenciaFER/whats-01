package models

import "time"

// SessionStats contém estatísticas de uma sessão do WhatsApp
type SessionStats struct {
	SessionID     string    `json:"session_id"`
	Contacts      int       `json:"contacts"`
	Groups        int       `json:"groups"`
	Conversations int       `json:"conversations"`
	UpdatedAt     time.Time `json:"updated_at"`
}

// StatsUpdate representa uma atualização de estatísticas
type StatsUpdate struct {
	Contacts      int `json:"contacts"`
	Groups        int `json:"groups"`
	Conversations int `json:"conversations"`
}
