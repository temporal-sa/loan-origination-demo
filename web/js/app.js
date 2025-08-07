// Main application initialization
document.addEventListener('DOMContentLoaded', function() {
    // Initialize persona manager
    personaManager.init();
    
    // Add some demo data on first load
    initializeDemoData();
});

async function initializeDemoData() {
    try {
        // Check if we already have data
        const loans = await api.getLoanApplications();
        if (loans == null || loans.length === 0) {
            // Create a sample loan application for demo purposes            
            document.getElementById('borrowerName').value = "Brendan Myers";
            document.getElementById('borrowerEmail').value = "brendan.myers@temporal.io";
            document.getElementById('borrowerPhone').value = "+61 400 123 456";
            document.getElementById('loanAmount').value = "1000000";
            document.getElementById('loanPurpose').selectedIndex = 1;

            // await api.createLoanApplication(sampleLoan);
            console.log('Demo loan application created');
        }
    } catch (error) {
        console.error('Error initializing demo data:', error);
    }
}

// Utility functions
function formatCurrency(amount) {
    return new Intl.NumberFormat('en-US', {
        style: 'currency',
        currency: 'USD'
    }).format(amount);
}

function formatDate(dateString) {
    return new Date(dateString).toLocaleDateString('en-US', {
        year: 'numeric',
        month: 'short',
        day: 'numeric'
    });
}

// Handle page visibility changes to pause/resume auto-refresh
document.addEventListener('visibilitychange', function() {
    if (document.hidden) {
        personaManager.stopAutoRefresh();
    } else {
        personaManager.startAutoRefresh();
    }
});
