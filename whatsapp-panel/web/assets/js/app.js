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

// Funções para gerenciar a conexão WhatsApp
async function checkConnection(sessionId) {
    try {
        console.log(`Verificando conexão para sessão ${sessionId}...`);
        const response = await fetch(`/connection-status?session_id=${sessionId}`);
        if (!response.ok) throw new Error('Network response was not ok');
        const data = await response.json();
        
        console.log(`Status da conexão para sessão ${sessionId}:`, data);
        
        if (data.connected) {
            notifications.success('WhatsApp conectado com sucesso!');
            closeModal();
            sessionManager.updateSessions();
            return true;
        }
        return data.connected;
    } catch (error) {
        console.error('Error checking connection:', error);
        return false;
    }
}

// Função para iniciar o polling de status
function startConnectionCheck(sessionId) {
    console.log(`Iniciando verificação de conexão para sessão ${sessionId}`);
    let attempts = 0;
    const maxAttempts = 60; // 60 tentativas = 2 minutos (2s cada)
    
    const checkInterval = setInterval(async () => {
        attempts++;
        console.log(`Tentativa ${attempts} de verificação para sessão ${sessionId}`);
        const isConnected = await checkConnection(sessionId);
        
        if (isConnected || attempts >= maxAttempts) {
            clearInterval(checkInterval);
            if (!isConnected && attempts >= maxAttempts) {
                notifications.warning('Tempo limite de conexão excedido. Tente novamente.');
                closeModal();
            }
        }
    }, 2000); // Verifica a cada 2 segundos
}

// Função para iniciar processo de conexão
function initializeConnection(sessionId) {
    console.log(`Inicializando conexão para sessão ${sessionId}`);
    startConnectionCheck(sessionId);
}

// Modal handling functions
function closeModal() {
    const modal = document.getElementById('qrcodeModalBackdrop');
    if (modal) {
        modal.remove();
        document.body.style.overflow = '';
    }
}