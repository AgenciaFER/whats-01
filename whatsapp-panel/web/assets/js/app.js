// Arquivo: whatsapp-panel/web/assets/js/app.js
// Substitua todo o conteúdo do arquivo pelo código abaixo:

// Sistema de notificações melhorado
const notifications = {
    queue: [],
    processing: false,

    show(message, type = "info") {
        this.queue.push({ message, type });
        if (!this.processing) {
            this.processQueue();
        }
    },

    async processQueue() {
        if (this.queue.length === 0) {
            this.processing = false;
            return;
        }

        this.processing = true;
        const { message, type } = this.queue.shift();
        
        const colors = {
            success: "bg-green-500",
            error: "bg-red-500",
            info: "bg-blue-500",
            warning: "bg-yellow-500"
        };

        const notification = document.createElement("div");
        notification.className = `notification ${colors[type]}`;
        notification.setAttribute('role', 'alert');
        notification.innerHTML = `
            <div class="flex items-center">
                <div class="flex-1">${message}</div>
                <button onclick="this.parentElement.parentElement.remove()" class="ml-4 text-white hover:text-gray-200">
                    <svg class="w-4 h-4" viewBox="0 0 20 20" fill="currentColor">
                        <path fill-rule="evenodd" d="M4.293 4.293a1 1 0 011.414 0L10 8.586l4.293-4.293a1 1 0 111.414 1.414L11.414 10l4.293 4.293a1 1 0 01-1.414 1.414L10 11.414l-4.293 4.293a1 1 0 01-1.414-1.414L8.586 10 4.293 5.707a1 1 0 010-1.414z" clip-rule="evenodd" />
                    </svg>
                </button>
            </div>`;
        document.body.appendChild(notification);

        // Animate in
        await new Promise(resolve => setTimeout(resolve, 4000));
        
        // Animate out if still in document
        if (document.body.contains(notification)) {
            notification.classList.add('animate-slide-out');
            await new Promise(resolve => setTimeout(resolve, 300));
            if (document.body.contains(notification)) {
                notification.remove();
            }
        }

        // Process next notification
        this.processQueue();
    },

    success(message) { this.show(message, "success"); },
    error(message) { this.show(message, "error"); },
    info(message) { this.show(message, "info"); },
    warning(message) { this.show(message, "warning"); }
};

// Gerenciamento de estado das sessões
const sessionManager = {
    pollInterval: null,
    retryAttempts: {},
    maxRetries: 3,

    startPolling() {
        this.stopPolling();
        this.pollInterval = setInterval(() => this.updateSessions(), 5000);
    },

    stopPolling() {
        if (this.pollInterval) {
            clearInterval(this.pollInterval);
            this.pollInterval = null;
        }
    },

    async updateSessions() {
        try {
            const response = await fetch('/sessions/list');
            if (!response.ok) throw new Error('Falha ao atualizar sessões');
            
            // Recarregar a página para atualizar a lista de sessões
            window.location.reload();
        } catch (error) {
            console.error("Falha ao atualizar sessões:", error);
        }
    },

    resetRetries(sessionId) {
        delete this.retryAttempts[sessionId];
    }
};

// Função para lidar com o modal
function closeModal() {
    const modal = document.getElementById('qrcodeModal');
    if (modal) {
        modal.remove();
        document.body.style.overflow = '';
    }
    
    const messageModal = document.getElementById('messageModal');
    if (messageModal) {
        messageModal.remove();
        document.body.style.overflow = '';
    }
}

// Função para iniciar o polling de status
function startConnectionCheck(sessionId) {
    if (!sessionId) return;
    
    let attempts = 0;
    const maxAttempts = 60; // 60 tentativas = 2 minutos (a cada 2 segundos)
    const countdownElement = document.getElementById('countdown');
    
    const checkInterval = setInterval(async () => {
        attempts++;
        try {
            const response = await fetch(`/connection-status?session_id=${sessionId}`);
            if (!response.ok) throw new Error('Erro na verificação de status');
            
            const data = await response.json();
            if (data.connected) {
                clearInterval(checkInterval);
                if (countdownElement) {
                    countdownElement.innerHTML = '<span class="text-green-500">Conectado com sucesso!</span>';
                }
                
                // Atualizar a lista de sessões e fechar o modal após 1 segundo
                setTimeout(() => {
                    closeModal();
                    // Recarregar a página
                    window.location.reload();
                }, 1000);
                
                return;
            }
            
            // Atualizar o contador regressivo
            if (countdownElement) {
                const timeLeft = maxAttempts - attempts;
                countdownElement.innerHTML = `<div class="loading-spinner"></div> Aguardando conexão... (${timeLeft}s)`;
            }
            
            if (attempts >= maxAttempts) {
                clearInterval(checkInterval);
                if (countdownElement) {
                    countdownElement.innerHTML = '<span class="text-red-500">Tempo esgotado. Tente novamente.</span>';
                }
                
                // Fechar o modal após mostrar a mensagem de tempo esgotado
                setTimeout(closeModal, 3000);
            }
        } catch (error) {
            console.error('Erro ao verificar conexão:', error);
            if (countdownElement) {
                countdownElement.innerHTML = '<span class="text-red-500">Erro ao verificar conexão</span>';
            }
            
            clearInterval(checkInterval);
            setTimeout(closeModal, 3000);
        }
    }, 2000); // Checar a cada 2 segundos
}

// Função para enviar mensagem
function sendMessage() {
    const sessionId = document.getElementById('sessionId')?.value;
    const phoneNumber = document.getElementById('phoneNumber')?.value;
    const message = document.getElementById('message')?.value;
    const resultDiv = document.getElementById('messageResult');
    
    if (!sessionId || !phoneNumber || !message) {
        if (resultDiv) {
            resultDiv.className = "p-3 rounded-md text-center bg-red-100 text-red-700";
            resultDiv.innerHTML = "Preencha todos os campos";
            resultDiv.classList.remove("hidden");
        }
        return;
    }
    
    // Adicionar indicador de carregamento
    const sendButton = document.querySelector('button[onclick="sendMessage()"]');
    if (sendButton) {
        sendButton.disabled = true;
        sendButton.innerHTML = '<div class="loading-spinner"></div> Enviando...';
    }
    
    fetch(`/sessions/${sessionId}/message`, {
        method: 'POST',
        headers: {
            'Content-Type': 'application/json',
        },
        body: JSON.stringify({
            phone_number: phoneNumber,
            message: message
        })
    })
    .then(response => response.json())
    .then(data => {
        if (sendButton) {
            sendButton.disabled = false;
            sendButton.innerHTML = 'Enviar Mensagem';
        }
        
        if (resultDiv) {
            if (data.success) {
                resultDiv.className = "p-3 rounded-md text-center bg-green-100 text-green-700";
                resultDiv.innerHTML = "Mensagem enviada com sucesso!";
                
                // Limpar o formulário após sucesso
                document.getElementById('messageForm')?.reset();
                
                // Fechar o modal após 2 segundos
                setTimeout(() => {
                    closeModal();
                }, 2000);
            } else {
                resultDiv.className = "p-3 rounded-md text-center bg-red-100 text-red-700";
                resultDiv.innerHTML = `Erro: ${data.error || data.details || "Ocorreu um erro desconhecido"}`;
            }
            resultDiv.classList.remove("hidden");
        }
    })
    .catch(error => {
        if (sendButton) {
            sendButton.disabled = false;
            sendButton.innerHTML = 'Enviar Mensagem';
        }
        
        if (resultDiv) {
            resultDiv.className = "p-3 rounded-md text-center bg-red-100 text-red-700";
            resultDiv.innerHTML = `Erro: ${error.message}`;
            resultDiv.classList.remove("hidden");
        }
    });
}

// Inicialização - Carrega após o DOM estar pronto
document.addEventListener("DOMContentLoaded", function() {
    // Inicializa o polling de sessões
    sessionManager.startPolling();
    
    // Adiciona listener para o botão de conectar
    const connectBtn = document.getElementById('connectBtn');
    if (connectBtn) {
        connectBtn.addEventListener('click', function() {
            // Criar div para o modal
            const modalContainer = document.createElement('div');
            modalContainer.id = 'qrcodeModal';
            document.body.appendChild(modalContainer);
            
            // Carregar QR Code
            fetch('/qrcode')
                .then(response => {
                    if (!response.ok) throw new Error('Erro ao carregar QR Code');
                    return response.text();
                })
                .then(html => {
                    // Adicionar o HTML do QR code ao modal
                    modalContainer.innerHTML = html;
                    
                    // Configurar eventos do modal
                    document.getElementById('qrcodeModalBackdrop')?.addEventListener('click', function(event) {
                        if (event.target === this) {
                            closeModal();
                        }
                    });
                    
                    // Capturar o ID da sessão e iniciar verificação
                    const sessionIdEl = modalContainer.querySelector('[data-session-id]');
                    if (sessionIdEl) {
                        const sessionId = sessionIdEl.getAttribute('data-session-id');
                        if (sessionId) {
                            startConnectionCheck(sessionId);
                        }
                    }
                })
                .catch(error => {
                    console.error('Erro ao carregar QR Code:', error);
                    notifications.error('Erro ao carregar QR Code. Por favor, tente novamente.');
                    closeModal();
                });
        });
    }
    
    // Adiciona listeners para botões de mensagem
    document.body.addEventListener('click', function(event) {
        const btn = event.target.closest('button[hx-get*="/message"]');
        if (!btn) return;
        
        event.preventDefault();
        event.stopPropagation();
        
        // Extrair o ID da sessão
        const hxGet = btn.getAttribute('hx-get');
        const sessionIdMatch = hxGet.match(/\/sessions\/(.+?)\/message/);
        if (!sessionIdMatch) return;
        
        const sessionId = sessionIdMatch[1];
        
        // Criar div para o modal de mensagem
        const modalContainer = document.createElement('div');
        modalContainer.id = 'messageModal';
        document.body.appendChild(modalContainer);
        
        // Carregar formulário de mensagem
        fetch(`/sessions/${sessionId}/message`)
            .then(response => {
                if (!response.ok) throw new Error('Erro ao carregar formulário de mensagem');
                return response.text();
            })
            .then(html => {
                // Adicionar o HTML do formulário ao modal
                modalContainer.innerHTML = `
                    <div class="fixed inset-0 flex items-center justify-center z-[9999]">
                        <div class="fixed inset-0 bg-black opacity-50" onclick="closeModal()"></div>
                        <div class="relative z-10">
                            ${html}
                        </div>
                    </div>
                `;
            })
            .catch(error => {
                console.error('Erro ao carregar formulário de mensagem:', error);
                notifications.error('Erro ao carregar formulário de mensagem. Por favor, tente novamente.');
                closeModal();
            });
    });
    
    // Adiciona listener para tecla Escape
    document.addEventListener('keydown', function(event) {
        if (event.key === 'Escape') {
            closeModal();
        }
    });
});