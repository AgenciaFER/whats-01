package whatsapp

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"sync"

	"github.com/google/uuid"
	"go.mau.fi/whatsmeow"
	waProto "go.mau.fi/whatsmeow/binary/proto"
	"go.mau.fi/whatsmeow/store/sqlstore"
	"go.mau.fi/whatsmeow/types"
	"go.mau.fi/whatsmeow/types/events"
	waLog "go.mau.fi/whatsmeow/util/log"
	"google.golang.org/protobuf/proto"

	"whatsapp-panel/internal/storage"
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
	storeDir := "storage/sessions"
	if err := os.MkdirAll(storeDir, 0755); err != nil {
		return nil, fmt.Errorf("erro ao criar diretório de sessões: %v", err)
	}

	dbPath := filepath.Join(storeDir, clientID+".db")
	logger := waLog.Stdout("whatsmeow", "DEBUG", true)

	// Criar store com timeout mais longo
	store, err := sqlstore.New("sqlite3", fmt.Sprintf("file:%s?_foreign_keys=on&_busy_timeout=5000", dbPath), logger)
	if err != nil {
		return nil, fmt.Errorf("erro ao criar store: %v", err)
	}

	device := store.NewDevice()
	client := whatsmeow.NewClient(device, logger)

	waCli := &Client{
		WAClient:  client,
		ID:        clientID,
		Store:     store,
		DB:        m.DB,
		Connected: false,
	}

	// Configurar handlers de eventos
	client.AddEventHandler(func(evt interface{}) {
		switch evt.(type) {
		case *events.Connected:
			logger.Infof("Cliente %s conectado com sucesso", clientID)
			waCli.setConnected(true)
		case *events.Disconnected:
			logger.Warnf("Cliente %s desconectado", clientID)
			waCli.setConnected(false)
		case *events.LoggedOut:
			logger.Warnf("Cliente %s deslogado", clientID)
			waCli.setConnected(false)
			go m.RemoveClient(clientID)
		}
	})

	m.Mutex.Lock()
	m.Clients[clientID] = waCli
	m.Mutex.Unlock()

	return waCli, nil
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

	return nil
}

// Implementar o método Disconnect para o cliente WhatsApp
func (c *Client) Disconnect() {
	if c.Connected {
		c.WAClient.Disconnect()
		c.Connected = false
	}
}

func (c *Client) setConnected(status bool) {
	c.Mutex.Lock()
	defer c.Mutex.Unlock()
	c.Connected = status
}

// Implementar o método GetQRChannel para o cliente WhatsApp
func (c *Client) GetQRChannel(ctx context.Context) (<-chan string, error) {
	if c.WAClient == nil {
		return nil, fmt.Errorf("cliente WhatsApp não inicializado")
	}

	qrChan := make(chan string)
	qrChanRaw, err := c.WAClient.GetQRChannel(ctx)
	if err != nil {
		return nil, fmt.Errorf("erro ao obter canal QR: %v", err)
	}

	// Processar QR codes em uma goroutine
	go func() {
		defer close(qrChan)
		for evt := range qrChanRaw {
			if evt.Event == "code" {
				select {
				case qrChan <- evt.Code:
				case <-ctx.Done():
					return
				}
			}
		}
	}()

	return qrChan, nil
}

// SendTextMessage envia uma mensagem de texto para um número de telefone
func (c *Client) SendTextMessage(phoneNumber, message string) error {
	if !c.Connected {
		return fmt.Errorf("cliente não está conectado")
	}

	// Converter número de telefone para formato JID (ID do WhatsApp)
	recipient, err := types.ParseJID(phoneNumber + "@s.whatsapp.net")
	if err != nil {
		return fmt.Errorf("número de telefone inválido: %v", err)
	}

	// Enviar mensagem
	_, err = c.WAClient.SendMessage(context.Background(), recipient, &waProto.Message{
		Conversation: proto.String(message),
	})

	if err != nil {
		return fmt.Errorf("erro ao enviar mensagem: %v", err)
	}

	return nil
}
