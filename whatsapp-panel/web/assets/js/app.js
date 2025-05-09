// Funções principais do painel WhatsApp

// Funções auxiliares para interação com a interface

// Atualizar automaticamente o status das sessões
function setupSessionStatusPolling() {
    setInterval(function() {
        if (document.getElementById('sessions')) {
            htmx.ajax('GET', '/sessions', {target: '#sessions', swap: 'innerHTML'});
        }
    }, 5000);
}

// Fechar modal do QR Code
function closeModal() {
    const modal = document.getElementById('qrcodeModal');
    if (modal) {
        modal.classList.add('hidden');
    }
}

// Formatar número de telefone
function formatPhoneNumber(phone) {
    if (!phone) return 'Não disponível';
    return phone.replace(/(\d{2})(\d{2})(\d{5})(\d{4})/, '+$1 ($2) $3-$4');
}

// Formatar data
function formatDate(date) {
    return new Date(date).toLocaleString('pt-BR');
}

// Inicializar quando o documento estiver pronto
document.addEventListener('DOMContentLoaded', function() {
    setupSessionStatusPolling();
});

// Handler global para o modal de QR Code: iniciar polling de conexão
document.body.addEventListener('htmx:afterSwap', function(evt) {
    if (evt.detail.target.id === 'qrcodeModal') {
        // Exibir o modal
        document.getElementById('qrcodeModal').classList.remove('hidden');
        // Obter session_id do atributo data-session-id
        const sessionId = document.querySelector('#qrcodeModal [data-session-id]')?.getAttribute('data-session-id');
        if (!sessionId) return;
        // Iniciar polling para verificar conexão, ignorando os primeiros segundos para evitar falso positivo imediato
        let pollCount = 0;
        const skipInitial = 5; // segundos para ignorar polling
        const connectionInterval = setInterval(() => {
            pollCount++;
            if (pollCount <= skipInitial) return;
            fetch(`/connection-status?session_id=${sessionId}`)
                .then(res => res.json())
                .then(data => {
                    if (data.connected) {
                        clearInterval(connectionInterval);
                        closeModal();
                        htmx.ajax('GET', '/sessions', { target: '#sessions', swap: 'innerHTML' });
                    }
                })
                .catch(() => clearInterval(connectionInterval));
        }, 1000);
    }
});

// Handler para confirmação de deleção
document.body.addEventListener('htmx:confirm', function(evt) {
    // Personalizar mensagem de confirmação
    if (evt.detail.path.includes('/sessions/')) {
        evt.detail.question = 'Tem certeza que deseja desconectar esta sessão do WhatsApp?';
    }
});

// Handler para erros de requisição
document.body.addEventListener('htmx:responseError', function(evt) {
    console.error('Erro na requisição:', evt.detail.error);
    // Implementar notificação visual do erro
});