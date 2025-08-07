package handlers

import (
	"net/http"
	"strings"
	"time"

	"loan-origination-system/internal/workflows"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"go.temporal.io/api/workflowservice/v1"
	"go.temporal.io/sdk/client"
)

type LoanHandler struct {
	temporalClient client.Client
	activeLoans    map[string]workflows.LoanApplication // Simple in-memory registry
}

func NewLoanHandler(temporalClient client.Client) *LoanHandler {
	return &LoanHandler{
		temporalClient: temporalClient,
		activeLoans:    make(map[string]workflows.LoanApplication),
	}
}

// CreateLoanApplication creates a new loan application and starts the workflow
func (h *LoanHandler) CreateLoanApplication(c *gin.Context) {
	var req struct {
		BorrowerName  string  `json:"borrower_name" binding:"required"`
		BorrowerEmail string  `json:"borrower_email" binding:"required"`
		BorrowerPhone string  `json:"borrower_phone" binding:"required"`
		LoanAmount    float64 `json:"loan_amount" binding:"required"`
		LoanPurpose   string  `json:"loan_purpose" binding:"required"`
		CreatedBy     string  `json:"created_by" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Create loan application data structure
	loanID := uuid.New().String()
	now := time.Now()

	loanApp := workflows.LoanApplication{
		ID:            loanID,
		BorrowerName:  req.BorrowerName,
		BorrowerEmail: req.BorrowerEmail,
		BorrowerPhone: req.BorrowerPhone,
		LoanAmount:    req.LoanAmount,
		LoanPurpose:   req.LoanPurpose,
		Status:        "pending",
		CreatedBy:     req.CreatedBy,
		CreatedAt:     now,
		UpdatedAt:     now,
		WorkflowID:    "loan-origination-" + loanID,
	}

	// Start Temporal workflow
	workflowOptions := client.StartWorkflowOptions{
		ID:        loanApp.WorkflowID,
		TaskQueue: "loan-origination-task-queue",
	}

	_, err := h.temporalClient.ExecuteWorkflow(
		c.Request.Context(),
		workflowOptions,
		workflows.LoanOriginationWorkflow,
		workflows.LoanOriginationWorkflowInput{
			LoanApplication: loanApp,
		},
	)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to start workflow"})
		return
	}

	// Store in active loans registry
	h.activeLoans[loanID] = loanApp

	c.JSON(http.StatusCreated, loanApp)
}

// GetLoanApplications returns all loan applications by querying active workflows
func (h *LoanHandler) GetLoanApplications(c *gin.Context) {
	h.activeLoans = make(map[string]workflows.LoanApplication)

	res, err := h.temporalClient.ListWorkflow(c.Request.Context(), &workflowservice.ListWorkflowExecutionsRequest{})
	if err == nil {
		for _, wf := range res.Executions {
			loanId, _ := strings.CutPrefix(wf.Execution.WorkflowId, "loan-origination-")
			h.activeLoans[loanId] = workflows.LoanApplication{}
		}
	}

	var loanResponses []map[string]interface{}

	// Query each active workflow for its current state
	for loanID, _ := range h.activeLoans {
		workflowID := "loan-origination-" + loanID

		// Query the workflow for current state
		resp, _ := h.temporalClient.QueryWorkflow(
			c.Request.Context(),
			workflowID,
			"",
			"getLoanApplication",
		)

		var loanData workflows.LoanOriginationState
		if err := resp.Get(&loanData); err == nil {
			// Flatten the response to match frontend expectations
			flatLoan := map[string]interface{}{
				"id":                    loanData.LoanApplication.ID,
				"borrower_name":         loanData.LoanApplication.BorrowerName,
				"borrower_email":        loanData.LoanApplication.BorrowerEmail,
				"borrower_phone":        loanData.LoanApplication.BorrowerPhone,
				"loan_amount":           loanData.LoanApplication.LoanAmount,
				"loan_purpose":          loanData.LoanApplication.LoanPurpose,
				"status":                loanData.LoanApplication.Status,
				"next_step":             loanData.NextStep,
				"created_by":            loanData.LoanApplication.CreatedBy,
				"created_at":            loanData.LoanApplication.CreatedAt,
				"updated_at":            loanData.LoanApplication.UpdatedAt,
				"workflow_id":           loanData.LoanApplication.WorkflowID,
				"documents":             loanData.Documents,
				"appraisal":             loanData.Appraisal,
				"underwriting_decision": loanData.UnderwritingDecision,
			}
			loanResponses = append(loanResponses, flatLoan)
		}
	}

	c.JSON(http.StatusOK, loanResponses)
}

// GetLoanApplication returns a specific loan application by querying the workflow
func (h *LoanHandler) GetLoanApplication(c *gin.Context) {
	loanID := c.Param("id")
	workflowID := "loan-origination-" + loanID

	// Query the workflow for current state
	resp, err := h.temporalClient.QueryWorkflow(
		c.Request.Context(),
		workflowID,
		"",
		"getLoanApplication",
	)

	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Loan application not found"})
		return
	}

	var loanData workflows.LoanOriginationState
	if err := resp.Get(&loanData); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to parse loan data"})
		return
	}

	c.JSON(http.StatusOK, loanData)
}

// UploadDocument handles document upload and sends signal to workflow
func (h *LoanHandler) UploadDocument(c *gin.Context) {
	loanID := c.Param("id")

	var req struct {
		DocumentType string `json:"document_type" binding:"required"`
		FileName     string `json:"file_name" binding:"required"`
		FilePath     string `json:"file_path" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Generate document ID
	documentID := uuid.New().String()

	// Send signal to workflow (workflow will store the document data)
	workflowID := "loan-origination-" + loanID
	err := h.temporalClient.SignalWorkflow(
		c.Request.Context(),
		workflowID,
		"",
		"document-uploaded",
		workflows.DocumentUploadedSignal{
			DocumentID:   documentID,
			DocumentType: req.DocumentType,
		},
	)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to send workflow signal"})
		return
	}

	// Return document info
	document := workflows.Document{
		ID:                 documentID,
		DocumentType:       req.DocumentType,
		FileName:           req.FileName,
		FilePath:           req.FilePath,
		VerificationStatus: "pending",
		UploadedAt:         time.Now(),
	}

	c.JSON(http.StatusCreated, document)
}

// VerifyDocument handles third-party document verification
func (h *LoanHandler) VerifyDocument(c *gin.Context) {
	loanID := c.Param("id")

	var req struct {
		DocumentID          string                 `json:"document_id" binding:"required"`
		VerificationStatus  string                 `json:"verification_status" binding:"required"`
		VerificationDetails map[string]interface{} `json:"verification_details"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Send signal to workflow (workflow will update document verification status)
	workflowID := "loan-origination-" + loanID
	err := h.temporalClient.SignalWorkflow(
		c.Request.Context(),
		workflowID,
		"",
		"document-verified",
		workflows.DocumentVerificationSignal{
			DocumentID:          req.DocumentID,
			VerificationStatus:  req.VerificationStatus,
			VerificationDetails: req.VerificationDetails,
		},
	)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to send workflow signal"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Document verification processed"})
}

// CompleteAppraisal handles appraisal completion
func (h *LoanHandler) CompleteAppraisal(c *gin.Context) {
	loanID := c.Param("id")

	var req struct {
		PropertyValue  float64 `json:"property_value" binding:"required"`
		AppraisalNotes string  `json:"appraisal_notes"`
		AppraiserID    string  `json:"appraiser_id" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Send signal to workflow (workflow will store the appraisal data)
	workflowID := "loan-origination-" + loanID
	err := h.temporalClient.SignalWorkflow(
		c.Request.Context(),
		workflowID,
		"",
		"appraisal-completed",
		workflows.AppraisalCompletedSignal{
			PropertyValue:  req.PropertyValue,
			AppraisalNotes: req.AppraisalNotes,
			AppraiserID:    req.AppraiserID,
		},
	)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to send workflow signal"})
		return
	}

	// Return appraisal info
	now := time.Now()
	appraisal := workflows.Appraisal{
		ID:             "appraisal-" + loanID,
		PropertyValue:  req.PropertyValue,
		AppraisalNotes: req.AppraisalNotes,
		AppraiserID:    req.AppraiserID,
		Status:         "completed",
		CompletedAt:    &now,
		CreatedAt:      now,
	}

	c.JSON(http.StatusOK, appraisal)
}

// MakeUnderwritingDecision handles underwriting decisions
func (h *LoanHandler) MakeUnderwritingDecision(c *gin.Context) {
	loanID := c.Param("id")

	var req struct {
		Decision      string `json:"decision" binding:"required"`
		Comments      string `json:"comments"`
		UnderwriterID string `json:"underwriter_id" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Send signal to workflow (workflow will store the underwriting decision)
	workflowID := "loan-origination-" + loanID
	err := h.temporalClient.SignalWorkflow(
		c.Request.Context(),
		workflowID,
		"",
		"underwriting-decision",
		workflows.UnderwritingDecisionSignal{
			Decision:      req.Decision,
			Comments:      req.Comments,
			UnderwriterID: req.UnderwriterID,
		},
	)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to send workflow signal"})
		return
	}

	// Return decision info
	decision := workflows.UnderwritingDecision{
		ID:            "decision-" + loanID,
		Decision:      req.Decision,
		Comments:      req.Comments,
		UnderwriterID: req.UnderwriterID,
		DecisionDate:  time.Now(),
	}

	c.JSON(http.StatusOK, decision)
}

// ProcessFunding handles funding completion
func (h *LoanHandler) ProcessFunding(c *gin.Context) {
	loanID := c.Param("id")

	var req struct {
		FundManagerID string  `json:"fund_manager_id" binding:"required"`
		FundingAmount float64 `json:"funding_amount" binding:"required"`
		FundingNotes  string  `json:"funding_notes"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Send signal to workflow (workflow will update status to funded)
	workflowID := "loan-origination-" + loanID
	err := h.temporalClient.SignalWorkflow(
		c.Request.Context(),
		workflowID,
		"",
		"funding-completed",
		workflows.FundingCompletedSignal{
			FundManagerID: req.FundManagerID,
			FundingAmount: req.FundingAmount,
			FundingNotes:  req.FundingNotes,
		},
	)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to send workflow signal"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Funding processed successfully"})
}

// GetWorkflowStatus returns the current workflow status
func (h *LoanHandler) GetWorkflowStatus(c *gin.Context) {
	loanID := c.Param("id")
	workflowID := "loan-origination-" + loanID

	resp, err := h.temporalClient.DescribeWorkflowExecution(
		c.Request.Context(),
		workflowID,
		"",
	)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get workflow status"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"workflow_id": workflowID,
		"status":      resp.WorkflowExecutionInfo.Status.String(),
		"start_time":  resp.WorkflowExecutionInfo.StartTime,
	})
}
