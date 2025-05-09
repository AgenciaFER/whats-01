// filepath: /Users/afv/Documents/whats-01/whatsapp-panel/internal/services/whatsapp/client.go
package whatsapp

import (
	"context"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"sync"
	"time"

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

// Configuração global para limites de conexão
var (
	// Tempo para aguardar antes de limpar uma sessão não conectada
	CleanupTimeout = 2 * time.Minute
)

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

	// Usar logger mais detalhado
	logger := waLog.Stdout("whatsmeow", "DEBUG", true)

	// Criar store com timeout mais longo e flags adicionais
	store, err := sqlstore.New("sqlite3", fmt.Sprintf("file:%s?_foreign_keys=on&_busy_timeout=30000&cache=shared", dbPath), logger)
	if err != nil {
		return nil, fmt.Errorf("erro ao criar store: %v", err)
	}

	device := store.NewDevice()

	// Configurar cliente WhatsApp
	client := whatsmeow.NewClient(device, logger)

	// Verificar se está configurado corretamente
	if client != nil && client.Store != nil {
		log.Printf("Store configurado: %T", client.Store)
	}

	waCli := &Client{
		WAClient:  client,
		ID:        clientID,
		Store:     store,
		DB:        m.DB,
		Connected: false,
	}

	// Configurar handlers de eventos
	client.AddEventHandler(func(evt interface{}) {
		switch e := evt.(type) {
		case *events.Connected:
			log.Printf("[Client %s] ✅ Cliente conectado com sucesso", clientID)
			waCli.setConnected(true)
		case *events.Disconnected:
			log.Printf("[Client %s] ⚠️ Cliente desconectado", clientID)
			waCli.setConnected(false)
		case *events.LoggedOut:
			log.Printf("[Client %s] ❌ Cliente deslogado", clientID)
			waCli.setConnected(false)
			go m.RemoveClient(clientID)
		case *events.QR:
			log.Printf("[Client %s] 📱 Evento QR recebido", clientID)
		case *events.ConnectFailure:
			log.Printf("[Client %s] ❌ Falha na conexão: %v", clientID, e)
		default:
			log.Printf("[Client %s] Evento recebido: %T", clientID, e)
		}
	})

	m.Mutex.Lock()
	m.Clients[clientID] = waCli
	m.Mutex.Unlock()

	// Configurar limpeza automática com timeout global
	m.CleanupClientAfterTimeout(clientID, CleanupTimeout)
	log.Printf("[Manager] Cliente %s criado com limpeza automática configurada", clientID)

	return waCli, nil
}

// Adicionar o método RemoveClient para gerenciar a remoção de clientes
func (m *Manager) RemoveClient(clientID string) {
	m.Mutex.Lock()
	defer m.Mutex.Unlock()

	if client, exists := m.Clients[clientID]; exists {
		client.Disconnect()
		delete(m.Clients, clientID)
		log.Printf("[Manager] Cliente %s removido do gerenciador", clientID)
	}
}

// CleanupClientAfterTimeout configura um temporizador para remover o cliente
// se ele não se conectar dentro de um determinado período de tempo.
// Também remove o arquivo de banco de dados associado.
func (m *Manager) CleanupClientAfterTimeout(clientID string, timeout time.Duration) {
	go func() {
		// Aguardar pelo timeout
		time.Sleep(timeout)

		// Verificar se o cliente ainda existe e não está conectado
		m.Mutex.Lock()
		client, exists := m.Clients[clientID]
		m.Mutex.Unlock()

		if exists && !client.Connected {
			// Cliente não se conectou dentro do timeout
			log.Printf("[Cleanup] Cliente %s não conectou em %s, removendo", clientID, timeout)

			// Remover cliente do gerenciador
			m.RemoveClient(clientID)

			// Remover arquivo de banco de dados
			storeDir := "storage/sessions"
			dbPath := filepath.Join(storeDir, clientID+".db")

			// Forçar a desconexão do cliente do WhatsApp
			if client != nil && client.WAClient != nil {
				client.WAClient.Disconnect()
			}

			if err := os.Remove(dbPath); err != nil {
				log.Printf("[Cleanup] Erro ao excluir arquivo de sessão %s: %v", dbPath, err)
			} else {
				log.Printf("[Cleanup] Arquivo de sessão removido: %s", dbPath)
			}
		}
	}()
}

// Modificar o método Connect para resolver problemas comuns
func (c *Client) Connect() error {
	c.Mutex.Lock()
	alreadyConnected := c.Connected
	c.Mutex.Unlock()

	if alreadyConnected {
		log.Printf("[Client %s] Cliente já conectado, ignorando solicitação de conexão", c.ID)
		return nil
	}

	log.Printf("[Client %s] Iniciando conexão...", c.ID)

	// Usar contexto para a conexão
	err := c.WAClient.Connect()
	if err != nil {
		log.Printf("[Client %s] Erro ao conectar: %v", c.ID, err)
		return fmt.Errorf("erro ao conectar: %v", err)
	}

	// Verificar se a conexão foi bem-sucedida
	connected := false
	for i := 0; i < 10; i++ {
		if c.WAClient.IsConnected() {
			connected = true
			break
		}
		time.Sleep(1 * time.Second)
	}

	if connected {
		log.Printf("[Client %s] Conexão estabelecida com sucesso", c.ID)
		c.setConnected(true)
	} else {
		log.Printf("[Client %s] Timeout ao estabelecer conexão", c.ID)
		return fmt.Errorf("timeout ao estabelecer conexão")
	}

	return nil
}

// Melhorar o método de desconexão
func (c *Client) Disconnect() {
	c.Mutex.Lock()
	wasConnected := c.Connected
	c.Mutex.Unlock()

	if wasConnected {
		log.Printf("[Client %s] Desconectando cliente que estava ativo", c.ID)
		c.WAClient.Disconnect()
		c.setConnected(false)
	} else {
		log.Printf("[Client %s] Tentativa de desconexão em cliente já desconectado", c.ID)
	}
}

// Melhorar o método setConnected para notificar quando o status mudar
func (c *Client) setConnected(status bool) {
	c.Mutex.Lock()
	oldStatus := c.Connected
	c.Connected = status
	c.Mutex.Unlock()

	// Se houve mudança de status, registrar no log
	if oldStatus != status {
		if status {
			log.Printf("[Client %s] Status mudou para CONECTADO", c.ID)
		} else {
			log.Printf("[Client %s] Status mudou para DESCONECTADO", c.ID)
		}
	}
}

// GetQRChannel retorna um canal que recebe códigos QR para pareamento
func (c *Client) GetQRChannel(ctx context.Context) (<-chan string, error) {
	if c.WAClient == nil {
		return nil, fmt.Errorf("cliente WhatsApp não inicializado")
	}

	log.Printf("[GetQRChannel] Iniciando obtenção de QR code para sessão %s", c.ID)

	// Canal para enviar os códigos QR
	qrChan := make(chan string, 1)

	// Remover conexão automática aqui; Connect será chamado no handler após obter o canal QR

	// Também obter o QR code pelo método padrão
	qrChanRaw, err := c.WAClient.GetQRChannel(ctx)
	if err != nil {
		log.Printf("[GetQRChannel] Erro ao obter canal QR original: %v", err)
		return nil, err
	}

	log.Printf("[GetQRChannel] Canal QR obtido com sucesso, aguardando código")

	// Encaminhar os códigos do canal original para o nosso canal
	go func() {
		defer close(qrChan)
		for evt := range qrChanRaw {
			log.Printf("[GetQRChannel] Evento QR recebido: %+v", evt)
			if evt.Event == "code" {
				log.Printf("[GetQRChannel] QR code recebido via canal original: %s", evt.Code)
				// Usar diretamente o código QR sem adicionar referência
				select {
				case qrChan <- evt.Code:
					log.Printf("[GetQRChannel] QR code enviado com sucesso")
				case <-ctx.Done():
					log.Printf("[GetQRChannel] Contexto encerrado durante envio de QR code")
					return
				}
			}
		}
		log.Printf("[GetQRChannel] Canal QR original fechado")
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
