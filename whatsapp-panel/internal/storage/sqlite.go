package storage

import (
	"database/sql"
	"fmt"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

type Database struct {
	db *sql.DB
}

// NewDatabase cria uma nova instância do banco de dados
func NewDatabase(dbPath string) (*Database, error) {
	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		return nil, fmt.Errorf("erro ao abrir banco de dados: %v", err)
	}

	// Verificar conexão
	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("erro ao conectar ao banco de dados: %v", err)
	}

	// Criar tabelas necessárias
	if err := createTables(db); err != nil {
		return nil, fmt.Errorf("erro ao criar tabelas: %v", err)
	}

	return &Database{db: db}, nil
}

// createTables cria todas as tabelas necessárias no banco de dados
func createTables(db *sql.DB) error {
	_, err := db.Exec(`
		CREATE TABLE IF NOT EXISTS whatsapp_sessions (
			id TEXT PRIMARY KEY,
			name TEXT NOT NULL,
			jid TEXT NOT NULL,
			phone_number TEXT,
			connected_at TIMESTAMP NOT NULL,
			last_active TIMESTAMP NOT NULL,
			status TEXT NOT NULL
		)
	`)
	if err != nil {
		return err
	}

	_, err = db.Exec(`
		CREATE TABLE IF NOT EXISTS session_stats (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			session_id TEXT NOT NULL,
			contacts INTEGER NOT NULL DEFAULT 0,
			groups INTEGER NOT NULL DEFAULT 0,
			conversations INTEGER NOT NULL DEFAULT 0,
			updated_at TIMESTAMP NOT NULL,
			FOREIGN KEY (session_id) REFERENCES whatsapp_sessions (id) ON DELETE CASCADE
		)
	`)
	if err != nil {
		return err
	}

	return nil
}

// Close fecha a conexão com o banco de dados
func (d *Database) Close() error {
	return d.db.Close()
}

// SaveSession armazena informações da sessão no banco de dados
func (d *Database) SaveSession(id, name, jid, phoneNumber string) error {
	now := time.Now()
	_, err := d.db.Exec(
		`INSERT INTO whatsapp_sessions (id, name, jid, phone_number, connected_at, last_active, status) 
		 VALUES (?, ?, ?, ?, ?, ?, ?)
		 ON CONFLICT(id) DO UPDATE SET 
			name = excluded.name,
			jid = excluded.jid,
			phone_number = excluded.phone_number,
			last_active = excluded.last_active,
			status = excluded.status`,
		id, name, jid, phoneNumber, now, now, "connected",
	)
	return err
}

// UpdateSessionStats atualiza as estatísticas de uma sessão
func (d *Database) UpdateSessionStats(sessionID string, contacts, groups, conversations int) error {
	now := time.Now()
	_, err := d.db.Exec(
		`INSERT INTO session_stats (session_id, contacts, groups, conversations, updated_at)
		 VALUES (?, ?, ?, ?, ?)
		 ON CONFLICT(session_id) DO UPDATE SET
		   contacts = excluded.contacts,
		   groups = excluded.groups,
		   conversations = excluded.conversations,
		   updated_at = excluded.updated_at`,
		sessionID, contacts, groups, conversations, now,
	)
	return err
}

// GetAllSessions retorna todas as sessões ativas
func (d *Database) GetAllSessions() ([]map[string]interface{}, error) {
	rows, err := d.db.Query(`
		SELECT s.id, s.name, s.jid, s.phone_number, s.connected_at, s.last_active, s.status,
		       COALESCE(st.contacts, 0) as contacts,
		       COALESCE(st.groups, 0) as groups,
		       COALESCE(st.conversations, 0) as conversations
		FROM whatsapp_sessions s
		LEFT JOIN session_stats st ON s.id = st.session_id
		WHERE s.status = 'connected'
		ORDER BY s.connected_at DESC
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	sessions := []map[string]interface{}{}
	for rows.Next() {
		var (
			id, name, jid, phoneNumber, status string
			connectedAt, lastActive            time.Time
			contacts, groups, conversations    int
		)
		if err := rows.Scan(&id, &name, &jid, &phoneNumber, &connectedAt, &lastActive, &status, &contacts, &groups, &conversations); err != nil {
			return nil, err
		}
		session := map[string]interface{}{
			"ID":          id,
			"Name":        name,
			"JID":         jid,
			"PhoneNumber": phoneNumber,
			"ConnectedAt": connectedAt.Format("02/01/2006 15:04:05"),
			"LastActive":  lastActive.Format("02/01/2006 15:04:05"),
			"Status":      status,
			"Stats": map[string]int{
				"Contacts":      contacts,
				"Groups":        groups,
				"Conversations": conversations,
			},
		}
		sessions = append(sessions, session)
	}
	return sessions, nil
}

// DeleteSession remove uma sessão do banco de dados
func (d *Database) DeleteSession(id string) error {
	_, err := d.db.Exec("DELETE FROM whatsapp_sessions WHERE id = ?", id)
	return err
}
