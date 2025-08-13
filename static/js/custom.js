// Custom JavaScript for the workflow manager

// Initialize HTMX
document.addEventListener('DOMContentLoaded', function() {
    // Add any custom initialization code here
    console.log('Workflow Manager initialized');
});

// Function to show toast notifications
function showToast(message, type = 'info') {
    // Create toast container if it doesn't exist
    let toastContainer = document.getElementById('toast-container');
    if (!toastContainer) {
        toastContainer = document.createElement('div');
        toastContainer.id = 'toast-container';
        toastContainer.className = 'toast toast-top toast-end';
        document.body.appendChild(toastContainer);
    }
    
    // Create toast
    const toast = document.createElement('div');
    toast.className = `alert alert-${type}`;
    toast.innerHTML = `
        <span>${message}</span>
    `;
    
    // Add to container
    toastContainer.appendChild(toast);
    
    // Remove after 3 seconds
    setTimeout(() => {
        toast.remove();
    }, 3000);
}

// Function to confirm actions
function confirmAction(message, callback) {
    if (confirm(message)) {
        callback();
    }
}

// Function to toggle sidebar on mobile
function toggleSidebar() {
    const sidebar = document.querySelector('.sidebar');
    sidebar.classList.toggle('hidden');
}

// Add event listeners for mobile menu toggle
document.addEventListener('DOMContentLoaded', function() {
    const menuToggle = document.getElementById('menu-toggle');
    if (menuToggle) {
        menuToggle.addEventListener('click', toggleSidebar);
    }
});

// HTMX event listeners
document.body.addEventListener('htmx:afterSwap', function(evt) {
    // Reinitialize any components that need it after HTMX swaps content
    console.log('HTMX content swapped');
});

document.body.addEventListener('htmx:beforeRequest', function(evt) {
    // Show loading indicator
    console.log('HTMX request started');
});

document.body.addEventListener('htmx:afterRequest', function(evt) {
    // Hide loading indicator
    console.log('HTMX request completed');
});