package storage

// Database define a interface para operações no banco de dados
type DatabaseInterface interface {
	Close() error
	SaveSession(id, name, jid, phoneNumber string) error
	UpdateSessionStats(sessionID string, contacts, groups, conversations int) error
	GetAllSessions() ([]map[string]interface{}, error)
	DeleteSession(id string) error
}

// Garantir que Database implementa DatabaseInterface
var _ DatabaseInterface = (*Database)(nil)
