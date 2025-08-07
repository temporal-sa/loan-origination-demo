class PersonaManager {
    constructor() {
        this.currentRole = 'loan-officer';
        this.loans = [];
        this.refreshInterval = null;
        console.log('PersonaManager initialized - Version 2.0'); // Debug version check
    }

    init() {
        this.setupRoleSelector();
        this.setupEventListeners();
        this.switchRole(this.currentRole);
        this.startAutoRefresh();
    }

    setupRoleSelector() {
        const roleSelect = document.getElementById('roleSelect');
        roleSelect.addEventListener('change', (e) => {
            this.switchRole(e.target.value);
        });
    }

    setupEventListeners() {
        // Loan Officer form
        const loanForm = document.getElementById('loan-application-form');
        if (loanForm) {
            loanForm.addEventListener('submit', (e) => this.handleLoanApplicationSubmit(e));
        }

        // Modal close
        const modal = document.getElementById('modal');
        const closeBtn = document.querySelector('.close');
        if (closeBtn) {
            closeBtn.addEventListener('click', () => {
                modal.style.display = 'none';
            });
        }

        // Close modal when clicking outside
        window.addEventListener('click', (e) => {
            if (e.target === modal) {
                modal.style.display = 'none';
            }
        });
    }

    switchRole(role) {
        this.currentRole = role;
        
        // Hide all role views
        document.querySelectorAll('.role-view').forEach(view => {
            view.style.display = 'none';
        });
        
        // Show current role view
        const currentView = document.getElementById(`${role}-view`);
        if (currentView) {
            currentView.style.display = 'block';
        }
        
        // Load data for current role
        this.loadRoleData();
    }

    async loadRoleData() {
        try {
            this.loans = await api.getLoanApplications();
            // Ensure loans is always an array
            if (!Array.isArray(this.loans)) {
                this.loans = [];
            }
            this.renderRoleView();
        } catch (error) {
            console.error('Error loading data:', error);
            this.loans = []; // Fallback to empty array
            this.showMessage('Error loading data: ' + error.message, 'error');
            this.renderRoleView(); // Still render the view with empty data
        }
    }

    renderRoleView() {
        switch (this.currentRole) {
            case 'loan-officer':
                this.renderLoanOfficerView();
                break;
            case 'customer':
                this.renderCustomerView();
                break;
            case 'loan-processor':
                this.renderLoanProcessorView();
                break;
            case 'appraiser':
                this.renderAppraiserView();
                break;
            case 'underwriter':
                this.renderUnderwriterView();
                break;
            case 'fund-manager':
                this.renderFundManagerView();
                break;
        }
    }

    renderLoanOfficerView() {
        const container = document.getElementById('officer-applications');
        const myLoans = this.loans.filter(loan => loan.created_by === 'loan-officer');
        
        container.innerHTML = myLoans.length === 0 ? 
            '<p>No applications created yet.</p>' : 
            myLoans.map(loan => this.createLoanCard(loan, ['view-details'])).join('');
    }

    renderCustomerView() {
        const container = document.getElementById('customer-loans');
        const processingLoans = this.loans.filter(loan => 
            loan.status === 'processing' || loan.status === 'pending'
        );
        
        container.innerHTML = processingLoans.length === 0 ? 
            '<p>No loans requiring document upload.</p>' : 
            processingLoans.map(loan => this.createLoanCard(loan, ['upload-documents'])).join('');
    }

    renderLoanProcessorView() {
        const container = document.getElementById('processor-applications');
        const pendingLoans = this.loans.filter(loan => 
            loan.status === 'processing' && 
            loan.documents && 
            loan.documents.some(doc => doc.verification_status === 'pending')
        );
        
        container.innerHTML = pendingLoans.length === 0 ? 
            '<p>No applications pending document verification.</p>' : 
            pendingLoans.map(loan => this.createLoanCard(loan, ['verify-documents'])).join('');
    }

    renderAppraiserView() {
        const container = document.getElementById('appraiser-applications');
        const appraisalLoans = this.loans.filter(loan => 
            loan.status === 'processing' && 
            (!loan.appraisal || loan.appraisal.status !== 'completed')
        );
        
        container.innerHTML = appraisalLoans.length === 0 ? 
            '<p>No applications requiring appraisal.</p>' : 
            appraisalLoans.map(loan => this.createLoanCard(loan, ['complete-appraisal'])).join('');
    }

    renderUnderwriterView() {
        const container = document.getElementById('underwriter-applications');
        const underwritingLoans = this.loans.filter(loan => 
            loan.status === 'processing' && 
            loan.appraisal && 
            loan.appraisal.status === 'completed' &&
            (!loan.underwriting_decision || loan.underwriting_decision.decision == 'needs_more_info')
        );
        
        container.innerHTML = underwritingLoans.length === 0 ? 
            '<p>No applications pending underwriting.</p>' : 
            underwritingLoans.map(loan => this.createLoanCard(loan, ['make-decision'])).join('');
    }

    renderFundManagerView() {
        const container = document.getElementById('fund-manager-applications');
        const approvedLoans = this.loans.filter(loan => 
            loan.underwriting_decision && 
            loan.underwriting_decision.decision === 'approved' &&
            (loan.status === 'approved' || loan.status === 'processing') &&
            loan.status !== 'funded'
        );
        
        console.log('Fund Manager - All loans:', this.loans);
        console.log('Fund Manager - Approved loans:', approvedLoans);
        
        container.innerHTML = approvedLoans.length === 0 ? 
            '<p>No approved applications for funding.</p>' : 
            approvedLoans.map(loan => this.createLoanCard(loan, ['process-funding'])).join('');
    }

    createLoanCard(loan, actions = []) {
        const documentsInfo = loan.documents ? 
            `<div class="info-item">
                <span class="info-label">Documents</span>
                <span class="info-value">${loan.documents.length} uploaded</span>
            </div>` : '';

        const appraisalInfo = loan.appraisal ? 
            `<div class="info-item">
                <span class="info-label">Property Value</span>
                <span class="info-value">$${loan.appraisal.property_value?.toLocaleString() || 'N/A'}</span>
            </div>` : '';

        const underwritingInfo = loan.underwriting_decision ? 
            `<div class="info-item">
                <span class="info-label">Decision</span>
                <span class="info-value">${loan.underwriting_decision.decision}</span>
            </div>` : '';

        const actionButtons = actions.map(action => {
            switch (action) {
                case 'view-details':
                    return `<button onclick="personaManager.viewLoanDetails('${loan.id}')">View Details</button>`;
                case 'upload-documents':
                    return `<button onclick="personaManager.showDocumentUpload('${loan.id}')">Upload Documents</button>`;
                case 'verify-documents':
                    return `<button onclick="personaManager.showDocumentVerification('${loan.id}')">Verify Documents</button>`;
                case 'complete-appraisal':
                    return `<button onclick="personaManager.showAppraisalForm('${loan.id}')">Complete Appraisal</button>`;
                case 'make-decision':
                    return `<button onclick="personaManager.showUnderwritingForm('${loan.id}')">Make Decision</button>`;
                case 'process-funding':
                    return `<button onclick="personaManager.processFunding('${loan.id}')">Process Funding</button>`;
                default:
                    return '';
            }
        }).join('');

        return `
            <div class="application-card">
                <h4>${loan.borrower_name} - $${loan.loan_amount?.toLocaleString()}</h4>
                <div class="application-info">
                    <div class="info-item">
                        <span class="info-label">Status</span>
                        <span class="info-value"><span class="status ${loan.status}">${loan.status}</span></span>
                    </div>
                    <div class="info-item">
                        <span class="info-label">Purpose</span>
                        <span class="info-value">${loan.loan_purpose}</span>
                    </div>
                    <div class="info-item">
                        <span class="info-label">Email</span>
                        <span class="info-value">${loan.borrower_email}</span>
                    </div>
                    ${documentsInfo}
                    ${appraisalInfo}
                    ${underwritingInfo}
                    <div class="info-item">
                        <span class="info-label">Next Step</span>
                        <span class="info-value">${loan.next_step}</span>
                    </div>
                    
                </div>
                <div class="actions">
                    ${actionButtons}
                </div>
            </div>
        `;
    }

    async handleLoanApplicationSubmit(e) {
        e.preventDefault();
        
        const formData = {
            borrower_name: document.getElementById('borrowerName').value,
            borrower_email: document.getElementById('borrowerEmail').value,
            borrower_phone: document.getElementById('borrowerPhone').value,
            loan_amount: parseFloat(document.getElementById('loanAmount').value),
            loan_purpose: document.getElementById('loanPurpose').value,
            created_by: 'loan-officer'
        };

        try {
            await api.createLoanApplication(formData);
            this.showMessage('Loan application created successfully!', 'success');
            e.target.reset();
            this.loadRoleData();
        } catch (error) {
            this.showMessage('Error creating loan application: ' + error.message, 'error');
        }
    }

    showDocumentUpload(loanId) {
        const modalBody = document.getElementById('modal-body');
        modalBody.innerHTML = `
            <h3>Upload Documents</h3>
            <form id="document-upload-form">
                <div class="form-group">
                    <label for="documentType">Document Type:</label>
                    <select id="documentType" required>
                        <option value="">Select Type</option>
                        <option value="income_statement">Income Statement</option>
                        <option value="bank_statement">Bank Statement</option>
                        <option value="id_proof">ID Proof</option>
                        <option value="employment_verification">Employment Verification</option>
                        <option value="tax_returns">Tax Returns</option>
                    </select>
                </div>
                <div class="form-group">
                    <label for="fileName">File Name:</label>
                    <input type="text" id="fileName" required placeholder="e.g., income_statement_2024.pdf">
                </div>
                <div class="form-group">
                    <label for="filePath">File Path:</label>
                    <input type="text" id="filePath" required placeholder="e.g., /uploads/documents/...">
                </div>
                <button type="submit">Upload Document</button>
            </form>
        `;

        document.getElementById('document-upload-form').addEventListener('submit', async (e) => {
            e.preventDefault();
            
            const documentData = {
                document_type: document.getElementById('documentType').value,
                file_name: document.getElementById('fileName').value,
                file_path: document.getElementById('filePath').value
            };

            try {
                await api.uploadDocument(loanId, documentData);
                this.showMessage('Document uploaded successfully!', 'success');
                document.getElementById('modal').style.display = 'none';
                this.loadRoleData();
            } catch (error) {
                this.showMessage('Error uploading document: ' + error.message, 'error');
            }
        });

        document.getElementById('modal').style.display = 'block';
    }

    showDocumentVerification(loanId) {
        const loan = this.loans.find(l => l.id === loanId);
        const pendingDocs = loan.documents.filter(doc => doc.verification_status === 'pending');

        const modalBody = document.getElementById('modal-body');
        modalBody.innerHTML = `
            <h3>Verify Documents</h3>
            <div class="documents-list">
                ${pendingDocs.map(doc => `
                    <div class="document-item">
                        <div class="document-info">
                            <strong>${doc.document_type}</strong><br>
                            <small>${doc.file_name}</small>
                        </div>
                        <div class="actions">
                            <button class="success" onclick="personaManager.verifyDocument('${loanId}', '${doc.id}', 'verified')">Verify</button>
                            <button class="danger" onclick="personaManager.verifyDocument('${loanId}', '${doc.id}', 'rejected')">Reject</button>
                        </div>
                    </div>
                `).join('')}
            </div>
        `;

        document.getElementById('modal').style.display = 'block';
    }

    async verifyDocument(loanId, documentId, status) {
        try {
            await api.verifyDocument(loanId, {
                document_id: documentId,
                verification_status: status,
                verification_details: { verified_by: 'loan-processor', timestamp: new Date().toISOString() }
            });
            
            this.showMessage(`Document ${status} successfully!`, 'success');
            document.getElementById('modal').style.display = 'none';
            this.loadRoleData();
        } catch (error) {
            this.showMessage('Error verifying document: ' + error.message, 'error');
        }
    }

    showAppraisalForm(loanId) {
        const modalBody = document.getElementById('modal-body');
        modalBody.innerHTML = `
            <h3>Complete Appraisal</h3>
            <form id="appraisal-form">
                <div class="form-group">
                    <label for="propertyValue">Property Value ($):</label>
                    <input type="number" id="propertyValue" step="0.01" required>
                </div>
                <div class="form-group">
                    <label for="appraisalNotes">Appraisal Notes:</label>
                    <textarea id="appraisalNotes" rows="4" placeholder="Enter appraisal notes and observations..."></textarea>
                </div>
                <button type="submit">Complete Appraisal</button>
            </form>
        `;

        document.getElementById('appraisal-form').addEventListener('submit', async (e) => {
            e.preventDefault();
            
            const appraisalData = {
                property_value: parseFloat(document.getElementById('propertyValue').value),
                appraisal_notes: document.getElementById('appraisalNotes').value,
                appraiser_id: 'appraiser-001'
            };

            try {
                await api.completeAppraisal(loanId, appraisalData);
                this.showMessage('Appraisal completed successfully!', 'success');
                document.getElementById('modal').style.display = 'none';
                this.loadRoleData();
            } catch (error) {
                this.showMessage('Error completing appraisal: ' + error.message, 'error');
            }
        });

        document.getElementById('modal').style.display = 'block';
    }

    showUnderwritingForm(loanId) {
        const modalBody = document.getElementById('modal-body');
        modalBody.innerHTML = `
            <h3>Underwriting Decision</h3>
            <form id="underwriting-form">
                <div class="form-group">
                    <label for="decision">Decision:</label>
                    <select id="decision" required>
                        <option value="">Select Decision</option>
                        <option value="approved">Approved</option>
                        <option value="rejected">Rejected</option>
                        <option value="needs_more_info">Needs More Information</option>
                    </select>
                </div>
                <div class="form-group">
                    <label for="comments">Comments:</label>
                    <textarea id="comments" rows="4" placeholder="Enter underwriting comments..."></textarea>
                </div>
                <button type="submit">Submit Decision</button>
            </form>
        `;

        document.getElementById('underwriting-form').addEventListener('submit', async (e) => {
            e.preventDefault();
            
            const decisionData = {
                decision: document.getElementById('decision').value,
                comments: document.getElementById('comments').value,
                underwriter_id: 'underwriter-001'
            };

            try {
                await api.makeUnderwritingDecision(loanId, decisionData);
                this.showMessage('Underwriting decision submitted successfully!', 'success');
                document.getElementById('modal').style.display = 'none';
                this.loadRoleData();
            } catch (error) {
                this.showMessage('Error submitting decision: ' + error.message, 'error');
            }
        });

        document.getElementById('modal').style.display = 'block';
    }

    async processFunding(loanId) {
        const loan = this.loans.find(l => l.id === loanId);
        
        const modalBody = document.getElementById('modal-body');
        modalBody.innerHTML = `
            <h3>Process Funding</h3>
            <div class="loan-summary">
                <p><strong>Borrower:</strong> ${loan.borrower_name}</p>
                <p><strong>Loan Amount:</strong> $${loan.loan_amount?.toLocaleString()}</p>
            </div>
            <form id="funding-form">
                <div class="form-group">
                    <label for="fundingAmount">Funding Amount ($):</label>
                    <input type="number" id="fundingAmount" step="0.01" value="${loan.loan_amount}" required>
                </div>
                <div class="form-group">
                    <label for="fundingNotes">Funding Notes:</label>
                    <textarea id="fundingNotes" rows="3" placeholder="Enter funding notes and instructions..."></textarea>
                </div>
                <button type="submit">Process Funding</button>
            </form>
        `;

        document.getElementById('funding-form').addEventListener('submit', async (e) => {
            e.preventDefault();
            
            const fundingData = {
                fund_manager_id: 'fund-manager-001',
                funding_amount: parseFloat(document.getElementById('fundingAmount').value),
                funding_notes: document.getElementById('fundingNotes').value
            };

            try {
                await api.processFunding(loanId, fundingData);
                this.showMessage('Funding processed successfully!', 'success');
                document.getElementById('modal').style.display = 'none';
                this.loadRoleData();
            } catch (error) {
                this.showMessage('Error processing funding: ' + error.message, 'error');
            }
        });

        document.getElementById('modal').style.display = 'block';
    }

    viewLoanDetails(loanId) {
        const loan = this.loans.find(l => l.id === loanId);
        const modalBody = document.getElementById('modal-body');
        
        modalBody.innerHTML = `
            <h3>Loan Application Details</h3>
            <div class="loan-details">
                <div class="detail-section">
                    <h4>Borrower Information</h4>
                    <p><strong>Name:</strong> ${loan.borrower_name}</p>
                    <p><strong>Email:</strong> ${loan.borrower_email}</p>
                    <p><strong>Phone:</strong> ${loan.borrower_phone}</p>
                </div>
                
                <div class="detail-section">
                    <h4>Loan Information</h4>
                    <p><strong>Amount:</strong> $${loan.loan_amount?.toLocaleString()}</p>
                    <p><strong>Purpose:</strong> ${loan.loan_purpose}</p>
                    <p><strong>Status:</strong> <span class="status ${loan.status}">${loan.status}</span></p>
                    <p><strong>Next step:</strong> ${loan.next_step}</p>
                </div>
                
                ${loan.documents && loan.documents.length > 0 ? `
                <div class="detail-section">
                    <h4>Documents</h4>
                    ${loan.documents.map(doc => `
                        <p><strong>${doc.document_type}:</strong> ${doc.file_name} 
                        <span class="status ${doc.verification_status}">${doc.verification_status}</span></p>
                    `).join('')}
                </div>
                ` : ''}
                
                ${loan.appraisal ? `
                <div class="detail-section">
                    <h4>Appraisal</h4>
                    <p><strong>Property Value:</strong> $${loan.appraisal.property_value?.toLocaleString()}</p>
                    <p><strong>Notes:</strong> ${loan.appraisal.appraisal_notes || 'N/A'}</p>
                </div>
                ` : ''}
                
                ${loan.underwriting_decision ? `
                <div class="detail-section">
                    <h4>Underwriting Decision</h4>
                    <p><strong>Decision:</strong> <span class="status ${loan.underwriting_decision.decision}">${loan.underwriting_decision.decision}</span></p>
                    <p><strong>Comments:</strong> ${loan.underwriting_decision.comments || 'N/A'}</p>
                </div>
                ` : ''}
            </div>
        `;

        document.getElementById('modal').style.display = 'block';
    }

    showMessage(message, type = 'info') {
        // Remove existing messages
        const existingMessages = document.querySelectorAll('.message');
        existingMessages.forEach(msg => msg.remove());

        // Create new message
        const messageDiv = document.createElement('div');
        messageDiv.className = `message ${type}`;
        messageDiv.textContent = message;

        // Insert at the top of the current role view
        const currentView = document.querySelector('.role-view[style*="block"]');
        if (currentView) {
            currentView.insertBefore(messageDiv, currentView.firstChild);
        }

        // Auto-remove after 5 seconds
        setTimeout(() => {
            messageDiv.remove();
        }, 5000);
    }

    startAutoRefresh() {
        // Refresh data every 30 seconds
        this.refreshInterval = setInterval(() => {
            this.loadRoleData();
        }, 30000);
    }

    stopAutoRefresh() {
        if (this.refreshInterval) {
            clearInterval(this.refreshInterval);
        }
    }
}

// Global persona manager instance
const personaManager = new PersonaManager();
