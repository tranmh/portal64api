// Main application logic and initialization

// Global API instance
const api = new Portal64API();

// Tab functionality
class TabManager {
    constructor(containerSelector) {
        this.container = document.querySelector(containerSelector);
        this.init();
    }

    init() {
        if (!this.container) return;

        const tabNavItems = this.container.querySelectorAll('.tab-nav-item');
        const tabContents = this.container.querySelectorAll('.tab-content');

        tabNavItems.forEach(item => {
            item.addEventListener('click', (e) => {
                e.preventDefault();
                const targetId = item.getAttribute('data-tab');
                this.switchTab(targetId, tabNavItems, tabContents);
            });
        });

        // Check for hash in URL and switch to that tab
        const hash = window.location.hash.replace('#', '').split('?')[0]; // Remove query params from hash
        if (hash) {
            const hashTab = document.querySelector(`[data-tab="${hash}"]`);
            if (hashTab) {
                this.switchTab(hash, tabNavItems, tabContents);
                return;
            }
        }

        // Show first tab by default if no hash match
        if (tabNavItems.length > 0) {
            const firstTab = tabNavItems[0].getAttribute('data-tab');
            this.switchTab(firstTab, tabNavItems, tabContents);
        }
    }

    switchTab(targetId, navItems, contents) {
        // Remove active class from all nav items and contents
        navItems.forEach(item => item.classList.remove('active'));
        contents.forEach(content => content.classList.remove('active'));

        // Add active class to target nav item and content
        const targetNav = document.querySelector(`[data-tab="${targetId}"]`);
        const targetContent = document.getElementById(targetId);

        if (targetNav) targetNav.classList.add('active');
        if (targetContent) targetContent.classList.add('active');
    }
}

// Modal functionality
class ModalManager {
    constructor() {
        this.init();
    }

    init() {
        // Close modal when clicking outside
        document.addEventListener('click', (e) => {
            if (e.target.classList.contains('modal')) {
                this.closeModal(e.target.id);
            }
        });

        // Close modal with close button
        document.addEventListener('click', (e) => {
            if (e.target.classList.contains('modal-close')) {
                const modal = e.target.closest('.modal');
                if (modal) {
                    this.closeModal(modal.id);
                }
            }
        });

        // Close modal with Escape key
        document.addEventListener('keydown', (e) => {
            if (e.key === 'Escape') {
                const activeModal = document.querySelector('.modal.show');
                if (activeModal) {
                    this.closeModal(activeModal.id);
                }
            }
        });
    }

    openModal(modalId) {
        const modal = document.getElementById(modalId);
        if (modal) {
            modal.classList.add('show');
            document.body.style.overflow = 'hidden';
        }
    }

    closeModal(modalId) {
        const modal = document.getElementById(modalId);
        if (modal) {
            modal.classList.remove('show');
            document.body.style.overflow = '';
        }
    }
}

// Code display and copy functionality
class CodeDisplayManager {
    constructor() {
        this.init();
    }

    init() {
        document.addEventListener('click', (e) => {
            if (e.target.classList.contains('code-copy')) {
                this.copyCode(e.target);
            }
        });
    }

    async copyCode(button) {
        const codeContainer = button.closest('.code-container');
        const codeContent = codeContainer.querySelector('.code-content');
        const text = codeContent.textContent;

        const success = await Utils.copyToClipboard(text);
        
        const originalText = button.textContent;
        button.textContent = success ? 'Copied!' : 'Failed';
        button.style.color = success ? '#28a745' : '#dc3545';

        setTimeout(() => {
            button.textContent = originalText;
            button.style.color = '';
        }, 2000);
    }

    displayResponse(containerId, data, title = 'API Response') {
        const container = document.getElementById(containerId);
        if (!container) return;

        const formattedData = Utils.formatJSON(data);
        
        container.innerHTML = `
            <div class="code-container">
                <div class="code-header">
                    <h4 class="code-title">${Utils.sanitizeHTML(title)}</h4>
                    <button class="code-copy">Copy</button>
                </div>
                <div class="code-content">${Utils.sanitizeHTML(formattedData)}</div>
            </div>
        `;
    }
}

// Advanced search toggle functionality
class AdvancedSearchManager {
    constructor() {
        this.init();
    }

    init() {
        document.addEventListener('click', (e) => {
            if (e.target.classList.contains('advanced-toggle')) {
                this.toggleAdvancedSearch(e.target);
            }
        });
    }

    toggleAdvancedSearch(button) {
        const form = button.closest('form');
        const advancedSection = form.querySelector('.advanced-search');
        
        if (advancedSection) {
            advancedSection.classList.toggle('show');
            button.textContent = advancedSection.classList.contains('show') 
                ? 'Erweiterte Optionen ausblenden' 
                : 'Erweiterte Optionen anzeigen';
        }
    }
}

// Health check functionality
async function performHealthCheck() {
    try {
        Utils.showLoading('health-result');
        const result = await api.healthCheck();
        
        const container = document.getElementById('health-result');
        container.innerHTML = `
            <div class="alert alert-success">
                <h4>✓ API is healthy</h4>
                <p>Status: ${result.status}</p>
                <p>Version: ${result.version}</p>
            </div>
        `;
    } catch (error) {
        Utils.showError('health-result', `Gesundheitsprüfung fehlgeschlagen: ${error.message}`);
    }
}

// Initialize application when DOM is loaded
document.addEventListener('DOMContentLoaded', function() {
    // Initialize managers
    new TabManager('.tabs');
    new ModalManager();
    new CodeDisplayManager();
    new AdvancedSearchManager();
    
    // Perform health check on index page
    if (document.getElementById('health-result')) {
        performHealthCheck();
    }
    
    console.log('Portal64 API Demo initialized');
});