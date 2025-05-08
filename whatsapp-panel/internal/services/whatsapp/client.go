package whatsapp

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"sync"

	"github.com/google/uuid"
	"go.mau.fi/whatsmeow"
	"go.mau.fi/whatsmeow/store/sqlstore"
	waLog "go.mau.fi/whatsmeow/util/log"

	"github.com/afv/whatsapp-panel/internal/storage"
)

type Client struct {
	WAClient  *whatsmeow.Client
	ID        string
	Store     *sqlstore.Container
	DB        *storage.Database
	Mutex     sync.Mutex
	Connected bool
}

type Manager struct {
	Clients map[string]*Client
	DB      *storage.Database
	Mutex   sync.Mutex
}

func NewManager(db *storage.Database) *Manager {
	return &Manager{
		Clients: make(map[string]*Client),
		DB:      db,
	}
}

func (m *Manager) NewClient() (*Client, error) {
	clientID := uuid.New().String()
	storeDir := filepath.Join(os.TempDir(), "whatsapp-store")
	os.MkdirAll(storeDir, 0755)

	store, err := sqlstore.New("sqlite3", fmt.Sprintf("file:%s?_foreign_keys=on", filepath.Join(storeDir, clientID+".db")), waLog.Stdout("whatsmeow", "INFO", true))
	if err != nil {
		return nil, fmt.Errorf("erro ao criar store: %v", err)
	}

	device := store.NewDevice()

	waClient := whatsmeow.NewClient(device, waLog.Stdout("whatsmeow", "INFO", true))

	client := &Client{
		WAClient:  waClient,
		ID:        clientID,
		Store:     store,
		DB:        m.DB,
		Connected: false,
	}

	m.Mutex.Lock()
	m.Clients[clientID] = client
	m.Mutex.Unlock()

	return client, nil
}

// Adicionar o método RemoveClient para gerenciar a remoção de clientes
func (m *Manager) RemoveClient(clientID string) {
	m.Mutex.Lock()
	defer m.Mutex.Unlock()

	if client, exists := m.Clients[clientID]; exists {
		client.Disconnect()
		delete(m.Clients, clientID)
	}
}

func (c *Client) Connect() error {
	if c.Connected {
		return nil
	}

	err := c.WAClient.Connect()
	if err != nil {
		return fmt.Errorf("erro ao conectar: %v", err)
	}

	c.Connected = true
	return nil
}

// Implementar o método Disconnect para o cliente WhatsApp
func (c *Client) Disconnect() {
	if c.Connected {
		c.WAClient.Disconnect()
		c.Connected = false
	}
}

// Implementar o método GetQRChannel para o cliente WhatsApp
func (c *Client) GetQRChannel(ctx context.Context) (<-chan string, error) {
	qrChan, err := c.WAClient.GetQRChannel(ctx)
	if err != nil {
		return nil, fmt.Errorf("erro ao obter canal de QR Code: %v", err)
	}

	convertedChan := make(chan string)
	go func() {
		for item := range qrChan {
			convertedChan <- item.Code
		}
		close(convertedChan)
	}()
	return convertedChan, nil
}
