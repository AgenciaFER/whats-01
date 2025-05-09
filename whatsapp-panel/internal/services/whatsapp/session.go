package whatsapp

import (
	"sync"
	"time"

	"whatsapp-panel/internal/models"
)

// SessionManager gerencia múltiplas sessões do WhatsApp
type SessionManager struct {
	sessions map[string]*Session
	mu       sync.RWMutex
}

// Session representa uma sessão individual do WhatsApp
type Session struct {
	ID           string
	Client       *Client
	LastActivity time.Time
	Stats        models.Stats
	mu           sync.RWMutex
}

// NewSessionManager cria uma nova instância do gerenciador de sessões
func NewSessionManager() *SessionManager {
	return &SessionManager{
		sessions: make(map[string]*Session),
	}
}

// CreateSession cria uma nova sessão
func (sm *SessionManager) CreateSession(client *Client) *Session {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	session := &Session{
		ID:           client.ID,
		Client:       client,
		LastActivity: time.Now(),
	}

	sm.sessions[client.ID] = session
	return session
}

// GetSession retorna uma sessão pelo ID
func (sm *SessionManager) GetSession(id string) (*Session, bool) {
	sm.mu.RLock()
	defer sm.mu.RUnlock()

	session, exists := sm.sessions[id]
	return session, exists
}

// RemoveSession remove uma sessão
func (sm *SessionManager) RemoveSession(id string) {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	if session, exists := sm.sessions[id]; exists {
		session.Client.Disconnect()
		delete(sm.sessions, id)
	}
}

// UpdateActivity atualiza o timestamp de última atividade
func (s *Session) UpdateActivity() {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.LastActivity = time.Now()
}

// UpdateStats atualiza as estatísticas da sessão
func (s *Session) UpdateStats(stats models.Stats) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.Stats = stats
}

// GetStats retorna as estatísticas atuais da sessão
func (s *Session) GetStats() models.Stats {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.Stats
}
