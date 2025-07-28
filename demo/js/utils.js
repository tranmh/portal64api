// Utility functions for the demo application

class Utils {
    // Show loading state
    static showLoading(containerId) {
        const container = document.getElementById(containerId);
        if (container) {
            container.innerHTML = `
                <div class="loading show">
                    <div class="spinner"></div>
                    <p>Loading...</p>
                </div>
            `;
        }
    }

    // Hide loading state
    static hideLoading(containerId) {
        const loading = document.querySelector(`#${containerId} .loading`);
        if (loading) {
            loading.remove();
        }
    }

    // Display error message
    static showError(containerId, message) {
        const container = document.getElementById(containerId);
        if (container) {
            container.innerHTML = `
                <div class="alert alert-error">
                    <h4>Error</h4>
                    <p>${message}</p>
                </div>
            `;
        }
    }

    // Display success message
    static showSuccess(containerId, message) {
        const container = document.getElementById(containerId);
        if (container) {
            container.innerHTML = `
                <div class="alert alert-success">
                    <h4>Success</h4>
                    <p>${message}</p>
                </div>
            `;
        }
    }

    // Format JSON for display
    static formatJSON(obj) {
        return JSON.stringify(obj, null, 2);
    }

    // Copy text to clipboard
    static async copyToClipboard(text) {
        try {
            await navigator.clipboard.writeText(text);
            return true;
        } catch (err) {
            console.error('Failed to copy text: ', err);
            return false;
        }
    }

    // Format date for display
    static formatDate(dateString) {
        if (!dateString) return 'N/A';
        const date = new Date(dateString);
        return date.toLocaleDateString('de-DE', {
            year: 'numeric',
            month: '2-digit',
            day: '2-digit'
        });
    }

    // Validate player ID format (C0101-1014, UNKNOWN-12345, D300H-1014, etc.)
    static validatePlayerID(id) {
        // Allow format: <CLUBID>-<PERSONID> where CLUBID can be alphanumeric and PERSONID is numeric
        const playerRegex = /^[A-Z0-9]+-\d+$/;
        return playerRegex.test(id);
    }

    // Validate club ID format (C0101, UNKNOWN, D300H, etc.)
    static validateClubID(id) {
        // Allow alphanumeric club IDs 
        const clubRegex = /^[A-Z0-9]+$/;
        return clubRegex.test(id);
    }

    // Validate tournament ID format (C529-K00-HT1, C339-400-442, etc.)
    static validateTournamentID(id) {
        // Allow various tournament ID formats: alphanumeric with dashes
        const tournamentRegex = /^[A-Z0-9]+-[A-Z0-9]+-[A-Z0-9]+$/;
        return tournamentRegex.test(id) || /^[A-Z0-9]+-\d+-\d+$/.test(id);
    }

    // Create pagination info text
    static getPaginationInfo(meta) {
        if (!meta) return '';
        const { offset = 0, limit = 20, total = 0 } = meta;
        const start = offset + 1;
        const end = Math.min(offset + limit, total);
        return `Showing ${start}-${end} of ${total} results`;
    }

    // Debounce function for search inputs
    static debounce(func, wait) {
        let timeout;
        return function executedFunction(...args) {
            const later = () => {
                clearTimeout(timeout);
                func(...args);
            };
            clearTimeout(timeout);
            timeout = setTimeout(later, wait);
        };
    }

    // Sanitize HTML to prevent XSS
    static sanitizeHTML(str) {
        const temp = document.createElement('div');
        temp.textContent = str;
        return temp.innerHTML;
    }

    // Get form data as object
    static getFormData(formElement) {
        const formData = new FormData(formElement);
        const data = {};
        
        for (let [key, value] of formData.entries()) {
            if (value !== '') {
                data[key] = value;
            }
        }
        
        return data;
    }

    // Reset form
    static resetForm(formElement) {
        formElement.reset();
        // Remove any error states
        const errorElements = formElement.querySelectorAll('.form-error.show');
        errorElements.forEach(el => el.classList.remove('show'));
        
        const errorInputs = formElement.querySelectorAll('.form-input.error, .form-select.error');
        errorInputs.forEach(el => el.classList.remove('error'));
    }

    // Show form validation error
    static showFormError(fieldName, message) {
        const field = document.querySelector(`[name="${fieldName}"]`);
        const errorElement = document.querySelector(`#${fieldName}-error`);
        
        if (field) {
            field.classList.add('error');
        }
        
        if (errorElement) {
            errorElement.textContent = message;
            errorElement.classList.add('show');
        }
    }

    // Clear form validation errors
    static clearFormErrors(formElement) {
        const errorElements = formElement.querySelectorAll('.form-error.show');
        errorElements.forEach(el => el.classList.remove('show'));
        
        const errorInputs = formElement.querySelectorAll('.form-input.error, .form-select.error');
        errorInputs.forEach(el => el.classList.remove('error'));
    }
}