class APIClient {
    constructor() {
        this.baseURL = 'http://localhost:8082/api/v1';
        console.log('APIClient initialized - Version 2.0'); // Debug version check
    }

    async request(endpoint, options = {}) {
        const url = `${this.baseURL}${endpoint}`;
        const config = {
            headers: {
                'Content-Type': 'application/json',
                ...options.headers
            },
            ...options
        };

        try {
            const response = await fetch(url, config);
            const data = await response.json();
            
            if (!response.ok) {
                throw new Error(data.error || 'Request failed');
            }
            
            return data;
        } catch (error) {
            console.error('API request failed:', error);
            throw error;
        }
    }

    // Loan Application APIs
    async createLoanApplication(applicationData) {
        return this.request('/loans', {
            method: 'POST',
            body: JSON.stringify(applicationData)
        });
    }

    async getLoanApplications() {
        return this.request('/loans');
    }

    async getLoanApplication(id) {
        return this.request(`/loans/${id}`);
    }

    async getWorkflowStatus(id) {
        return this.request(`/loans/${id}/status`);
    }

    // Document APIs
    async uploadDocument(loanId, documentData) {
        return this.request(`/loans/${loanId}/documents`, {
            method: 'POST',
            body: JSON.stringify(documentData)
        });
    }

    async verifyDocument(loanId, verificationData) {
        return this.request(`/loans/${loanId}/verify-documents`, {
            method: 'POST',
            body: JSON.stringify(verificationData)
        });
    }

    // Appraisal APIs
    async completeAppraisal(loanId, appraisalData) {
        return this.request(`/loans/${loanId}/appraisal`, {
            method: 'POST',
            body: JSON.stringify(appraisalData)
        });
    }

    // Underwriting APIs
    async makeUnderwritingDecision(loanId, decisionData) {
        return this.request(`/loans/${loanId}/underwriting`, {
            method: 'POST',
            body: JSON.stringify(decisionData)
        });
    }

    // Funding APIs
    async processFunding(loanId, fundingData) {
        return this.request(`/loans/${loanId}/funding`, {
            method: 'POST',
            body: JSON.stringify(fundingData)
        });
    }
}

// Global API client instance
const api = new APIClient();
