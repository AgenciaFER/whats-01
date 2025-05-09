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
        const sessionsElement = document.getElementById("sessions");
        if (!sessionsElement) return;

        try {
            await htmx.ajax("GET", "/sessions", {
                target: "#sessions",
                swap: "innerHTML",
                headers: { "X-Silent": "true" }
            });
        } catch (error) {
            console.error("Failed to update sessions:", error);
        }
    },

    handleSessionError(sessionId) {
        if (!this.retryAttempts[sessionId]) {
            this.retryAttempts[sessionId] = 0;
        }

        this.retryAttempts[sessionId]++;
        
        if (this.retryAttempts[sessionId] >= this.maxRetries) {
            notifications.error(`Falha na conexão da sessão ${sessionId}. Tente reconectar.`);
            return false;
        }
        return true;
    },

    resetRetries(sessionId) {
        delete this.retryAttempts[sessionId];
    }
};

// Gerenciamento do QR Code
const qrCodeManager = {
    currentSessionId: null,
    checkInterval: null,
    expirationTimer: null,
    qrCodeExpirationTime: 60, // segundos

    setupNewQRCode(sessionId) {
        this.cleanup();
        this.currentSessionId = sessionId;
        
        // Iniciar timer de expiração
        let timeLeft = this.qrCodeExpirationTime;
        const countdownElement = document.getElementById('countdown');
        
        if (countdownElement) {
            const updateCountdown = () => {
                countdownElement.textContent = `Expira em ${timeLeft} segundos`;
                if (timeLeft <= 0) {
                    this.handleExpiration();
                    return;
                }
                timeLeft--;
            };
            
            updateCountdown();
            this.expirationTimer = setInterval(updateCountdown, 1000);
        }

        // Iniciar verificação de status
        this.startStatusCheck();
    },

    startStatusCheck() {
        if (!this.currentSessionId) return;

        this.checkInterval = setInterval(async () => {
            try {
                const response = await fetch(`/connection-status?session_id=${this.currentSessionId}`);
                const data = await response.json();

                if (data.connected) {
                    this.handleSuccess();
                }
            } catch (error) {
                if (!sessionManager.handleSessionError(this.currentSessionId)) {
                    this.cleanup();
                    closeModal();
                }
            }
        }, 1000);
    },

    handleSuccess() {
        this.cleanup();
        closeModal();
        sessionManager.resetRetries(this.currentSessionId);
        sessionManager.updateSessions();
        notifications.success("WhatsApp conectado com sucesso!");
    },

    handleExpiration() {
        this.cleanup();
        notifications.warning("QR Code expirado. Tente novamente.");
        closeModal();
    },

    cleanup() {
        this.currentSessionId = null;
        if (this.checkInterval) {
            clearInterval(this.checkInterval);
            this.checkInterval = null;
        }
        if (this.expirationTimer) {
            clearInterval(this.expirationTimer);
            this.expirationTimer = null;
        }
    }
};

// Gerenciador de modais
const modalManager = {
    showModal(modalId) {
        const modal = document.getElementById(modalId);
        if (modal) {
            modal.classList.remove('hidden');
            document.body.classList.add('overflow-hidden');
        }
    },
    
    hideModal(modalId) {
        const modal = document.getElementById(modalId);
        if (modal) {
            modal.classList.add('hidden');
            document.body.classList.remove('overflow-hidden');
        }
    }
};

// Gerenciador de mensagens
const messageManager = {
    sendMessage(sessionId, phoneNumber, message) {
        return fetch(`/sessions/${sessionId}/message`, {
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
            if (data.success) {
                notifications.success("Mensagem enviada com sucesso!");
                return true;
            } else {
                notifications.error(`Erro: ${data.error || data.details || "Ocorreu um erro ao enviar a mensagem"}`);
                return false;
            }
        })
        .catch(error => {
            notifications.error(`Erro: ${error.message}`);
            return false;
        });
    }
};

// Modal handling functions
function closeModal() {
    const modal = document.getElementById('qrcodeModal');
    if (modal) {
        modal.classList.add('hidden');
        document.body.style.overflow = '';
    }
}

function showModal() {
    const modal = document.getElementById('qrcodeModal');
    if (modal) {
        modal.classList.remove('hidden');
        document.body.style.overflow = 'hidden';
    }
}

document.addEventListener('DOMContentLoaded', function() {
    // Inicializa os listeners para o modal
    document.body.addEventListener('htmx:afterSwap', function(evt) {
        if (evt.detail.target.id === 'qrcodeModal') {
            showModal();
        }
    });
});

// UI Helpers
document.addEventListener('keydown', function(event) {
    if (event.key === 'Escape') {
        closeModal();
    }
});

// Função para gerenciar o modal
function handleModal() {
    document.body.style.overflow = 'hidden';
    
    // Fechar modal ao clicar fora
    document.getElementById('qrcodeModalBackdrop')?.addEventListener('click', function(event) {
        if (event.target === this) {
            closeModal();
        }
    });
}

// Inicializar modal quando carregado
document.addEventListener('htmx:afterSwap', function(event) {
    if (event.detail.target.id === 'qrcodeModal') {
        handleModal();
    }
});

function closeModalOnBackdrop(event) {
    if (event.target.id === 'qrcodeModalBackdrop') {
        closeModal();
    }
}

function setLoadingState(element, loading = true) {
    if (!element) return;
    
    if (loading) {
        element.classList.add('loading');
        element.setAttribute('disabled', 'disabled');
    } else {
        element.classList.remove('loading');
        element.removeAttribute('disabled');
    }
}

// Formatadores
function formatPhoneNumber(phone) {
    if (!phone) return "Não disponível";
    return phone.replace(/(\d{2})(\d{2})(\d{5})(\d{4})/, "+$1 ($2) $3-$4");
}

function formatDate(date) {
    return new Date(date).toLocaleString("pt-BR");
}

// Funções para gerenciar a conexão WhatsApp
async function checkConnection(sessionId) {
    try {
        const response = await fetch(`/connection-status?session_id=${sessionId}`);
        if (!response.ok) throw new Error('Network response was not ok');
        const data = await response.json();
        
        if (data.connected) {
            notifications.show('WhatsApp conectado com sucesso!', 'success');
            closeModal();
            document.dispatchEvent(new Event('sessionUpdate'));
        }
        return data.connected;
    } catch (error) {
        console.error('Error checking connection:', error);
        return false;
    }
}

// Função para iniciar o polling de status
function startConnectionCheck(sessionId) {
    let attempts = 0;
    const maxAttempts = 30; // 30 tentativas = 1 minuto
    
    const checkInterval = setInterval(async () => {
        attempts++;
        const isConnected = await checkConnection(sessionId);
        
        if (isConnected || attempts >= maxAttempts) {
            clearInterval(checkInterval);
            if (!isConnected && attempts >= maxAttempts) {
                notifications.show('Tempo limite de conexão excedido. Tente novamente.', 'error');
                closeModal();
            }
        }
    }, 2000); // Verifica a cada 2 segundos
}

// Função para iniciar processo de conexão
function initializeConnection(sessionId) {
    startConnectionCheck(sessionId);
}

// Inicialização
document.addEventListener("DOMContentLoaded", () => {
    sessionManager.startPolling();
});

// Handlers de eventos HTMX
document.body.addEventListener("htmx:beforeRequest", function(evt) {
    const target = evt.detail.target;
    setLoadingState(target, true);
});

document.body.addEventListener("htmx:afterRequest", function(evt) {
    const target = evt.detail.target;
    setLoadingState(target, false);

    // Notificar sucesso em operações específicas
    if (evt.detail.successful) {
        if (evt.detail.pathInfo.requestPath.includes("/disconnect")) {
            notifications.success("Sessão desconectada com sucesso");
        } else if (evt.detail.pathInfo.requestPath.includes("/sessions") && evt.detail.xhr.status === 204) {
            notifications.success("Sessão removida com sucesso");
        }
    }
});

document.body.addEventListener("htmx:afterSwap", function(evt) {
    if (evt.detail.target.id === "qrcodeModal") {
        modalManager.showModal('qrcodeModal');
        document.body.style.overflow = 'hidden';
    } else if (evt.detail.target.id === "messageModal") {
        modalManager.showModal('messageModal');
    }
});

document.body.addEventListener("htmx:beforeSwap", function(evt) {
    if (evt.detail.target.id === 'qrcodeModal') {
        const oldModal = document.getElementById('qrcodeModalBackdrop');
        if (oldModal) {
            oldModal.remove();
        }
    }
});

document.body.addEventListener("htmx:confirm", function(evt) {
    if (evt.detail.path.includes("/sessions/")) {
        evt.detail.question = "Tem certeza que deseja desconectar esta sessão do WhatsApp?";
    }
});

document.body.addEventListener("htmx:responseError", function(evt) {
    console.error("Erro na requisição:", evt.detail.error);
    notifications.error("Ocorreu um erro. Por favor, tente novamente.");
});