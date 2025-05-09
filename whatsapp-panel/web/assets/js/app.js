// Funções de notificação
const notifications = {
    show: function(message, type = 'info') {
        const notificationContainer = document.getElementById('notification-container');
        if (!notificationContainer) {
            // Criar container de notificações se não existir
            const container = document.createElement('div');
            container.id = 'notification-container';
            container.style.position = 'fixed';
            container.style.top = '1rem';
            container.style.right = '1rem';
            container.style.zIndex = '9999';
            document.body.appendChild(container);
        }
        
        const notification = document.createElement('div');
        notification.className = `notification ${type}`;
        notification.innerHTML = `
            <div class="p-4 rounded-lg shadow-lg mb-3 ${type === 'error' ? 'bg-red-500' : type === 'success' ? 'bg-green-500' : 'bg-blue-500'} text-white">
                <div class="flex justify-between items-center">
                    <span>${message}</span>
                    <button class="ml-4 text-white hover:text-gray-200">&times;</button>
                </div>
            </div>
        `;
        
        // Adicionar event listener para fechar notificação
        notification.querySelector('button').addEventListener('click', function() {
            notification.remove();
        });
        
        document.getElementById('notification-container').appendChild(notification);
        
        // Auto-remover após 5 segundos
        setTimeout(() => {
            if (notification.parentNode) {
                notification.remove();
            }
        }, 5000);
    }
};

// Funções para modal
function closeModal() {
    const modal = document.getElementById('qrcodeModalBackdrop');
    if (modal) {
        modal.remove();
        document.body.style.overflow = '';
        
        // Limpar qualquer intervalo ativo
        if (window.activeCheckInterval) {
            clearInterval(window.activeCheckInterval);
            window.activeCheckInterval = null;
        }
    }
}

function showModal() {
    const modal = document.getElementById('qrcodeModalBackdrop');
    if (modal) {
        document.body.style.overflow = 'hidden';
    }
}

// Função para verificar status de conexão
async function checkConnection(sessionId) {
    if (!sessionId) {
        console.error("checkConnection chamado sem sessionId");
        return false;
    }
    
    try {
        console.log(`Verificando conexão para sessão ${sessionId}...`);
        const response = await fetch(`/connection-status?session_id=${sessionId}`);
        
        if (!response.ok) {
            console.error(`Erro ao verificar conexão: ${response.status}`);
            return false;
        }
        
        const data = await response.json();
        console.log(`Status de conexão: ${JSON.stringify(data)}`);
        return data.connected === true;
    } catch (error) {
        console.error(`Erro ao verificar conexão: ${error}`);
        return false;
    }
}

// Função para inicializar verificação de conexão
function initializeConnection(sessionId) {
    if (!sessionId) {
        console.error("initializeConnection chamado sem sessionId");
        return;
    }
    
    console.log(`Iniciando verificação de conexão para sessão ${sessionId}`);
    
    // Limpar intervalo anterior se existir
    if (window.activeCheckInterval) {
        clearInterval(window.activeCheckInterval);
    }
    
    let attempts = 0;
    const maxAttempts = 60; // 2 minutos (verificando a cada 2 segundos)
    const countdownElement = document.getElementById('countdown');
    
    if (countdownElement) {
        countdownElement.innerHTML = '<div class="loading-spinner"></div><span>Aguardando conexão...</span>';
    }
    
    window.activeCheckInterval = setInterval(async () => {
        attempts++;
        
        if (attempts > maxAttempts) {
            clearInterval(window.activeCheckInterval);
            window.activeCheckInterval = null;
            notifications.show("Tempo limite de conexão excedido. Tente novamente.", "error");
            closeModal();
            return;
        }
        
        const isConnected = await checkConnection(sessionId);
        console.log(`Tentativa ${attempts}: conectado = ${isConnected}`);
        
        if (isConnected) {
            clearInterval(window.activeCheckInterval);
            window.activeCheckInterval = null;
            notifications.show("WhatsApp conectado com sucesso!", "success");
            
            // Atualizar a lista de sessões (se estiver na página correta)
            if (window.location.pathname === '/sessions/' || window.location.pathname === '/') {
                window.location.reload();
            } else {
                closeModal();
            }
        }
    }, 2000); // Verificar a cada 2 segundos
}

// Event listener para manipulação do QR code
document.addEventListener('DOMContentLoaded', function() {
    // Adicionar container de notificações
    if (!document.getElementById('notification-container')) {
        const container = document.createElement('div');
        container.id = 'notification-container';
        container.style.position = 'fixed';
        container.style.top = '1rem';
        container.style.right = '1rem';
        container.style.zIndex = '9999';
        document.body.appendChild(container);
    }
    
    // Event listener para htmx após swap
    document.body.addEventListener('htmx:afterSwap', function(evt) {
        if (evt.detail.target.id === 'qrcodeModal' || 
            (evt.detail.target.querySelector && evt.detail.target.querySelector('#qrcodeModalBackdrop'))) {
            
            showModal();
            
            // Extrair o sessionId do novo modal
            const modalContent = document.querySelector('[data-session-id]');
            if (modalContent) {
                const sessionId = modalContent.getAttribute('data-session-id');
                if (sessionId) {
                    console.log(`Modal QR aberto para sessão: ${sessionId}`);
                    setTimeout(() => initializeConnection(sessionId), 500);
                }
            }
        }
            }
        }
    });
});